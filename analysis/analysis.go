package analysis

import (
	"fmt"
	"go/types"
	"regexp"
	"strings"
)

func (fm *FuncMaker) getFuncName(dstType, srcType types.Type) string {
	dstName := fm.formatPkgType(dstType)
	srcName := fm.formatPkgType(srcType)

	re := regexp.MustCompile(`\.`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("")))

	re = regexp.MustCompile(`\[\]`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("S")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("S")))

	re = regexp.MustCompile(`\*`)
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("P")))
	dstName = string(re.ReplaceAll([]byte(dstName), []byte("P")))

	return fmt.Sprintf("Conv%sTo%s", srcName, dstName)
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

// TODO *types.Var -> *types.Package
func (fm *FuncMaker) pkgVisiable(field *types.Var) bool {
	if fm.pkg.Path() == field.Pkg().Path() {
		return true
	}
	return field.Exported()
}

func (fm *FuncMaker) formatPkgString(str string) string {
	// TODO fix only pointer, slice and badic
	re := regexp.MustCompile(`[\w\./]*/`)
	last := string(re.ReplaceAll([]byte(str), []byte("")))

	tmp := strings.Split(last, ".")
	p := string(regexp.MustCompile(`\[\]|\*`).ReplaceAll([]byte(tmp[0]), []byte("")))

	path := strings.Split(fm.pkg.Path(), "/")
	if p == path[len(path)-1] {
		re := regexp.MustCompile(`[\w]*\.`)
		return string(re.ReplaceAll([]byte(last), []byte("")))
	}
	return last

}

func (fm *FuncMaker) formatPkgType(t types.Type) string {
	return fm.formatPkgString(t.String())
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
