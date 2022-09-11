package analysis

import (
	"bytes"
	"fmt"
	"go/types"
)

func selectorGen(selector string, field *types.Var) string {
	return fmt.Sprintf("%s.%s", selector, field.Name())
}

// FuncMaker generate function
type FuncMaker struct {
	funcName string
	buf      *bytes.Buffer
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
	fm := &FuncMaker{
		buf:                new(bytes.Buffer),
		pkg:                pkg,
		dstWrittenSelector: map[string]struct{}{},
	}
	tmp := make([]*FuncMaker, 0, 10)
	fm.childFunc = &tmp

	return fm
}

// MakeFunc make function
// TODO fix only named type
func (fm *FuncMaker) MakeFunc(dstType, srcType Type, export bool) {
	dstName, _ := fm.formatPkgType(dstType.typ)
	srcName, _ := fm.formatPkgType(srcType.typ)

	var err error
	fm.funcName, err = fm.getFuncName(dstType.typ, srcType.typ)
	if !export {
		fm.funcName = tolowerFuncName(fm.funcName)
	}
	if err != nil {
		return
	}

	fmt.Fprintf(fm.buf, "func %s(src %s) (dst %s) {\n",
		fm.funcName, srcName, dstName)
	fm.makeFunc(Type{typ: dstType.typ}, Type{typ: srcType.typ}, "dst", "src", "", nil)
	fmt.Fprintf(fm.buf, "return\n}\n\n")
}

// WriteBytes 全ての関数を書き出す。
func (fm *FuncMaker) WriteBytes() (out []byte) {
	out = fm.buf.Bytes()
	if fm.childFunc != nil {
		for _, child := range *fm.childFunc {
			out = append(out, child.WriteBytes()...)
		}
	}
	return
}

func (fm *FuncMaker) deferWrite(f func(*FuncMaker) bool) bool {
	tmpFm := &FuncMaker{
		funcName:   fm.funcName,
		buf:        new(bytes.Buffer),
		pkg:        fm.pkg,
		parentFunc: fm.parentFunc,
		childFunc:  fm.childFunc,

		dstWrittenSelector: fm.dstWrittenSelector,
	}

	written := f(tmpFm)
	if written {
		fm.buf.Write(tmpFm.buf.Bytes())
		// fm.childFunc = tmpFm.childFunc
		fm.dstWrittenSelector = tmpFm.dstWrittenSelector
	}
	return written
}

func nextIndex(index string) string {
	if index == "" {
		return "i"
	}
	return string(index[0] + 1)
}

func (fm *FuncMaker) makeFunc(dst, src Type, dstSelector, srcSelector, index string, history [][2]types.Type) bool {
	if fm.dstWritten(dstSelector) {
		return false
	}

	if checkHistory(dst.typ, src.typ, history) {
		return false
	}
	history = append(history, [2]types.Type{dst.typ, src.typ})

	if types.IdenticalIgnoreTags(dst.typ, src.typ) {
		if dst.name != "" && dst.name != src.name {
			fmt.Fprintf(fm.buf, "%s = %s(%s)\n", dstSelector, fm.formatPkgString(dst.name), srcSelector)
		} else {
			fmt.Fprintf(fm.buf, "%s = %s\n", dstSelector, srcSelector)
		}

		fm.dstWrittenSelector[dstSelector] = struct{}{}
		return true
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
			return fm.sliceAndOther(TypeSlice{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history) ||
				fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
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
			return fm.otherAndSlice(dst, TypeSlice{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history) ||
				fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.structAndStruct(TypeStruct{typ: dstT, name: dst.name}, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history) ||
				fm.structAndOther(TypeStruct{typ: dstT, name: dst.name}, src, dstSelector, srcSelector, index, history) ||
				fm.otherAndStruct(dst, TypeStruct{typ: srcT, name: src.name}, dstSelector, srcSelector, index, history)
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

	return false
}
