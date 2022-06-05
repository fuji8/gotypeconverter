package gotypeconverter

import (
	"flag"
	"os"
	"testing"

	"github.com/fuji8/gotypeconverter/ui"
	"github.com/gostaticanalysis/codegen/codegentest"
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
