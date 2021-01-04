package goconvertstruct_test

import (
	"flag"
	"os"
	"testing"

	"github.com/fuji8/goconvertstruct"
	"github.com/gostaticanalysis/codegen/codegentest"
	"github.com/stretchr/testify/assert"
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

	rs := codegentest.Run(t, "/home/fuji/go", goconvertstruct.Generator, "../workspace/tools/goconvertstruct/testdata/src/normal")
	assert.Equal(t, 1, len(rs))
	rs[0].Dir = "/home/fuji/workspace/tools/goconvertstruct/testdata/src/normal"
	codegentest.Golden(t, rs, flagUpdate)
}
