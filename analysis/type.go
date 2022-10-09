package analysis

import (
	"bytes"
	"fmt"
	"go/types"
)

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

func (fm *FuncMaker) structAndOther(dstT TypeStruct, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	var written float64 = 0
	for i := 0; i < dstT.typ.NumFields(); i++ {
		if !fm.varVisiable(dstT.typ.Field(i)) {
			continue
		}

		// if struct tag "cvt" exists, use struct tag
		_, _, _, dOption := getTag(dstT.typ.Tag(i))
		if dOption == Ignore || dOption == ReadOnly {
			continue
		}

		written += float64(1/dstT.typ.NumFields()) * fm.makeFunc(Type{typ: dstT.typ.Field(i).Type()}, src,
			selectorGen(dstSelector, dstT.typ.Field(i)),
			srcSelector,
			index,
			history,
		)
		if written > 0 {
			break
		}
	}
	return written
}

func (fm *FuncMaker) otherAndStruct(dst Type, srcT TypeStruct, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	var written float64 = 0
	for j := 0; j < srcT.typ.NumFields(); j++ {
		if !fm.varVisiable(srcT.typ.Field(j)) {
			continue
		}
		// if struct tag "cvt" exists, use struct tag
		_, _, _, sOption := getTag(srcT.typ.Tag(j))
		if sOption == Ignore || sOption == WriteOnly {
			continue
		}

		written += float64(1/srcT.typ.NumFields()) * fm.makeFunc(dst, Type{typ: srcT.typ.Field(j).Type()},
			dstSelector,
			selectorGen(srcSelector, srcT.typ.Field(j)),
			index,
			history,
		)
		if written > 0 {
			break
		}
	}
	return written
}

func (fm *FuncMaker) structAndStruct(dstT, srcT TypeStruct, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	var written float64 = 0

	// field 同士の比較

	for i := 0; i < dstT.typ.NumFields(); i++ {
		if !fm.varVisiable(dstT.typ.Field(i)) {
			continue
		}
		// if struct tag "cvt" exists, use struct tag
		dField, _, dWriteField, dOption := getTag(dstT.typ.Tag(i))
		if dWriteField != "" {
			dField = dWriteField
		}
		if dField == "" {
			dField = dstT.typ.Field(i).Name()
		}
		if dOption == Ignore || dOption == ReadOnly {
			continue
		}

		if dstT.typ.Field(i).Embedded() {
			written += float64(1/dstT.typ.NumFields()) * fm.makeFunc(Type{typ: dstT.typ.Field(i).Type()}, Type{typ: srcT.typ, name: srcT.name},
				selectorGen(dstSelector, dstT.typ.Field(i)),
				srcSelector,
				index,
				history,
			)
			continue
		}
		for j := 0; j < srcT.typ.NumFields(); j++ {
			if !fm.varVisiable(srcT.typ.Field(j)) {
				continue
			}
			// if struct tag "cvt" exists, use struct tag
			sField, sReadField, _, sOption := getTag(srcT.typ.Tag(j))
			if sReadField != "" {
				sField = sReadField
			}
			if sField == "" {
				sField = srcT.typ.Field(j).Name()
			}
			if sOption == Ignore || sOption == WriteOnly {
				continue
			}

			if srcT.typ.Field(j).Embedded() {
				continue
			}

			if dField == sField {
				written += float64(1/dstT.typ.NumFields()) * fm.makeFunc(Type{typ: dstT.typ.Field(i).Type()}, Type{typ: srcT.typ.Field(j).Type()},
					selectorGen(dstSelector, dstT.typ.Field(i)),
					selectorGen(srcSelector, srcT.typ.Field(j)),
					index,
					history,
				)
			}
		}
	}

	for j := 0; j < srcT.typ.NumFields(); j++ {
		if srcT.typ.Field(j).Embedded() {
			_, _, _, sOption := getTag(srcT.typ.Tag(j))
			if sOption == Ignore || sOption == WriteOnly {
				continue
			}

			written = float64(1/dstT.typ.NumFields()) * fm.makeFunc(Type{typ: dstT.typ, name: dstT.name}, Type{typ: srcT.typ.Field(j).Type()},
				dstSelector,
				selectorGen(srcSelector, srcT.typ.Field(j)),
				index,
				history,
			)
		}
	}

	return written
}

func (fm *FuncMaker) sliceAndOther(dstT TypeSlice, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		dt, err := tmpFm.formatPkgType(dstT.typ)
		if err != nil {
			return 0
		}
		fmt.Fprintf(tmpFm.buf, "%s = make(%s, 1)\n", dstSelector, dt)
		return tmpFm.makeFunc(Type{typ: dstT.typ.Elem()}, src, dstSelector+"[0]", srcSelector, index, history)
	})
}

func (fm *FuncMaker) otherAndSlice(dst Type, srcT TypeSlice, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		fmt.Fprintf(tmpFm.buf, "if len(%s)>0 {\n", srcSelector)
		written := tmpFm.makeFunc(dst, Type{typ: srcT.typ.Elem()}, dstSelector, srcSelector+"[0]", index, history)
		fmt.Fprintln(tmpFm.buf, "}")
		return written
	})
}

func (fm *FuncMaker) sliceAndSlice(dstT, srcT TypeSlice, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	index = nextIndex(index)

	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		dt, err := tmpFm.formatPkgType(dstT.typ)
		if err != nil {
			return 0
		}

		fmt.Fprintf(tmpFm.buf, "%s = make(%s, len(%s))\n", dstSelector, dt, srcSelector)
		fmt.Fprintf(tmpFm.buf, "for %s := range %s {\n", index, srcSelector)
		written := tmpFm.makeFunc(Type{typ: dstT.typ.Elem()}, Type{typ: srcT.typ.Elem()},
			dstSelector+"["+index+"]",
			srcSelector+"["+index+"]",
			index,
			history,
		)
		fmt.Fprintf(tmpFm.buf, "}\n")
		if written > 0 {
			tmpFm.dstWrittenSelector[dstSelector] = struct{}{}
		}
		return written
	})
}

func (fm *FuncMaker) named(namedT TypeNamed, selector string) (Type, string) {
	namedT.typ.Obj().Pkg()
	return Type{typ: namedT.typ.Underlying(), name: namedT.typ.String()}, selector
}

func (fm *FuncMaker) namedAndOther(dstT TypeNamed, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	dst, dstSelector := fm.named(dstT, dstSelector)
	return fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
}

func (fm *FuncMaker) otherAndNamed(dst Type, srcT TypeNamed, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	src, srcSelector := fm.named(srcT, srcSelector)
	return fm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
}

func (fm *FuncMaker) namedAndNamed(dstT, srcT TypeNamed, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	funcName, err := fm.getFuncName(dstT.typ, srcT.typ)
	if err != nil {
		return 0
	}
	if !fm.isAlreadyExist(funcName) {
		newFM := &FuncMaker{
			buf:                new(bytes.Buffer),
			pkg:                fm.pkg,
			parentFunc:         fm,
			dstWrittenSelector: map[string]struct{}{},
		}
		tmp := make([]*FuncMaker, 0, 10)
		newFM.childFunc = &tmp

		*fm.childFunc = append(*fm.childFunc, newFM)
		newFM.MakeFunc(Type{typ: dstT.typ, name: dstT.name}, Type{typ: srcT.typ, name: srcT.name}, false)
	}
	if tolowerFuncName(funcName) == tolowerFuncName(fm.funcName) {
		return fm.makeFunc(Type{typ: dstT.typ.Underlying(), name: dstT.typ.String()}, Type{typ: srcT.typ.Underlying(), name: srcT.typ.String()}, dstSelector, srcSelector, index, history)
	}

	fmt.Fprintf(fm.buf, "%s = %s(%s)\n", dstSelector, tolowerFuncName(funcName), srcSelector)
	fm.dstWrittenSelector[dstSelector] = struct{}{}
	return 1
}

// TODO fix pointer

func (fm *FuncMaker) pointer(pointerT TypePointer, selector string) (Type, string) {
	return Type{typ: pointerT.typ.Elem()}, fmt.Sprintf("(*%s)", selector)
}

func (fm *FuncMaker) pointerAndOther(dstT TypePointer, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		selector := dstSelector
		dst, dstSelector := fm.pointer(dstT, dstSelector)
		dt, err := tmpFm.formatPkgType(dstT.typ.Elem())
		if err != nil {
			return 0
		}
		fmt.Fprintf(tmpFm.buf, "%s = new(%s)\n", selector, dt)
		return tmpFm.makeFunc(dst, src, dstSelector, srcSelector, index, history)
	})
}

func (fm *FuncMaker) otherAndPointer(dst Type, srcT TypePointer, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		fmt.Fprintf(tmpFm.buf, "if %s != nil {\n", srcSelector)

		src, srcSelector := fm.pointer(srcT, srcSelector)
		written := tmpFm.makeFunc(dst, src, dstSelector, srcSelector, index, history)

		fmt.Fprintf(tmpFm.buf, "}\n")
		return written
	})
}

func (fm *FuncMaker) pointerAndPointer(dstT, srcT TypePointer, dstSelector, srcSelector, index string, history [][2]types.Type) float64 {
	return fm.deferWrite(func(tmpFm *FuncMaker) float64 {
		fmt.Fprintf(tmpFm.buf, "if %s != nil {\n", srcSelector)

		selector := dstSelector
		dst, dstSelector := fm.pointer(dstT, dstSelector)
		dt, err := tmpFm.formatPkgType(dstT.typ.Elem())
		if err != nil {
			return 0
		}
		fmt.Fprintf(tmpFm.buf, "%s = new(%s)\n", selector, dt)
		src, srcSelector := fm.pointer(srcT, srcSelector)
		written := tmpFm.makeFunc(dst, src, dstSelector, srcSelector, index, history)

		fmt.Fprintf(tmpFm.buf, "}\n")
		return written
	})
}
