package main

import (
	"github.com/fuji8/gotypeconverter"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	gotypeconverter.Init()
	singlegenerator.Main(gotypeconverter.Generator) // os.Exit
}
