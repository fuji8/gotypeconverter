package structtag

type structtagSrc struct {
	foo  int
	bar  float64
	x    string
	y    string
	read int `cvt:"->"` // read only
	baz  int
}

type structtagDst struct {
	foo   int `cvt:"-"`
	bar   float32
	xxxx  string `cvt:"x"`
	y     bool
	read  int
	write int `cvt:"baz, <-"` // write only
}
