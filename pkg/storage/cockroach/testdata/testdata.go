package testdata

import (
	"strconv"
	"time"

	"github.com/krixlion/dev_forum-user/pkg/entity"
)

var Users = initUsers()

func initUsers() map[string]entity.User {
	count := 3
	users := make(map[string]entity.User, count)

	for i := 1; i <= count; i++ {
		id := strconv.Itoa(i)
		users[id] = entity.User{
			Id:        id,
			Name:      "name-" + id,
			Email:     "email-" + id,
			Password:  "pass-" + id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return users
}
