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
	pointer    *pointer.SRC
	samename   samename.Hoge
	samename2  samename.SRC
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
	pointer    *pointer.DST
	samename   foo.Hoge
	samename2  foo.DST
	slice      slice.DST
	structtag  structtag.DST
	cast       cast.Bar
	ignoretags ignoretags.DST
}
