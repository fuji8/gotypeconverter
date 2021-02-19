package gotypeconverter

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/imports"
)

const doc = "gotypeconverter is ..."

var (
	flagOutput string

	flagSrc, flagDst string
	// flagImportPkg 解析に必要なためのpkg
	flagImportPkg string

	tmpFilePath    string
	uniqueFuncName string

	ops uint64 = 0
)

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file name")
	Generator.Flags.StringVar(&flagSrc, "s", "", "source struct")
	Generator.Flags.StringVar(&flagDst, "d", "", "destination struct")
}

func CreateTmpFile(path string) {
	ops = 0

	// tmpFilePath = path + "/tmp-001.go"
	// rand.Seed(time.Now().UnixNano())
	tmpFilePath = fmt.Sprintf("%s/tmp%03d.go", path, rand.Int63n(1e3))
	fullPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	pkg := filepath.Base(fullPath)

	src := fmt.Sprintf("package %s\n", pkg)
	uniqueFuncName = fmt.Sprintf("unique%03d", rand.Int63n(1e3))
	src += fmt.Sprintf("func %s(){var (a %s\n b %s\n)\nfmt.Println(a, b)}\n",
		uniqueFuncName, flagSrc, flagDst)

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
	Name: "gotypeconverter",
	Doc:  doc,
	Run:  run,
}

func run(pass *codegen.Pass) error {
	// delete tmp file
	defer func() {
		os.Remove(tmpFilePath)
	}()

	var srcAST, dstAST ast.Expr
	for _, f := range pass.Files {
		// TODO read tmp-001.go only
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Name.Name != uniqueFuncName {
					continue
				}

				//ast.Inspect(fd, func(n ast.Node) bool {
				//ast.Print(pass.Fset, n)
				//fmt.Println() // \n したい...
				//return false
				//})

				ast.Inspect(fd, func(n ast.Node) bool {
					if gd, ok := n.(*ast.GenDecl); ok {
						//gd.Specs[0].(*)
						for _, s := range gd.Specs {
							s, ok := s.(*ast.ValueSpec)
							if !ok {
								return false
							}
							switch s.Names[0].Name {
							case "a":
								srcAST = s.Type
							case "b":
								dstAST = s.Type
							}
						}
					}
					return true
				})
			}
		}
		if srcAST != nil && dstAST != nil {
			break
		}
	}

	if srcAST == nil || dstAST == nil {
		return errors.New("-s or -d are invalid")
	}
	if atomic.LoadUint64(&ops) != 0 {
		return nil
	}
	// ファイルを書くのは、一回のみ
	atomic.AddUint64(&ops, 1)

	outPkg := pass.Pkg.Name()
	buf := &bytes.Buffer{}

	srcType := pass.TypesInfo.TypeOf(srcAST)
	dstType := pass.TypesInfo.TypeOf(dstAST)
	// 生成
	fmt.Fprintf(buf, "// Code generated by ...\n")
	fmt.Fprintf(buf, "package %s\n", outPkg)

	funcMaker := &FuncMaker{
		buf: new(bytes.Buffer),
		pkg: outPkg,
	}
	funcMaker.MakeFunc(dstType, srcType)

	if flagOutput == "" {
		buf.Write(funcMaker.buf.Bytes())

		src, err := imports.Process(tmpFilePath, buf.Bytes(), &imports.Options{
			Fragment: true,
			Comments: true,
		})
		if err != nil {
			return err
		}

		pass.Print(string(src))
		return nil
	}

	var src []byte
	if output, err := ioutil.ReadFile(flagOutput); err == nil {
		// already exist
		output = append(output, funcMaker.buf.Bytes()...)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, flagOutput, output, parser.ParseComments)
		if err != nil {
			return err
		}

		// delete same name func
		funcDeclMap := make(map[string]*ast.FuncDecl)
		for _, d := range file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				funcDeclMap[fd.Name.Name] = fd
			}
		}
		newDecls := make([]ast.Decl, 0)
		for _, d := range file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if _, ok := funcDeclMap[fd.Name.Name]; ok {
					continue
				}
			}

			newDecls = append(newDecls, d)
		}
		for _, lastFd := range funcDeclMap {
			newDecls = append(newDecls, lastFd)
		}
		file.Decls = newDecls

		// sort function
		sort.Slice(file.Decls, func(i, j int) bool {
			fdi, iok := file.Decls[i].(*ast.FuncDecl)
			if !iok {
				return true
			}
			fdj, jok := file.Decls[j].(*ast.FuncDecl)
			if !jok {
				return false
			}
			return fdi.Name.Name < fdj.Name.Name
		})

		dst := new(bytes.Buffer)
		err = format.Node(dst, fset, file)
		if err != nil {
			return err
		}

		src = dst.Bytes()
	} else {
		buf.Write(funcMaker.buf.Bytes())
		src = buf.Bytes()
	}
	// TODO fix
	src, err := imports.Process(flagOutput, src, &imports.Options{
		Fragment: true,
		Comments: true,
	})
	src, _ = format.Source(src)
	if err != nil {
		return err
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

// FuncMaker generate function
type FuncMaker struct {
	buf *bytes.Buffer

	// output package
	pkg string
}

// MakeFunc make function
func (fm *FuncMaker) MakeFunc(dstType, srcType types.Type) {
	dstName := fm.formatPkgType(dstType)
	srcName := fm.formatPkgType(srcType)

	re := regexp.MustCompile(`\.|\[\]|\*`)
	srcStructName := re.ReplaceAll([]byte(srcName), []byte(""))
	dstStructName := re.ReplaceAll([]byte(dstName), []byte(""))

	fmt.Fprintf(fm.buf, "func Convert%sTo%s(src %s) (dst %s) {\n",
		srcStructName, dstStructName, srcName, dstName)
	fm.makeFunc(dstType.Underlying(), srcType.Underlying(), "dst", "src", "")
	fmt.Fprintf(fm.buf, "return\n}\n\n")
}

func selectorGen(selector string, field *types.Var) string {
	return fmt.Sprintf("%s.%s", selector, field.Name())
}

// TODO fix name
func typeStep(t types.Type, selector string) (types.Type, string) {
	switch ty := t.(type) {
	case *types.Named:
		return ty.Underlying(), selector
	case *types.Pointer:
		return typeStep(ty.Elem(), selector)
	}
	return t, selector
}

func (fm *FuncMaker) pkgVisiable(field *types.Var) bool {
	if fm.pkg == field.Pkg().Name() {
		return true
	}
	return field.Exported()
}

func (fm *FuncMaker) formatPkgType(t types.Type) string {
	// TODO fix only pointer, slice and badic
	re := regexp.MustCompile(`[\w\./]*/`)
	last := string(re.ReplaceAll([]byte(t.String()), []byte("")))

	tmp := strings.Split(last, ".")
	p := string(regexp.MustCompile(`\[\]|\*`).ReplaceAll([]byte(tmp[0]), []byte("")))

	if p == fm.pkg {
		re := regexp.MustCompile(`[\w]*\.`)
		return string(re.ReplaceAll([]byte(last), []byte("")))
	}
	return last
}

func (fm *FuncMaker) deferWrite(f func(*FuncMaker) bool) bool {
	tmpFm := &FuncMaker{
		buf: new(bytes.Buffer),
		pkg: fm.pkg,
	}

	written := f(tmpFm)
	if written {
		fm.buf.Write(tmpFm.buf.Bytes())
	}
	return written
}

func nextIndex(index string) string {
	if index == "" {
		return "i"
	}
	return string(index[0] + 1)
}

func (fm *FuncMaker) makeFunc(dst, src types.Type, dstSelector, srcSelector, index string) bool {
	if types.IdenticalIgnoreTags(dst, src) {
		// same
		fmt.Fprintf(fm.buf, "%s = %s\n", dstSelector, srcSelector)
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
			written := false

			for i := 0; i < dstT.NumFields(); i++ {
				if !fm.pkgVisiable(dstT.Field(i)) {
					continue
				}
				if dstT.Field(i).Embedded() {
					written = fm.makeFunc(dstT.Field(i).Type(), src,
						selectorGen(dstSelector, dstT.Field(i)),
						srcSelector,
						index,
					) || written
					continue
				}
				for j := 0; j < srcT.NumFields(); j++ {
					if !fm.pkgVisiable(srcT.Field(j)) {
						continue
					}
					if srcT.Field(j).Embedded() {
						if i == 0 {
							written = fm.makeFunc(dst, srcT.Field(j).Type(),
								dstSelector,
								selectorGen(srcSelector, srcT.Field(j)),
								index,
							) || written
						}
						continue
					}
					if dstT.Field(i).Name() == srcT.Field(j).Name() {
						written = fm.makeFunc(dstT.Field(i).Type(), srcT.Field(j).Type(),
							selectorGen(dstSelector, dstT.Field(i)),
							selectorGen(srcSelector, srcT.Field(j)),
							index,
						) || written
					}
				}
			}
			return written
		// case *types.Array:
		case *types.Slice:
			dstT := dst.(*types.Slice)
			srcT := src.(*types.Slice)

			index = nextIndex(index)

			return fm.deferWrite(func(tmpFm *FuncMaker) bool {
				fmt.Fprintf(tmpFm.buf, "%s = make(%s, len(%s))\n", dstSelector, fm.formatPkgType(dst), srcSelector)
				fmt.Fprintf(tmpFm.buf, "for %s := range %s {\n", index, srcSelector)
				written := tmpFm.makeFunc(dstT.Elem(), srcT.Elem(),
					dstSelector+"["+index+"]",
					srcSelector+"["+index+"]",
					index,
				)
				fmt.Fprintf(tmpFm.buf, "}\n")
				return written
			})
		}
	} else if dstRT.String() == "*types.Slice" || srcRT.String() == "*types.Slice" {
		if dstT, ok := dst.(*types.Slice); ok {
			return fm.deferWrite(func(tmpFm *FuncMaker) bool {
				fmt.Fprintf(tmpFm.buf, "%s = make(%s, 1)\n", dstSelector, fm.formatPkgType(dst))
				return tmpFm.makeFunc(dstT.Elem(), src, dstSelector+"[0]", srcSelector, index)
			})
		} else if srcT, ok := src.(*types.Slice); ok {
			return fm.deferWrite(func(tmpFm *FuncMaker) bool {
				fmt.Fprintf(tmpFm.buf, "if len(%s)>=1 {\n", srcSelector)
				written := tmpFm.makeFunc(dst, srcT.Elem(), dstSelector, srcSelector+"[0]", index)
				fmt.Fprintln(tmpFm.buf, "}")
				return written
			})
		}
	} else if dstRT.String() == "*types.Struct" || srcRT.String() == "*types.Struct" {

		if dstT, ok := dst.(*types.Struct); ok {
			for i := 0; i < dstT.NumFields(); i++ {
				if dstT.Field(i).Embedded() {
					written := fm.makeFunc(dstT.Field(i).Type(), src,
						selectorGen(dstSelector, dstT.Field(i)),
						srcSelector,
						index,
					)
					if written {
						return true
					}
				}
			}
		} else if srcT, ok := src.(*types.Struct); ok {
			for j := 0; j < srcT.NumFields(); j++ {
				written := fm.makeFunc(dst, srcT.Field(j).Type(),
					dstSelector,
					selectorGen(srcSelector, srcT.Field(j)),
					index,
				)
				if written {
					return true
				}
			}
		}
	}
	return false
}
