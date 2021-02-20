package named

type roomSrc struct {
	room  float64
	named []namedSrc
}

type roomDst struct {
	room  float64
	named []namedDst
}

type namedSrc struct {
	foo  int
	room roomSrc
}
type namedDst struct {
	foo  int
	room roomDst
}
