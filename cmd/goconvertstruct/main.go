package main

import (
	"github.com/fuji8/goconvertstruct"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	goconvertstruct.Init()
	singlegenerator.Main(goconvertstruct.Generator) // os.Exit
}
