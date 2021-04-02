package named

type roomSrc struct {
	Room  float64
	Named []SRC
}

type roomDst struct {
	Room  float64
	Named []DST
}

type SRC struct {
	Foo  int
	Room roomSrc
}
type DST struct {
	Foo  int
	Room roomDst
}
