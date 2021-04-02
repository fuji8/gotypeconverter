package normal

type E struct {
	E       bool
	M       string
	X       int
	Members []uint8
}

type ug struct {
	Uaiueo uint8
	Gaaaa  uint8
}

type SRC struct {
	X struct {
		A int
		B bool
		C string
		D string
	}
	Y int
	Z float64
	M string

	Members []ug
}

type DST struct {
	E
	X string
	Y int
	Z float64
}

func main() {

}
