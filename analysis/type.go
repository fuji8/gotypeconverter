package analysis

import (
	"fmt"
	"go/types"
	"strings"
)

const M = 100

func InitType(typ types.Type, name string) Type {
	return Type{
		typ:  typ,
		name: name,
	}
}

type TType[T types.Type] struct {
	typ  T
	name string
}
type Type struct {
	typ  types.Type
	name string
}

type TypeStruct struct {
	typ  *types.Struct
	name string
}

type TypeSlice struct {
	typ  *types.Slice
	name string
}

type TypePointer struct {
	typ  *types.Pointer
	name string
}

type TypeBasic struct {
	typ  *types.Basic
	name string
}

type TypeNamed struct {
	typ  *types.Named
	name string
}

// getFields embed を含めて、フィールドを取得
func getFields(t *types.Struct) ([]*types.Var, []string) {
	fields := make([]*types.Var, 0)
	tags := make([]string, 0)
	for i := 0; i < t.NumFields(); i++ {
		if t.Field(i).Embedded() {
			if tE, ok := t.Field(i).Type().Underlying().(*types.Struct); ok {
				f, t := getFields(tE)
				fields = append(fields, f...)
				tags = append(tags, t...)
				continue
			}
		}
		fields = append(fields, t.Field(i))
		tags = append(tags, t.Tag(i))
	}

	return fields, tags
}

func (fm *FuncMaker) structAndOther(dstT TypeStruct, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	var written float64
	var conv string
	fields, tags := getFields(dstT.typ)
	for i, f := range fields {
		if !fm.varVisiable(f) {
			continue
		}

		// if struct tag "cvt" exists, use struct tag
		_, _, _, dOption := getTag(tags[i])
		if dOption == Ignore || dOption == ReadOnly {
			continue
		}

		var score float64
		score, conv = fm.makeFunc(Type{typ: f.Type()}, src,
			selectorGen(dstSelector, f),
			srcSelector,
			index,
			history,
		)

		written = 1 / float64(len(fields)) * score
		if written > 0 {
			break
		}
	}
	return written, conv
}

func (fm *FuncMaker) otherAndStruct(dst Type, srcT TypeStruct, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	var written float64
	var conv string
	fields, tags := getFields(srcT.typ)
	for j, f := range fields {
		if !fm.varVisiable(f) {
			continue
		}
		// if struct tag "cvt" exists, use struct tag
		_, _, _, sOption := getTag(tags[j])
		if sOption == Ignore || sOption == WriteOnly {
			continue
		}

		var score float64
		score, conv = fm.makeFunc(dst, Type{typ: f.Type()},
			dstSelector,
			selectorGen(srcSelector, f),
			index,
			history,
		)

		written = 1 / float64(len(fields)) * score
		if written > 0 {
			break
		}
	}
	return written, conv
}

func (fm *FuncMaker) structAndStruct(dstT, srcT TypeStruct, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	var written float64 = 0
	convs := make([]string, 0)
	// field 同士の比較
	dFields, dtags := getFields(dstT.typ)
	sFields, stags := getFields(srcT.typ)
	for i, df := range dFields {
		if !fm.varVisiable(df) {
			continue
		}
		// if struct tag "cvt" exists, use struct tag
		dField, _, dWriteField, dOption := getTag(dtags[i])
		if dWriteField != "" {
			dField = dWriteField
		}
		if dField == "" {
			dField = df.Name()
		}
		if dOption == Ignore || dOption == ReadOnly {
			continue
		}
		for j, sf := range sFields {
			if !fm.varVisiable(sf) {
				continue
			}
			// if struct tag "cvt" exists, use struct tag
			sField, sReadField, _, sOption := getTag(stags[j])
			if sReadField != "" {
				sField = sReadField
			}
			if sField == "" {
				sField = sf.Name()
			}
			if sOption == Ignore || sOption == WriteOnly {
				continue
			}

			if dField == sField {
				score, conv := fm.makeFunc(Type{typ: df.Type()}, Type{typ: sf.Type()},
					selectorGen(dstSelector, df),
					selectorGen(srcSelector, sf),
					index,
					history,
				)
				if score == 0 {
					continue
				}
				written += 1 / float64(len(dFields)) * score
				convs = append(convs, conv)
			}
		}
	}

	return written, strings.Join(convs, "\n")
}

func (fm *FuncMaker) sliceAndOther(dstT TypeSlice, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	score, conv := fm.makeFunc(Type{typ: dstT.typ.Elem()}, src, dstSelector+"[0]", srcSelector, index, history)
	if score == 0 {
		return 0, ""
	}
	dt, err := fm.formatPkgType(dstT.typ)
	if err != nil {
		return 0, ""
	}
	convs := []string{
		fmt.Sprintf("%s = make(%s, 1)", dstSelector, dt),
		conv,
	}
	return score / M, strings.Join(convs, "\n")
}

func (fm *FuncMaker) otherAndSlice(dst Type, srcT TypeSlice, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	score, conv := fm.makeFunc(dst, Type{typ: srcT.typ.Elem()}, dstSelector, srcSelector+"[0]", index, history)
	if score == 0 {
		return 0, ""
	}
	convs := []string{
		fmt.Sprintf("if len(%s)>0 {", srcSelector),
		conv,
		"}",
	}

	return score / M, strings.Join(convs, "\n")
}

func (fm *FuncMaker) sliceAndSlice(dstT, srcT TypeSlice, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	index = nextIndex(index)

	score, conv := fm.makeFunc(Type{typ: dstT.typ.Elem()}, Type{typ: srcT.typ.Elem()},
		dstSelector+"["+index+"]",
		srcSelector+"["+index+"]",
		index,
		history,
	)
	if score == 0 {
		return 0, ""
	}
	dt, err := fm.formatPkgType(dstT.typ)
	if err != nil {
		return 0, ""
	}

	convs := []string{
		fmt.Sprintf("%s = make(%s, len(%s))", dstSelector, dt, srcSelector),
		fmt.Sprintf("for %s := range %s {", index, srcSelector),
		conv,
		"}",
	}

	return score, strings.Join(convs, "\n")
}

func (fm *FuncMaker) named(namedT TypeNamed, selector string) (Type, string) {
	namedT.typ.Obj().Pkg()
	return Type{typ: namedT.typ.Underlying(), name: namedT.typ.String()}, selector
}

func (fm *FuncMaker) namedAndOther(dstT TypeNamed, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	dst, dstSelector := fm.named(dstT, dstSelector)
	return fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
}

func (fm *FuncMaker) otherAndNamed(dst Type, srcT TypeNamed, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	src, srcSelector := fm.named(srcT, srcSelector)
	return fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
}

func (fm *FuncMaker) namedAndNamed(dstT, srcT TypeNamed, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	funcName, err := fm.getFuncName(dstT.typ, srcT.typ)
	if err != nil {
		return 0, ""
	}
	if !fm.isAlreadyExist(funcName) {
		newFM := &FuncMaker{
			pkg:                fm.pkg,
			parentFunc:         fm,
			dstWrittenSelector: map[string]struct{}{},
		}
		tmp := make([]*FuncMaker, 0, 10)
		newFM.childFunc = &tmp

		*fm.childFunc = append(*fm.childFunc, newFM)
		Buf.WriteString("\n" + newFM.MakeFunc(Type{typ: dstT.typ, name: dstT.name}, Type{typ: srcT.typ, name: srcT.name}, false))
	}
	if tolowerFuncName(funcName) == tolowerFuncName(fm.funcName) {
		return fm.makeFunc(Type{typ: dstT.typ.Underlying(), name: dstT.typ.String()}, Type{typ: srcT.typ.Underlying(), name: srcT.typ.String()}, dstSelector, srcSelector, index, history)
	}

	conv := fmt.Sprintf("%s = %s(%s)", dstSelector, tolowerFuncName(funcName), srcSelector)
	fm.dstWrittenSelector[dstSelector] = struct{}{}
	return 1, conv
}

// TODO fix pointer

func (fm *FuncMaker) pointer(pointerT TypePointer, selector string) (Type, string) {
	return Type{typ: pointerT.typ.Elem()}, fmt.Sprintf("(*%s)", selector)
}

func (fm *FuncMaker) pointerAndOther(dstT TypePointer, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	preDstselector := dstSelector
	dst, dstSelector := fm.pointer(dstT, dstSelector)
	score, conv := fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
	if score == 0 {
		return 0, ""
	}

	dt, err := fm.formatPkgType(dstT.typ.Elem())
	if err != nil {
		return 0, ""
	}
	convs := []string{
		fmt.Sprintf("%s = new(%s)", preDstselector, dt),
		conv,
	}
	return score, strings.Join(convs, "\n")
}

func (fm *FuncMaker) otherAndPointer(dst Type, srcT TypePointer, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	preSrcSelector := srcSelector
	src, srcSelector := fm.pointer(srcT, srcSelector)
	score, conv := fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
	if score == 0 {
		return 0, ""
	}

	convs := []string{
		fmt.Sprintf("if %s != nil {", preSrcSelector),
		conv,
		"}",
	}
	return score, strings.Join(convs, "\n")
}

func (fm *FuncMaker) pointerAndPointer(dstT, srcT TypePointer, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	preDstselector := dstSelector
	dst, dstSelector := fm.pointer(dstT, dstSelector)
	preSrcSelector := srcSelector
	src, srcSelector := fm.pointer(srcT, srcSelector)
	score, conv := fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
	if score == 0 {
		return 0, ""
	}
	dt, err := fm.formatPkgType(dstT.typ.Elem())
	if err != nil {
		return 0, ""
	}

	convs := []string{

		fmt.Sprintf("if %s != nil {", preSrcSelector),
		fmt.Sprintf("%s = new(%s)", preDstselector, dt),
		conv,
		fmt.Sprintf("}"),
	}

	return score, strings.Join(convs, "\n")
}
