package db

import (
	"time"

	"github.com/krixlion/dev_forum-user/pkg/entity"
)

type sqlUser struct {
	Id        string `db:"id" goqu:"skipupdate,omitempty"`
	Name      string `db:"name" goqu:"omitempty"`
	Email     string `db:"email" goqu:"omitempty"`
	Password  string `db:"password" goqu:"omitempty"`
	CreatedAt string `db:"created_at" goqu:"skipupdate,omitempty"`
	UpdatedAt string `db:"updated_at" goqu:"omitempty"`
}

func datasetFromUser(v entity.User) sqlUser {
	return sqlUser{
		Id:        v.Id,
		Name:      v.Name,
		Password:  v.Password,
		Email:     v.Email,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
		UpdatedAt: v.UpdatedAt.Format(time.RFC3339),
	}
}

func (v sqlUser) User() (entity.User, error) {
	createdAt, err := time.Parse(time.RFC3339, v.CreatedAt)
	if err != nil {
		return entity.User{}, err
	}

	updatedAt, err := time.Parse(time.RFC3339, v.UpdatedAt)
	if err != nil {
		return entity.User{}, err
	}

	return entity.User{
		Id:        v.Id,
		Name:      v.Name,
		Password:  v.Password,
		Email:     v.Email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func usersFromDatasets(vs []sqlUser) ([]entity.User, error) {
	users := []entity.User{}
	for _, v := range vs {
		user, err := v.User()
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
