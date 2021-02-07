package main

type e struct {
	e bool
	m string
	x int
}

type normalSRC struct {
	x struct {
		A int
		B bool
		C string `cvt:""`
		D string
	}
	y int
	z float64
	m string
}

type normalDST struct {
	e
	x string
	y int
	z float64
}

func main() {
}
