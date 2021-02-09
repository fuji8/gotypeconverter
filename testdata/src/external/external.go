package external

import (
	"fmt"
	"sync"

	"github.com/labstack/echo"
)

type externalDst struct {
	Debug bool
	pool  sync.Pool
}

func foo() {
	e := echo.New()
	fmt.Println(e)
}
