package normal

import (
	"fmt"
)

type e struct {
	e bool
	m string
	x int
}

type a struct {
	e2 e
	x  struct {
		A int
		B bool
		C string `cvt:""`
		D string
	}
	y int
	z float64
	m string
}

type b struct {
	e
	x string
	y int
	z float64
}

func main() {
	fmt.Println("hello, world")
}
