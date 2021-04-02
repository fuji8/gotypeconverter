package pointer

import (
	"time"
)

type sqlNullTine struct {
	Time time.Time
}
type DeleteAt sqlNullTine

type SrcModel struct {
	CreatedAt time.Time
	DeleteAt  DeleteAt
}

type SRC struct {
	ID    int
	Room  string
	Group string
	SrcModel
}

type DstRoomGroup struct {
	Room  string
	Group string
}

type DstModel struct {
	CreatedAt time.Time
	DeleteAt  *time.Time
}

type DST struct {
	ID int
	DstRoomGroup
	DstModel
}

//func main() {
//src := []*pointerSrc{
//{
//SrcModel: SrcModel{
//CreatedAt: time.Now(),
//DeleteAt: DeleteAt{
//Time: time.Now(),
//},
//},
//},
//}
//pointerDst := ConvertSlicePointerpointerSrcToSlicePointerpointerDst(src)
//fmt.Println(pointerDst[0].DeleteAt)
//}
