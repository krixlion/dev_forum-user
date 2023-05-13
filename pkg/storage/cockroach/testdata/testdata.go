package testdata

import (
	"strconv"
	"time"

	"github.com/krixlion/dev_forum-user/pkg/entity"
)

var (
	Users map[string]entity.User
)

func initTestData() error {
	count := 3
	Users = make(map[string]entity.User, count)

	for i := 1; i <= count; i++ {
		id := strconv.Itoa(i)
		Users[id] = entity.User{
			Id:        id,
			Name:      "name-" + id,
			Email:     "email-" + id,
			Password:  "pass-" + id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return nil
}
