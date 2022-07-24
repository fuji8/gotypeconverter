package analysis

import (
	"errors"
	"fmt"
	"go/types"
	"regexp"
	"strings"
)

func (fm *FuncMaker) getFuncName(dstType, srcType types.Type) (string, error) {
	dstName, derr := fm.formatPkgType(dstType)
	srcName, serr := fm.formatPkgType(srcType)
	var err error
	if derr != nil || serr != nil {
		err = errors.New("cannot type")
	}

	re := regexp.MustCompile(`\.`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("")))

	re = regexp.MustCompile(`\[\]`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("S")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("S")))

	re = regexp.MustCompile(`\*`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("P")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("P")))

	return fmt.Sprintf("Conv%sTo%s", srcName, dstName), err
}

func (fm *FuncMaker) isAlreadyExist(funcName string) bool {
	// 1. rootまで遡る。
	var root *FuncMaker
	var goBackRoot func(*FuncMaker) *FuncMaker
	goBackRoot = func(fm *FuncMaker) *FuncMaker {
		if fm.parentFunc == nil {
			return fm
		}
		return goBackRoot(fm.parentFunc)
	}
	root = goBackRoot(fm)

	// 2. 存在しているか見る。
	var inspectSamaFuncName func(*FuncMaker) bool
	inspectSamaFuncName = func(fm *FuncMaker) bool {
		if fm.funcName == funcName {
			return true
		}
		if fm.childFunc != nil {
			for _, child := range *fm.childFunc {
				exist := inspectSamaFuncName(child)
				if exist {
					return true
				}
			}
		}
		return false
	}
	return inspectSamaFuncName(root)
}

func (fm *FuncMaker) varVisiable(v *types.Var) bool {
	if fm.samePkg(v.Pkg()) {
		return true
	}
	return v.Exported()
}

func (fm *FuncMaker) typeNameVisiable(v *types.TypeName) bool {
	if fm.samePkg(v.Pkg()) {
		return true
	}
	return v.Exported()
}

func (fm *FuncMaker) samePkg(pkg *types.Package) bool {
	return fm.pkg.Path() == pkg.Path()
}

// TODO fix
func (fm *FuncMaker) formatPkgString(fullTypeStr string) string {
	// TODO fix only pointer, slice and basic
	re := regexp.MustCompile(`[\w-_\./]*/`)
	last := string(re.ReplaceAll([]byte(fullTypeStr), []byte("")))

	tmp := strings.Split(last, ".")
	p := string(regexp.MustCompile(`\[\]|\*`).ReplaceAll([]byte(tmp[0]), []byte("")))

	re = regexp.MustCompile(`[\w-]*\.`)
	typeStr := string(re.ReplaceAll([]byte(last), []byte("")))
	path := strings.Split(fm.pkg.Path(), "/")
	if p == path[len(path)-1] {
		return typeStr
	}
	return last
}

func (fm *FuncMaker) formatPkgType(t types.Type) (string, error) {
	if namedT, ok := t.(*types.Named); ok {
		if !fm.typeNameVisiable(namedT.Obj()) {
			return "", errors.New("not exported")
		}
	}
	return fm.formatPkgString(t.String()), nil
}

// 無限ループを防ぐ
func checkHistory(dst, src types.Type, history [][2]types.Type) bool {
	for _, his := range history {
		if types.Identical(his[0], dst) && types.Identical(his[1], src) {
			return true
		}
	}
	return false
}

func (fm *FuncMaker) dstWritten(dstSelector string) bool {
	_, ok := fm.dstWrittenSelector[dstSelector]
	if ok {
		return true
	}

	// 前方一致
	// TODO fix pointer selector
	for sel := range fm.dstWrittenSelector {

		// TODO fix replace * ( ) [ ] . -> \* \( ...
		re := regexp.MustCompile(fmt.Sprintf(`\^%s[\.\(\[]`,
			strings.Replace(sel, "*", "\\*", -1)))
		written := re.Match([]byte(dstSelector))
		if written {
			return true
		}
	}
	return false
}
