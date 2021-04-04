package a

import (
	"a/basic"
	"a/cast"
	"a/external"
	"a/ignoretags"
	"a/named"
	"a/normal"
	"a/pointer"
	"a/samename"
	"a/samename/foo"
	"a/slice"
	"a/structtag"
	"fmt"

	"github.com/labstack/echo"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

// フィールド名が存在する場合は、
// 他のフィールドに依存しない結果が得られることを利用

type SRC struct {
	basic      basic.SRC
	external   []echo.Echo
	knoq       db.Event
	named      named.SRC
	normal     normal.SRC
	pointer    pointer.SRC
	samename   samename.Hoge
	slice      slice.SRC
	structtag  structtag.SRC
	cast       cast.Foo
	ignoretags ignoretags.SRC
}

type DST struct {
	basic      basic.DST
	external   external.DST
	knoq       domain.Event
	named      named.DST
	normal     normal.DST
	pointer    pointer.DST
	samename   foo.Hoge
	slice      slice.DST
	structtag  structtag.DST
	cast       cast.Bar
	ignoretags ignoretags.DST
}

type A struct {
	foo int
}

func castSample() {
	f := cast.Foo(0)
	b := cast.Bar(f)

	a := A{}
	var c struct {
		foo int
	}
	a = c

	fmt.Println(f, b, a, c)
}
