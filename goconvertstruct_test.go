package goconvertstruct_test

import (
	"flag"
	"io/ioutil"
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

	fileInfos, err := ioutil.ReadDir(codegentest.TestData() + "/src")
	if err != nil {
		panic(err)
	}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			goconvertstruct.Generator.Flags.Set("s", fi.Name()+"SRC")
			goconvertstruct.Generator.Flags.Set("d", fi.Name()+"DST")
			rs := codegentest.Run(t, codegentest.TestData(), goconvertstruct.Generator, fi.Name())
			codegentest.Golden(t, rs, flagUpdate)
		}
	}
}
