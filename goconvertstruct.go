package goconvertstruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/imports"
)

const doc = "goconvertstruct is ..."

var (
	flagOutput string

	flagSrc, flagDst string
	// flagImportPkg 解析に必要なためのpkg
	flagImportPkg string

	tmpFilePath string
)

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file name")
	Generator.Flags.StringVar(&flagSrc, "s", "", "source struct")
	Generator.Flags.StringVar(&flagDst, "d", "", "destination struct")
	Generator.Flags.StringVar(&flagImportPkg, "import", "hello", "import pkg")
}

func CreateTmpFile(path string) {
	tmpFilePath = path + "/tmp-001.go"
	pkg := filepath.Base(path)

	src := fmt.Sprintf("package %s\n", pkg)
	src += fmt.Sprintf("func unique(){fmt.Println(%s{},%s{})}\n", flagSrc, flagDst)

	// goimports do not imports from go.mod
	res, err := imports.Process(tmpFilePath, []byte(src), &imports.Options{
		Fragment: true,
	})
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(tmpFilePath, res, 0755)
	if err != nil {
		panic(err)
	}

}

// Init 解析のための一時ファイルを作成する
func Init() {
	err := Generator.Flags.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	path := os.Args[len(os.Args)-1]
	CreateTmpFile(path)
}

var Generator = &codegen.Generator{
	Name: "goconvertstruct",
	Doc:  doc,
	Run:  run,
}

var buf bytes.Buffer

func run(pass *codegen.Pass) error {
	// initialize
	buf = bytes.Buffer{}
	t := strings.Split(flagSrc, ".")
	srcStructName := t[len(t)-1]
	t = strings.Split(flagDst, ".")
	dstStructName := t[len(t)-1]

	// delete tmp file
	defer func() {
		os.Remove(tmpFilePath)
	}()

	var srcAST, dstAST *ast.Ident
	for _, f := range pass.Files {
		// TODO read tmp-001.go only
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Name.Name != "unique" {
					continue
				}

				ast.Inspect(fd, func(n ast.Node) bool {
					if ident, ok := n.(*ast.Ident); ok {
						switch ident.Name {
						case srcStructName:
							srcAST = ident
						case dstStructName:
							dstAST = ident
						}
						return false
					}
					return true
				})

			}
		}
		if srcAST != nil && dstAST != nil {
			break
		}
	}

	srcType := pass.TypesInfo.TypeOf(srcAST)
	dstType := pass.TypesInfo.TypeOf(dstAST)
	// 生成
	fmt.Fprintf(&buf, "// Code generated by ...\n")
	fmt.Fprintf(&buf, "package %s\n", pass.Pkg.Name())
	fmt.Fprintf(&buf, "func Convert%sTo%s(src %s) (dst %s) {\n",
		srcStructName, dstStructName, flagSrc, flagDst)

	makeFunc(dstType, srcType, "dst", "src")
	fmt.Fprintf(&buf, "return\n}\n")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("format error:" + buf.String())
		return err
	}
	src, err = imports.Process(tmpFilePath, src, &imports.Options{
		Fragment: true,
		Comments: true,
	})
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
	//if field.Embedded() {
	//return selector
	//}
	return fmt.Sprintf("%s.%s", selector, field.Name())
}

func typeStep(t types.Type, selector string) (types.Type, string) {
	switch ty := t.(type) {
	case *types.Named:
		return ty.Underlying(), selector
	}
	return t, selector
}

func makeFunc(dst, src types.Type, dstSelector, srcSelector string) bool {
	if types.Identical(dst, src) {
		// same
		fmt.Fprintf(&buf, "%s = %s\n", dstSelector, srcSelector)
		return true
	}

	dst, dstSelector = typeStep(dst, dstSelector)
	src, srcSelector = typeStep(src, srcSelector)

	dstRT := reflect.TypeOf(dst)
	srcRT := reflect.TypeOf(src)
	if dstRT.String() == srcRT.String() {
		// same type
		switch dst.(type) {
		case *types.Struct:
			dstT := dst.(*types.Struct)
			srcT := src.(*types.Struct)

			for i := 0; i < dstT.NumFields(); i++ {
				if dstT.Field(i).Embedded() {
					makeFunc(dstT.Field(i).Type(), src,
						selectorGen(dstSelector, dstT.Field(i)),
						srcSelector,
					)
					continue
				}
				for j := 0; j < srcT.NumFields(); j++ {
					// TODO fix
					if dstT.Field(i).Name() == srcT.Field(j).Name() {
						makeFunc(dstT.Field(i).Type(), srcT.Field(j).Type(),
							selectorGen(dstSelector, dstT.Field(i)),
							selectorGen(srcSelector, srcT.Field(j)),
						)
					}
				}
			}
		// case *types.Array:
		case *types.Slice:
			dstT := dst.(*types.Slice)
			srcT := src.(*types.Slice)

			// TODO fix unique i, v
			fmt.Fprintf(&buf, "%s = make(%s, len(%s))\n", dstSelector, dst.String(), srcSelector)
			fmt.Fprintf(&buf, "for i, _ := range %s {\n", srcSelector)
			makeFunc(dstT.Elem(), srcT.Elem(),
				dstSelector+"[i]",
				srcSelector+"[i]")
			fmt.Fprintf(&buf, "}\n")
		}
	} else if dstRT.String() == "*types.Slice" || srcRT.String() == "*types.Slice" {
		if dstT, ok := dst.(*types.Slice); ok {
			fmt.Fprintf(&buf, "%s = make(%s, 1)\n", dstSelector, dst.String())
			return makeFunc(dstT.Elem(), src, dstSelector+"[0]", srcSelector)
		} else if srcT, ok := src.(*types.Slice); ok {
			return makeFunc(dst, srcT.Elem(), dstSelector, srcSelector+"[0]")
		}
	} else if dstRT.String() == "*types.Struct" || srcRT.String() == "*types.Struct" {

		if dstT, ok := dst.(*types.Struct); ok {
			for i := 0; i < dstT.NumFields(); i++ {
				if dstT.Field(i).Embedded() {
					written := makeFunc(dstT.Field(i).Type(), src,
						selectorGen(dstSelector, dstT.Field(i)),
						srcSelector,
					)
					if written {
						return true
					}
				}
			}
		} else if srcT, ok := src.(*types.Struct); ok {
			for j := 0; j < srcT.NumFields(); j++ {
				written := makeFunc(dst, srcT.Field(j).Type(),
					dstSelector,
					selectorGen(srcSelector, srcT.Field(j)),
				)
				if written {
					return true
				}
			}
		}
	}
	return false
}
