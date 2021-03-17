package gotypeconverter

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gostaticanalysis/codegen/codegentest"
)

var flagUpdate bool

type flagValue struct {
	s string
	d string
	o string
}

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	m := map[string]flagValue{
		"external": flagValue{
			s: "[]echo.Echo",
			d: "externalDst",
		},
		"pointer": flagValue{
			s: "[]*pointerSrc",
			d: "[]*pointerDst",
		},
		"samename": flagValue{
			s: "Hoge",
			d: "foo.Hoge",
		},
		"knoq": flagValue{
			s: "db.Event",
			d: "domain.Event",
		},
	}

	fileInfos, err := ioutil.ReadDir(codegentest.TestData() + "/src")
	if err != nil {
		panic(err)
	}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			fv, ok := m[fi.Name()]
			if !ok {
				Generator.Flags.Set("s", fi.Name()+"Src")
				Generator.Flags.Set("d", fi.Name()+"Dst")
				Generator.Flags.Set("o", "")
			} else {
				Generator.Flags.Set("s", fv.s)
				Generator.Flags.Set("d", fv.d)
				Generator.Flags.Set("o", fv.o)
			}

			CreateTmpFile(codegentest.TestData() + "/src/" + fi.Name())

			rs := codegentest.Run(t, codegentest.TestData(), Generator, fi.Name())
			codegentest.Golden(t, rs, flagUpdate)
		}
	}
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
