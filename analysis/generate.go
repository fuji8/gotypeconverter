package analysis

import (
	"bytes"
	"fmt"
	"go/types"
	"text/template"
)

func selectorGen(selector string, field *types.Var) string {
	return fmt.Sprintf("%s.%s", selector, field.Name())
}

// FuncMaker generate function
type FuncMaker struct {
	funcName string
	// output package
	pkg *types.Package

	parentFunc *FuncMaker
	childFunc  *[]*FuncMaker

	// 同じselectorに対して書き込むのは一回のみ
	dstWrittenSelector map[string]struct{}
}

func (fm *FuncMaker) Pkg() *types.Package {
	return fm.pkg
}

func InitFuncMaker(pkg *types.Package) *FuncMaker {
	Buf = new(bytes.Buffer)
	fm := &FuncMaker{
		pkg:                pkg,
		dstWrittenSelector: map[string]struct{}{},
	}
	tmp := make([]*FuncMaker, 0, 10)
	fm.childFunc = &tmp

	return fm
}

// MakeFunc make function
// TODO fix only named type
func (fm *FuncMaker) MakeFunc(dstType, srcType Type, export bool) string {
	dstName, _ := fm.formatPkgType(dstType.typ)
	srcName, _ := fm.formatPkgType(srcType.typ)

	var err error
	fm.funcName, err = fm.getFuncName(dstType.typ, srcType.typ)
	if !export {
		fm.funcName = tolowerFuncName(fm.funcName)
	}
	if err != nil {
		return ""
	}

	_, conv := fm.makeFunc(Type{typ: dstType.typ}, Type{typ: srcType.typ}, "dst", "src", "", nil)
	v := map[string]interface{}{
		"fn": map[string]interface{}{
			"name":    fm.funcName,
			"srcType": srcName,
			"dstType": dstName,
		},
		"body": conv,
	}
	templ := `func {{.fn.name}} (src {{.fn.srcType}}) (dst {{.fn.dstType}}) {
{{.body}}
return
}`
	b := new(bytes.Buffer)
	template.Must(template.New("").Parse(templ)).Execute(b, v)

	return b.String()
}

var Buf *bytes.Buffer

// WriteBytes 全ての関数を書き出す。
func (fm *FuncMaker) WriteBytes() (out []byte) {
	out = Buf.Bytes()
	return
}

func nextIndex(index string) string {
	if index == "" {
		return "i"
	}
	return string(index[0] + 1)
}

type scoreConv struct {
	score float64
	conv  string
}

func convScoreConv(s float64, c string) scoreConv {
	return scoreConv{
		score: s,
		conv:  c,
	}
}

func maxScore(s ...scoreConv) (float64, string) {
	var score float64 = -1
	Mi := -1
	for i, v := range s {
		if v.score > score {
			score = v.score
			Mi = i
		}
	}
	if Mi == -1 {
		return 0, ""
	}
	return s[Mi].score, s[Mi].conv
}

func (fm *FuncMaker) makeFunc(dst, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) (float64, string) {
	if fm.dstWritten(dstSelector) {
		return 0, ""
	}

	if checkHistory(dst.typ, src.typ, history) {
		return 0, ""
	}
	history = append(history, [2]types.Type{dst.typ, src.typ})

	if types.IdenticalIgnoreTags(dst.typ, src.typ) {
		var conv string
		if dst.name != "" && dst.name != src.name {
			conv = fmt.Sprintf("%s = %s(%s)", dstSelector, fm.formatPkgString(dst.name), srcSelector)
		} else {
			conv = fmt.Sprintf("%s = %s", dstSelector, srcSelector)
		}

		fm.dstWrittenSelector[dstSelector] = struct{}{}
		return 1, conv
	}

	switch dstT := dst.typ.(type) {
	case *types.Basic:
		switch srcT := src.typ.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.otherAndNamed(dst, TypeNamed{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.otherAndSlice(dst, TypeSlice{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, TypePointer{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		default:
		}

	case *types.Named:
		switch srcT := src.typ.(type) {
		case *types.Basic:
			return fm.namedAndOther(TypeNamed{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.namedAndNamed(TypeNamed{typ: dstT, name: dst.name}, TypeNamed{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.namedAndOther(TypeNamed{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.namedAndOther(TypeNamed{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.namedAndOther(TypeNamed{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		default:
			return fm.namedAndOther(TypeNamed{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		}

	case *types.Slice:
		switch srcT := src.typ.(type) {
		case *types.Basic:
			return fm.sliceAndOther(TypeSlice{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.otherAndNamed(dst, TypeNamed{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.sliceAndSlice(TypeSlice{typ: dstT, name: dst.name}, TypeSlice{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return maxScore(scoreConv(
				convScoreConv(fm.sliceAndOther(TypeSlice{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history))),
				convScoreConv(fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)),
			)
		case *types.Pointer:
			return fm.otherAndPointer(dst, TypePointer{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		default:
			return fm.sliceAndOther(TypeSlice{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		}

	case *types.Struct:
		switch srcT := src.typ.(type) {
		case *types.Basic:
			return fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.otherAndNamed(dst, TypeNamed{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return maxScore(
				convScoreConv(fm.otherAndSlice(dst, TypeSlice{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)),
				convScoreConv(fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)),
			)
		case *types.Struct:
			return maxScore(
				convScoreConv(fm.structAndStruct(TypeStruct{typ: dstT, name: dst.name}, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)),
				convScoreConv(fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)),
				convScoreConv(fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)),
			)
		case *types.Pointer:
			return fm.otherAndPointer(dst, TypePointer{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		default:
			return fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		}

	case *types.Pointer:
		switch srcT := src.typ.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.pointerAndOther(TypePointer{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.pointerAndOther(TypePointer{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.pointerAndOther(TypePointer{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.pointerAndPointer(TypePointer{typ: dstT, name: dst.name}, TypePointer{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		default:
			return fm.pointerAndOther(TypePointer{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		}

	default:
		switch srcT := src.typ.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.otherAndNamed(dst, TypeNamed{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.otherAndSlice(dst, TypeSlice{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, TypePointer{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
		default:
		}

	}

	return 0, ""
}
