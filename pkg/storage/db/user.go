package db

import (
	"time"

	"github.com/krixlion/dev_forum-user/pkg/entity"
)

type userDataset struct {
	Id        string    `db:"id" goqu:"skipupdate,omitempty"`
	Name      string    `db:"name" goqu:"omitempty"`
	Email     string    `db:"email" goqu:"omitempty"`
	Password  string    `db:"password" goqu:"omitempty"`
	CreatedAt time.Time `db:"created_at" goqu:"skipupdate,omitempty"`
	UpdatedAt time.Time `db:"updated_at" goqu:"omitempty"`
}

func datasetFromUser(v entity.User) userDataset {
	return userDataset{
		Id:        v.Id,
		Name:      v.Name,
		Password:  v.Password,
		Email:     v.Email,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}

}

func userFromDataset(v userDataset) entity.User {
	return entity.User{
		Id:        v.Id,
		Name:      v.Name,
		Password:  v.Password,
		Email:     v.Email,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func usersFromDatasets(vs []userDataset) []entity.User {
	users := []entity.User{}
	for _, v := range vs {
		user := entity.User{
			Id:        v.Id,
			Name:      v.Name,
			Password:  v.Password,
			Email:     v.Email,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		users = append(users, user)
	}
	return users
}
