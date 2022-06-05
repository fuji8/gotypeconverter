package gotypeconverter

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"testing"

	"github.com/fuji8/gotypeconverter/ui"
	"github.com/gostaticanalysis/codegen/codegentest"
	"golang.org/x/tools/go/packages"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	Generator.Flags.Set("s", "SRC")
	Generator.Flags.Set("d", "DST")

	ui.TmpFilePath = codegentest.TestData() + "/src/a/tmp.go"
	rs := codegentest.Run(t, codegentest.TestData(), Generator, "a")
	codegentest.Golden(t, rs, flagUpdate)
}

func TypeOf4(pkg *types.Package, pkgname, typename string) types.Type {
	if typename == "" {
		return nil
	}

	if typename[0] == '*' {
		obj := TypeOf4(pkg, pkgname, typename[1:])
		if obj == nil {
			return nil
		}
		return types.NewPointer(obj)
	}

	if typename[0] == '[' {
		obj := TypeOf4(pkg, pkgname, typename[1:])
		if obj == nil {
			return nil
		}
		return types.NewSlice(obj)
	}
	if pkgname == "" {
		obj := pkg.Scope().Lookup(typename)
		return obj.Type()
	}

	obj := pkg.Scope().Lookup(typename)
	if obj == nil {
		return nil
	}
	return obj.Type()
}

func TestXXX(t *testing.T) {
	tmp, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  codegentest.TestData() + "/src/a",
	}, "a")
	tmp2 := TypeOf4(tmp[0].Types, "", "SRC")
	fmt.Println(tmp2, err)
	fmt.Println(tmp2.String())
}

func TestYYY(t *testing.T) {
	tmp, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  codegentest.TestData() + "/issue/x016",
	}, "x016")
	fmt.Println(tmp, err)
}
