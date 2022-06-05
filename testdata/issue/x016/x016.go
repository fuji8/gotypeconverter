package x016

type SRC struct {
	Foo int
	Bar float64
	X   string
	Y   string
}

type DST struct {
	Foo int
	Bar float32
	X   string
	Y   bool
}

func basic() {
	var src SRC
	var dst DST
	dst.Foo = src.Bar
}
