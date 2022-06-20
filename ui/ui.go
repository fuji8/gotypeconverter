package ui

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"sort"

	ana "github.com/fuji8/gotypeconverter/analysis"
	"golang.org/x/tools/imports"
)

// TmpFilePath is output tmp file
var TmpFilePath = "./generated.go"

func sortFunction(data []byte, fileName string) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fileName, data, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// sort function
	sort.Slice(file.Decls, func(i, j int) bool {
		fdi, iok := file.Decls[i].(*ast.FuncDecl)
		if !iok {
			return true
		}
		fdj, jok := file.Decls[j].(*ast.FuncDecl)
		if !jok {
			return false
		}
		return fdi.Name.Name < fdj.Name.Name
	})

	dst := new(bytes.Buffer)
	err = format.Node(dst, fset, file)
	return dst.Bytes(), err
}

func NoInfoGeneration(fm *ana.FuncMaker) (string, error) {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "// Code generated by gotypeconverter; DO NOT EDIT.\n")
	fmt.Fprintf(buf, "package %s\n", fm.Pkg().Name())

	buf.Write(fm.WriteBytes())

	fmt.Println(buf.String())

	sortedData, err := sortFunction(buf.Bytes(), TmpFilePath)
	if err != nil {
		return "", err
	}

	importedData, err := imports.Process(TmpFilePath, sortedData, &imports.Options{
		Fragment: true,
		Comments: true,
	})
	if err != nil {
		return "", err
	}
	return string(importedData), nil
}

// FileNameGeneration 新規の関数を追加、同名の関数を置き換え、既存の関数は変更せず、
// ソートした結果を返します。
func FileNameGeneration(fm *ana.FuncMaker, outputFilename string) (string, error) {
	fileData, err := ioutil.ReadFile(outputFilename)
	if err != nil {
		return NoInfoGeneration(fm)
	}

	output := append(fileData, fm.WriteBytes()...)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, outputFilename, output, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// delete same name func
	funcDeclMap := make(map[string]*ast.FuncDecl)
	for _, d := range file.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			funcDeclMap[fd.Name.Name] = fd
		}
	}
	newDecls := make([]ast.Decl, 0)
	for _, d := range file.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			if _, ok := funcDeclMap[fd.Name.Name]; ok {
				continue
			}
		}

		newDecls = append(newDecls, d)
	}
	for _, lastFd := range funcDeclMap {
		newDecls = append(newDecls, lastFd)
	}
	file.Decls = newDecls

	// sort function
	sort.Slice(file.Decls, func(i, j int) bool {
		fdi, iok := file.Decls[i].(*ast.FuncDecl)
		if !iok {
			return true
		}
		fdj, jok := file.Decls[j].(*ast.FuncDecl)
		if !jok {
			return false
		}
		return fdi.Name.Name < fdj.Name.Name
	})

	dst := new(bytes.Buffer)
	err = format.Node(dst, fset, file)
	if err != nil {
		return "", err
	}

	sortedData := dst.Bytes()
	importedData, err := imports.Process(outputFilename, sortedData, &imports.Options{
		Fragment: true,
		Comments: true,
	})

	if err != nil {
		return "", err
	}
	return string(importedData), nil
}
