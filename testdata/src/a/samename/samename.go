package samename

import (
	"a/samename/bar"
	"a/samename/foo"
	"fmt"
)

type Hoge struct {
	A int
	b string
}

func main() {
	fmt.Println(bar.Hoge{}, foo.Hoge{})
}
