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

type Huga struct {
	C int
}

type SRC struct {
	Huga
}

func main() {
	fmt.Println(bar.Hoge{}, foo.Hoge{})
}
