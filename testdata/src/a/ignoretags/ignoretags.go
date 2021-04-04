package ignoretags

type SRC struct {
	Foo int `json:"foo"`
}

type DST struct {
	Foo int `xml:"FOO"`
}
