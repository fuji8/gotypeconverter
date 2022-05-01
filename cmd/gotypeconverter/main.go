package main

import (
	"os"

	"github.com/fuji8/gotypeconverter"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	gotypeconverter.Init()
	defer func() {
		os.Remove(gotypeconverter.TmpFilePath)
	}()
	singlegenerator.Main(gotypeconverter.Generator)
}
