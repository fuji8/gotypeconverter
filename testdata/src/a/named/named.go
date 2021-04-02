package named

type roomSrc struct {
	room  float64
	named []SRC
}

type roomDst struct {
	room  float64
	named []DST
}

type SRC struct {
	foo  int
	room roomSrc
}
type DST struct {
	foo  int
	room roomDst
}
