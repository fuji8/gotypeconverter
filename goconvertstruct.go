package goconvertstruct

import (
	"fmt"
	"go/format"
	"os"

	"github.com/fuji8/goconvertstruct/internal"
	"github.com/gostaticanalysis/codegen"
)

const doc = "goconvertstruct is ..."

var (
	flagOutput string

	FlagSrc, FlagDst string
)

func init() {
	Generator.Flags.StringVar(&flagOutput, "o", "", "output file name")
	Generator.Flags.StringVar(&FlagSrc, "s", "", "source struct")
	Generator.Flags.StringVar(&FlagDst, "d", "", "destination struct")
}

var Generator = &codegen.Generator{
	Name: "goconvertstruct",
	Doc:  doc,
	Run:  run,
}

func run(pass *codegen.Pass) error {
	g := new(internal.Generator)
	var data []byte

	for _, f := range pass.Files {
		g.Init(f)
		var err error
		data, err = g.Generate(FlagSrc, FlagDst)
		if err != nil {
			return err
		}

		// TODO
		break
	}

	src, err := format.Source(data)
	if err != nil {
		return err
	}

	if flagOutput == "" {
		pass.Print(string(src))
		return nil
	}

	f, err := os.Create(flagOutput)
	if err != nil {
		return err
	}

	fmt.Fprint(f, string(src))

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
