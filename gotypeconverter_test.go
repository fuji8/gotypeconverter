package gotypeconverter

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fuji8/gotypeconverter/ui"
	"github.com/google/go-cmp/cmp"
	"github.com/gostaticanalysis/codegen/codegentest"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

// func TestGenerator(t *testing.T) {
// Gen.Flags.Set("s", "SRC")
// Gen.Flags.Set("d", "DST")

// ui.TmpFilePath = codegentest.TestData() + "/src/a/tmp.go"
// rs := codegentest.Run(t, codegentest.TestData(), Generator, "a")
// codegentest.Golden(t, rs, flagUpdate)
// }

func TestGenerator(t *testing.T) {
	Gen.Flags.Set("s", "SRC")
	Gen.Flags.Set("d", "DST")
	ui.TmpFilePath = codegentest.TestData() + "/src/a/tmp.go"

	pkgs, _ := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  codegentest.TestData() + "/src/a",
	}, "")
	got, err := run(pkgs)
	require.NoError(t, err)

	fpath := fmt.Sprintf("%s.golden", codegentest.TestData()+"/src/a/gotypeconverter")
	gf, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	//fmt.Println(got)
	if diff := cmp.Diff(string(gf), got); diff != "" {
		gname := "gotypeconverter"
		t.Errorf("%s's output is different from the golden file(%s):\n%s", gname, fpath, diff)
	}
}
