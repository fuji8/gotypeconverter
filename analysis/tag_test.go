package analysis

import (
	"fmt"
	"testing"
)

func Test_getTag(t *testing.T) {
	StructTag = "cvt"
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
		wantOption    OptionTag
	}{
		{
			name: "read, write",
			args: args{
				tag: fmt.Sprintf(templ, "read:foo, write:bar"),
			},
			wantName:      "",
			wantReadName:  "foo",
			wantWriteName: "bar",
			wantOption:    OptionTag(0),
		},
		{
			name: "fix Name",
			args: args{
				tag: fmt.Sprintf(templ, "Foo, write:Baz, -"),
			},
			wantName:      "Foo",
			wantReadName:  "",
			wantWriteName: "Baz",
			wantOption:    Ignore,
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
