package normal

import (
	"fmt"

	"github.com/gostaticanalysis/codegen"
)

type a struct {
	x codegen.Generator
	y int
	z float64
}

type b struct {
	x string
	y int
	z float64
}

func main() {
	fmt.Println("hello, world")
}
