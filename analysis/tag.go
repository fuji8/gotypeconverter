package analysis

import (
	"strings"

	"github.com/fatih/structtag"
)

var StructTag = "cvt"

type OptionTag int

const (
	Ignore OptionTag = iota + 1
	ReadOnly
	WriteOnly
)

func getTag(tag string) (name, readName, writeName string, option OptionTag) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return
	}
	cvtTag, err := tags.Get(StructTag)
	if err != nil {
		return
	}

	for _, tag := range append(cvtTag.Options, cvtTag.Name) {
		tag = strings.Trim(tag, " ")

		if strings.HasPrefix(tag, "read:") {
			readName = tag[5:]
			continue
		}
		if strings.HasPrefix(tag, "write:") {
			writeName = tag[6:]
			continue
		}

		switch tag {
		case "-":
			option = Ignore
		case "->":
			option = ReadOnly
		case "<-":
			option = WriteOnly
		default:
			name = tag
		}
	}
	return
}
