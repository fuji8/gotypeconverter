package gotypeconverter

import (
	"errors"
	"fmt"
	"go/ast"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	ana "github.com/fuji8/gotypeconverter/analysis"
	"github.com/fuji8/gotypeconverter/ui"
	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/imports"
)

const (
	doc     = "gotypeconverter generates a function that converts two different named types."
	version = "v0.1.12"
)

var (
	flagOutput  string
	flagVersion bool

	flagSrc, flagDst, flagPkg, flagStructTag string

	tmpFilePath    string
	uniqueFuncName string

	ops uint64 = 0
)

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file; if nil, output stdout")
	Generator.Flags.StringVar(&flagSrc, "s", "", "source type")
	Generator.Flags.StringVar(&flagDst, "d", "", "destination type")
	Generator.Flags.BoolVar(&flagVersion, "v", false, "version")
	Generator.Flags.StringVar(&flagPkg, "pkg", "", "output package; if nil, the directoryName and packageName must be same and will be used")
	Generator.Flags.StringVar(&flagStructTag, "structTag", "cvt", "")
}

func CreateTmpFile(path string) {
	ops = 0

	// tmpFilePath = path + "/tmp-001.go"
	rand.Seed(time.Now().UnixNano())
	tmpFilePath = fmt.Sprintf("%s/tmp%03d.go", path, rand.Int63n(1e3))
	fullPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	pkg := flagPkg
	if flagPkg == "" {
		pkg = filepath.Base(fullPath)
	}

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

	if flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if Generator.Flags.NArg() == 0 {
		return
	}

	path := os.Args[len(os.Args)-1]
	CreateTmpFile(path)
}

var Generator = &codegen.Generator{
	Name:             "gotypeconverter",
	Doc:              doc,
	Run:              run,
	RunDespiteErrors: true,
}

func run(pass *codegen.Pass) error {
	ui.TmpFilePath = tmpFilePath[:len(tmpFilePath)-3] + "_generated.go"

	// delete tmp file
	defer func() {
		os.Remove(tmpFilePath)
	}()

	var srcAST, dstAST ast.Expr
	existTargeFile := false
	for _, f := range pass.Files {
		// TODO read tmp*.go only
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				if fd.Name.Name != uniqueFuncName {
					continue
				}

				existTargeFile = true

				//ast.Inspect(fd, func(n ast.Node) bool {
				//ast.Print(pass.Fset, n)
				//fmt.Println() // \n したい...
				//return false
				//})

				ast.Inspect(fd, func(n ast.Node) bool {
					if gd, ok := n.(*ast.GenDecl); ok {
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
				break
			}
		}
		if existTargeFile {
			break
		}
	}

	if !existTargeFile {
		// 解析対象のpassでない
		return nil
	}

	if srcAST == nil || dstAST == nil {
		return errors.New("-s or -d are invalid")
	}
	if atomic.LoadUint64(&ops) != 0 {
		return nil
	}
	// ファイルを書くのは、一回のみ
	atomic.AddUint64(&ops, 1)

	srcType := pass.TypesInfo.TypeOf(srcAST)
	dstType := pass.TypesInfo.TypeOf(dstAST)

	funcMaker := ana.InitFuncMaker(pass.Pkg)
	funcMaker.MakeFunc(ana.InitType(dstType, flagDst), ana.InitType(srcType, flagSrc), true)

	if flagOutput == "" {
		src, err := ui.NoInfoGeneration(funcMaker)
		if err != nil {
			return err
		}
		pass.Print(src)
		return nil
	}

	src, err := ui.FileNameGeneration(funcMaker, flagOutput)
	if err != nil {
		return err
	}

	f, err := os.Create(flagOutput)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprint(f, src)

	return nil
}
