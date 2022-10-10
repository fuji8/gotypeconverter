package foo

type Hoge struct {
	A int
	b string
}

type Huga struct {
	C int
	Hoge
}

type DST struct {
	Huga
}
