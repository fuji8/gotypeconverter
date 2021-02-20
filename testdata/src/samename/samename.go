package samename

import (
	"fmt"
	"samename/bar"
	"samename/foo"
)

type Hoge struct {
	A int
	b string
}

func main() {
	fmt.Println(bar.Hoge{}, foo.Hoge{})
}
