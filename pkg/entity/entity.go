package entity

import (
	"time"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
)

type User struct {
	Id        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
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
