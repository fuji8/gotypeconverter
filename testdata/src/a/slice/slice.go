package slice

type SRC struct {
	Ns  []float64
	N   float64
	Arr []struct {
		Foo string
		Bar int
	}
	Strs []string
	Sarr []struct {
		Foo string
		Bar int
	}
	Hellos []string
}

type DST struct {
	Ns  float64
	N   []float64
	Arr struct {
		Bar int
	}
	Strs []string
	Sarr []struct {
		Bar int
		Hii float32
	}
}

func main() {
}
