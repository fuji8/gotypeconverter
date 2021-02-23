package structtag

type structtagSrc struct {
	foo int
	bar float64
	x   string
	y   string
}

type structtagDst struct {
	foo  int `cvt:"fooo"`
	bar  float32
	xxxx string `cvt:"x"`
	y    bool
}
