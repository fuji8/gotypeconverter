package main

import (
	"github.com/fuji8/gotypeconverter"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	singlegenerator.Main(gotypeconverter.Generator)
}
