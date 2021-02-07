package main

type e struct {
	e bool
	m string
	x int
}

type a struct {
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

type b struct {
	e
	x string
	y int
	z float64
}

func main() {
	//src := a{
	//e2: e{
	//e: false,
	//m: "fooo",
	//x: 100,
	//},
	//y: 99,
	//z: 3.14,
	//m: "emmmmm",
	//}
	//src.x.C = "ccccc"
	//fmt.Println(ConvertaTob(src))
}
