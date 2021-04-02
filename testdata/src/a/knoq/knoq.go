package knoq

import (
	"fmt"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/repository"
	"github.com/traPtitech/knoQ/router/service"
)

func foo() {
	fmt.Println(db.Event{}, repository.Event{}, domain.Event{}, service.EventRes{})
}
