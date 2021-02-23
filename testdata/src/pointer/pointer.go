package pointer

import (
	"time"
)

type sqlNullTine struct {
	Time time.Time
}
type deleteAt sqlNullTine

type srcModel struct {
	createdAt time.Time
	deleteAt  deleteAt
}

type pointerSrc struct {
	id    int
	room  string
	group string
	srcModel
}

type dstRoomGroup struct {
	room  string
	group string
}

type dstModel struct {
	createdAt time.Time
	deleteAt  *time.Time
}

type pointerDst struct {
	id int
	dstRoomGroup
	dstModel
}
