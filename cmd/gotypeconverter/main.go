package main

import (
	"os"

	"github.com/fuji8/gotypeconverter"
	"github.com/fuji8/gotypeconverter/ui"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	path := os.Args[len(os.Args)-1]
	ui.TmpFilePath = path + "/tmp.go"
	singlegenerator.Main(gotypeconverter.Generator)
}
