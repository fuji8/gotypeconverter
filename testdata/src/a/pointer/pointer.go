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

type SRC struct {
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

type DST struct {
	id int
	dstRoomGroup
	dstModel
}

//func main() {
//src := []*pointerSrc{
//{
//srcModel: srcModel{
//createdAt: time.Now(),
//deleteAt: deleteAt{
//Time: time.Now(),
//},
//},
//},
//}
//pointerDst := ConvertSlicePointerpointerSrcToSlicePointerpointerDst(src)
//fmt.Println(pointerDst[0].deleteAt)
//}
