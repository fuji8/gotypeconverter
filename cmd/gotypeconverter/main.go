package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/fuji8/gotypeconverter"
	ana "github.com/fuji8/gotypeconverter/analysis"
	"github.com/fuji8/gotypeconverter/ui"
	"golang.org/x/tools/go/packages"
)

func getVersion() string {
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return i.Main.Version
}

func main() {
	gotypeconverter.Gen.Flags.Usage = func() {
		paras := strings.Split(gotypeconverter.Gen.Doc, "\n\n")
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", gotypeconverter.Gen.Name, paras[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [-flag] [package]\n\n", gotypeconverter.Gen.Name)
		if len(paras) > 1 {
			fmt.Fprintln(os.Stderr, strings.Join(paras[1:], "\n\n"))
		}
		fmt.Fprintln(os.Stderr, "\nFlags:")
		gotypeconverter.Gen.Flags.PrintDefaults()
	}

	err := gotypeconverter.Gen.Flags.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
	if gotypeconverter.FlagVersion {
		fmt.Println(getVersion())
		os.Exit(0)
	}
	if gotypeconverter.Gen.Flags.NArg() == 0 {
		os.Exit(1)
	}

	path := os.Args[len(os.Args)-1]
	ui.TmpFilePath = path + "/tmp.go"
	if gotypeconverter.FlagStructTag != "" {
		ana.StructTag = gotypeconverter.FlagStructTag
	}

	pkgs, _ := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  path,
	}, "")
	got, err := gotypeconverter.Gen.Run(pkgs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
}
