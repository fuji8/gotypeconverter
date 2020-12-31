package main

import (
	"github.com/fuji8/goconvertstruct"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	singlegenerator.Main(goconvertstruct.Generator)
}
