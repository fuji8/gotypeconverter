package main

import (
	"os"

	"github.com/fuji8/gotypeconverter"
	"github.com/fuji8/gotypeconverter/ui"
	"golang.org/x/tools/go/packages"
)

func main() {
	path := os.Args[len(os.Args)-1]
	gotypeconverter.Gen.Flags.Parse(os.Args[1:])
	ui.TmpFilePath = path + "/tmp.go"
	pkgs, _ := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  path,
	}, ".")
	gotypeconverter.Gen.Run(pkgs)
}
