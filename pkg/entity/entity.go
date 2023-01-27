package entity

import (
	"time"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
)

type User struct {
	Id        string    `db:"id" goqu:"skipupdate,omitempty"`
	Name      string    `db:"name" goqu:"omitempty"`
	Email     string    `db:"email" goqu:"omitempty"`
	Password  string    `db:"password" goqu:"omitempty"`
	CreatedAt time.Time `db:"created_at" goqu:"omitempty"`
	UpdatedAt time.Time `db:"updated_at" goqu:"omitempty"`
}

func UserFromPB(v *pb.User) User {
	return User{
		Id:        v.GetId(),
		Name:      v.GetName(),
		Password:  v.GetPassword(),
		Email:     v.GetEmail(),
		CreatedAt: v.GetCreatedAt().AsTime(),
		UpdatedAt: v.GetUpdatedAt().AsTime(),
	}
}
