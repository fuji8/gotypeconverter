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

type flagValue struct {
	s      string
	d      string
	o      string
	inport string
}

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	m := map[string]flagValue{
		"external": flagValue{
			s:      "[]echo.Echo",
			d:      "externalDst",
			o:      "",
			inport: "",
		},
	}

	fileInfos, err := ioutil.ReadDir(codegentest.TestData() + "/src")
	if err != nil {
		panic(err)
	}
	for _, fi := range fileInfos {
		if fi.IsDir() && fi.Name() == "external" {
			fv, ok := m[fi.Name()]
			if !ok {
				goconvertstruct.Generator.Flags.Set("s", fi.Name()+"Src")
				goconvertstruct.Generator.Flags.Set("d", fi.Name()+"Dst")
				goconvertstruct.Generator.Flags.Set("o", "")
				goconvertstruct.Generator.Flags.Set("import", "")
			} else {
				goconvertstruct.Generator.Flags.Set("s", fv.s)
				goconvertstruct.Generator.Flags.Set("d", fv.d)
				goconvertstruct.Generator.Flags.Set("o", fv.o)
				goconvertstruct.Generator.Flags.Set("import", fv.inport)
			}

			goconvertstruct.CreateTmpFile(codegentest.TestData() + "/src/" + fi.Name())

			rs := codegentest.Run(t, codegentest.TestData(), goconvertstruct.Generator, fi.Name())
			codegentest.Golden(t, rs, flagUpdate)
		}
	}
}
