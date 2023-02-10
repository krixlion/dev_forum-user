package db

import (
	"time"

	"github.com/krixlion/dev_forum-user/pkg/entity"
)

type userDataset struct {
	Id        string `db:"id" goqu:"skipupdate,omitempty"`
	Name      string `db:"name" goqu:"omitempty"`
	Email     string `db:"email" goqu:"omitempty"`
	Password  string `db:"password" goqu:"omitempty"`
	CreatedAt string `db:"created_at" goqu:"skipupdate,omitempty"`
	UpdatedAt string `db:"updated_at" goqu:"omitempty"`
}

func datasetFromUser(v entity.User) userDataset {
	return userDataset{
		Id:        v.Id,
		Name:      v.Name,
		Password:  v.Password,
		Email:     v.Email,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
		UpdatedAt: v.UpdatedAt.Format(time.RFC3339),
	}

}

func userFromDataset(v userDataset) (entity.User, error) {
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

func usersFromDatasets(vs []userDataset) ([]entity.User, error) {
	users := []entity.User{}
	for _, v := range vs {
		createdAt, err := time.Parse(time.RFC3339, v.CreatedAt)
		if err != nil {
			return nil, err
		}

		updatedAt, err := time.Parse(time.RFC3339, v.UpdatedAt)
		if err != nil {
			return nil, err
		}

		user := entity.User{
			Id:        v.Id,
			Name:      v.Name,
			Password:  v.Password,
			Email:     v.Email,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		users = append(users, user)
	}
	return users, nil
}
