package gotypeconverter

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	ana "github.com/fuji8/gotypeconverter/analysis"
	"github.com/fuji8/gotypeconverter/ui"
	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/go/analysis"
)

const (
	doc     = "gotypeconverter generates a function that converts two different named types."
	version = "v0.1.12"
)

var (
	flagOutput  string
	flagVersion bool

	flagSrc, flagDst, flagPkg, flagStructTag string

	TmpFilePath    string
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

var Generator = &codegen.Generator{
	Name:             "gotypeconverter",
	Doc:              doc,
	Run:              run,
	RunDespiteErrors: true,
}

func TypeOf3(pass *analysis.Pass, pkg, name string) types.Type {
	if name == "" {
		return nil
	}

	if name[0] == '*' {
		obj := TypeOf3(pass, pkg, name[1:])
		if obj == nil {
			return nil
		}
		return types.NewPointer(obj)
	}

	if name[0] == '[' {
		obj := TypeOf3(pass, pkg, name[1:])
		if obj == nil {
			return nil
		}
		return types.NewSlice(obj)
	}

	if pkg == "" {
		obj := pass.Pkg.Scope().Lookup(name)
		return obj.Type()
	}

	var obj types.Object
	for i := 0; i < pass.Pkg.Scope().NumChildren(); i++ {
		tpkg, ok := pass.Pkg.Scope().Child(i).Lookup(pkg).(*types.PkgName)
		if !ok {
			return nil
		}
		obj = tpkg.Imported().Scope().Lookup(name)
		if obj != nil {
			break
		}
	}
	return obj.Type()
}

func run(pass *codegen.Pass) error {
	aPass := analysis.Pass{
		Analyzer:          pass.Generator.ToAnalyzer(),
		Fset:              pass.Fset,
		Files:             pass.Files,
		OtherFiles:        pass.OtherFiles,
		IgnoredFiles:      []string{},
		Pkg:               pass.Pkg,
		TypesInfo:         pass.TypesInfo,
		TypesSizes:        pass.TypesSizes,
		ResultOf:          pass.ResultOf,
		ImportObjectFact:  pass.ImportObjectFact,
		ImportPackageFact: pass.ImportPackageFact,
	}

	srcTstr := strings.Split(flagSrc, ".")
	var srcType types.Type
	if len(srcTstr) == 1 {
		srcType = TypeOf3(&aPass, "", srcTstr[0])
	} else if len(srcTstr) == 2 {
		srcType = TypeOf3(&aPass, srcTstr[0], srcTstr[1])
	}
	dstTstr := strings.Split(flagDst, ".")
	var dstType types.Type
	if len(dstTstr) == 1 {
		dstType = TypeOf3(&aPass, "", dstTstr[0])
	} else if len(dstTstr) == 2 {
		dstType = TypeOf3(&aPass, dstTstr[0], dstTstr[1])
	}

	funcMaker := ana.InitFuncMaker(pass.Pkg)
	funcMaker.MakeFunc(ana.InitType(dstType, flagDst), ana.InitType(srcType, flagSrc))

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
