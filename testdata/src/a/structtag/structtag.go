package structtag

type SRC struct {
	Foo  int
	Bar  float64
	X    string
	Y    string
	Read int `cvt:"->"` // Read only
	Baz  int
}

type DST struct {
	Foo   int `cvt:"-"`
	Bar   float32
	Xxxx  string `cvt:"X"`
	Y     bool
	Read  int
	Write int `cvt:"Baz, <-"` // Write only
}
