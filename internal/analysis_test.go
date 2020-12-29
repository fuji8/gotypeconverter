package internal

import (
	"bytes"
	"go/ast"
	"testing"
)

const testDataDir = "../testdata"

func Test_print(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				filename: testDataDir + "/normal.go",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := print(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	type fields struct {
		top ast.Node
		buf bytes.Buffer
		src *ast.TypeSpec
		dst *ast.TypeSpec
	}
	type args struct {
		srcStructName string
		dstStructName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				srcStructName: "a",
				dstStructName: "b",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				top: tt.fields.top,
				buf: tt.fields.buf,
				src: tt.fields.src,
				dst: tt.fields.dst,
			}
			g.Init(testDataDir + "/normal.go")
			if err := g.Generate(tt.args.srcStructName, tt.args.dstStructName); (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
