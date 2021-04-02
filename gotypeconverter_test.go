package gotypeconverter

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/gostaticanalysis/codegen/codegentest"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	Generator.Flags.Set("s", "SRC")
	Generator.Flags.Set("d", "DST")

	CreateTmpFile(codegentest.TestData() + "/src/a")
	rs := codegentest.Run(t, codegentest.TestData(), Generator, "a")
	codegentest.Golden(t, rs, true)
}

func Test_getTag(t *testing.T) {
	flagStructTag = "cvt"
	templ := "cvt:\"%s\""
	type args struct {
		tag string
	}
	tests := []struct {
		name          string
		args          args
		wantName      string
		wantReadName  string
		wantWriteName string
		wantOption    optionTag
	}{
		{
			name: "read, write",
			args: args{
				tag: fmt.Sprintf(templ, "read:foo, write:bar"),
			},
			wantName:      "",
			wantReadName:  "foo",
			wantWriteName: "bar",
			wantOption:    optionTag(0),
		},
		{
			name: "fix Name",
			args: args{
				tag: fmt.Sprintf(templ, "Foo, write:Baz, -"),
			},
			wantName:      "Foo",
			wantReadName:  "",
			wantWriteName: "Baz",
			wantOption:    ignore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotReadName, gotWriteName, gotOption := getTag(tt.args.tag)
			if gotName != tt.wantName {
				t.Errorf("getTag() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotReadName != tt.wantReadName {
				t.Errorf("getTag() gotReadName = %v, want %v", gotReadName, tt.wantReadName)
			}
			if gotWriteName != tt.wantWriteName {
				t.Errorf("getTag() gotWriteName = %v, want %v", gotWriteName, tt.wantWriteName)
			}
			if gotOption != tt.wantOption {
				t.Errorf("getTag() gotOption = %v, want %v", gotOption, tt.wantOption)
			}
		})
	}
}
