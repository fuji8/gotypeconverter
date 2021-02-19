package normal

type e struct {
	e       bool
	m       string
	x       int
	members []uint8
}

type ug struct {
	uaiueo uint8
	gaaaa  uint8
}

type normalSrc struct {
	x struct {
		A int
		B bool
		C string
		D string
	}
	y int
	z float64
	m string

	members []ug
}

type normalDst struct {
	e
	x string
	y int
	z float64
}

func main() {

}
