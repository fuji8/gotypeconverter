package goconvertstruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"os"

	"github.com/gostaticanalysis/codegen"
)

const doc = "goconvertstruct is ..."

var (
	flagOutput string

	FlagSrc, FlagDst string
)

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file name")
	Generator.Flags.StringVar(&FlagSrc, "s", "", "source struct")
	Generator.Flags.StringVar(&FlagDst, "d", "", "destination struct")
}

var Generator = &codegen.Generator{
	Name: "goconvertstruct",
	Doc:  doc,
	Run:  run,
}

func run(pass *codegen.Pass) error {
	var buf bytes.Buffer

	// ASTを探索
	var srcAST, dstAST *ast.TypeSpec
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			ast.Print(pass.Fset, n)
			fmt.Println() // \n したい...
			return false
		})

		ast.Inspect(f, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				switch ts.Name.Name {
				case FlagSrc:
					if _, ok := ts.Type.(*ast.StructType); ok {
						srcAST = ts
					}
				case FlagDst:
					if _, ok := ts.Type.(*ast.StructType); ok {
						dstAST = ts
					}
				}
			}
			return true
		})
		if srcAST != nil && dstAST != nil {
			break
		}
	}

	fmt.Println(srcAST, dstAST)
	srcType := pass.TypesInfo.TypeOf(srcAST.Type)
	dstType := pass.TypesInfo.TypeOf(dstAST.Type)
	// 生成
	makeFunc(dstType, srcType, "dst", "src")

	// s, _ := srcS.Type.(*ast.StructType)
	// fi := s.Fields.List[0]
	// fmt.Println(pass.TypesInfo.TypeOf(fi.Type).Underlying().String())
	// fmt.Println(pass.TypesInfo.ObjectOf(fi.Type.(*ast.Ident)).Type().Underlying())
	// fmt.Printf("%T\n", pass.TypesInfo.TypeOf(fi.Type).Underlying())
	// fmt.Println(dstS)
	// fmt.Fprintln(&buf, pass.TypesInfo.TypeOf(fi.Type))

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	if flagOutput == "" {
		pass.Print(string(src))
		return nil
	}

	f, err := os.Create(flagOutput)
	if err != nil {
		return err
	}

	fmt.Fprint(f, string(src))

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func selectorGen(selector string, field *types.Var) string {
	if field.Embedded() {
		return selector
	}
	return fmt.Sprintf("%s.%s", selector, field.Name())
}

func makeFunc(dst, src types.Type, dstSelector, srcSelector string) bool {
	if types.Identical(dst, src) {
		// same
		fmt.Printf("%s = %s\n", dstSelector, srcSelector)
		return true
	}

	switch dstT := dst.(type) {
	case *types.Basic:

		switch srcT := src.(type) {
		case *types.Struct:
			for j := 0; j < srcT.NumFields(); j++ {
				written := makeFunc(dstT, srcT.Field(j).Type(),
					dstSelector,
					fmt.Sprintf("%s.%s", srcSelector, srcT.Field(j).Name()),
				)
				if written {
					return true
				}
			}

		}
	case *types.Array:

	case *types.Struct:
		switch srcT := src.(type) {
		case *types.Basic:
			for i := 0; i < dstT.NumFields(); i++ {
				written := makeFunc(dstT.Field(i).Type(), srcT,
					fmt.Sprintf("%s.%s", dstSelector, dstT.Field(i).Name()),
					srcSelector,
				)
				if written {
					return true
				}

			}
		case *types.Struct:
			for i := 0; i < dstT.NumFields(); i++ {
				if dstT.Field(i).Embedded() {
					makeFunc(dstT.Field(i).Type(), srcT,
						fmt.Sprintf("%s.%s", dstSelector, dstT.Field(i).Name()),
						srcSelector,
					)
					continue
				}
				for j := 0; j < srcT.NumFields(); j++ {
					if dstT.Field(i).Name() == srcT.Field(j).Name() {
						makeFunc(dstT.Field(i).Type(), srcT.Field(j).Type(),
							fmt.Sprintf("%s.%s", dstSelector, dstT.Field(i).Name()),
							fmt.Sprintf("%s.%s", srcSelector, srcT.Field(j).Name()),
						)
					}
				}
			}

		}
	case *types.Named:
		makeFunc(dstT.Underlying(), src, dstSelector, srcSelector)
	}
	return false
}
