package goconvertstruct_test

import (
	"flag"
	"os"
	"testing"

	"github.com/fuji8/goconvertstruct"
	"github.com/gostaticanalysis/codegen/codegentest"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	goconvertstruct.FlagSrc = "a"
	goconvertstruct.FlagDst = "b"

	rs := codegentest.Run(t, codegentest.TestData(), goconvertstruct.Generator, "normal")
	codegentest.Golden(t, rs, flagUpdate)
}
