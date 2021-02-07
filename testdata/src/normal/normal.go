package normal

type e struct {
	e bool
	m string
	x int
}

type src struct {
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

type dst struct {
	e
	x string
	y int
	z float64
}

func main() {

}
