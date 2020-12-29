package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func print(filename string) error {
	fset := token.NewFileSet()

	var node ast.Node
	node, err := parser.ParseFile(fset, filename, nil, parser.Mode(0))

	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if x, ok := n.(*ast.TypeSpec); ok {
			ast.Print(fset, x)
			fmt.Println() // \n したい...
		}
		return true
	})

	return nil
}
