package external

import (
	"fmt"
	"sync"

	"github.com/labstack/echo"
)

type DST struct {
	Debug bool
	pool  sync.Pool
}

func foo() {
	e := echo.New()
	fmt.Println(e)
}
