package entity

import "github.com/krixlion/dev_forum-proto/user_service/pb"

type User struct {
	Id       string `db:"id,omitempty" goqu:"skipinsert,skipupdate,omitempty"`
	Name     string `db:"name" goqu:"omitempty"`
	Email    string `db:"email" goqu:"omitempty"`
	Password string `db:"password" goqu:"omitempty"`
}

func UserFromPB(v *pb.User) User {
	return User{
		Id:       v.GetId(),
		Name:     v.GetName(),
		Password: v.GetPassword(),
		Email:    v.GetEmail(),
	}
}
