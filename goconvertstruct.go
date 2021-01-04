package goconvertstruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"os"

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
	var buf bytes.Buffer

	var srcS, dstS *ast.TypeSpec
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				switch ts.Name.Name {
				case FlagSrc:
					srcS = ts
				case FlagDst:
					dstS = ts
				}
			}
			return true
		})
	}

	s, _ := srcS.Type.(*ast.StructType)
	fi := s.Fields.List[0]
	fmt.Println(pass.TypesInfo.TypeOf(fi.Type))
	fmt.Println(dstS)
	fmt.Fprintln(&buf, pass.TypesInfo.TypeOf(fi.Type))

	src, err := format.Source(buf.Bytes())
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
