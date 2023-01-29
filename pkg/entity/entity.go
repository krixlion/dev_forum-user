package entity

import (
	"time"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
)

type User struct {
	Id        string    `db:"id" goqu:"skipupdate,omitempty" json:"id,omitempty"`
	Name      string    `db:"name" goqu:"omitempty" json:"name,omitempty"`
	Email     string    `db:"email" goqu:"omitempty" json:"email,omitempty"`
	Password  string    `db:"password" goqu:"omitempty" json:"password,omitempty"`
	CreatedAt time.Time `db:"created_at" goqu:"skipupdate,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at" goqu:"omitempty" json:"updated_at,omitempty"`
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

// Add its own MarshalJSON implementation so that time.Time can be formated as needed.
// func (u *User) MarshalJSON() ([]byte, error) {
// 	type Alias User
// 	return json.Marshal(&struct {
// 		*Alias
// 		CreatedAt string `json:"created_at,omitempty"`
// 		UpdatedAt string `json:"updated_at,omitempty"`
// 	}{
// 		Alias: (*Alias)(u),
// 		CreatedAt: u.CreatedAt.Format("2017-01-15T01:30:15.01Z"),
// 		UpdatedAt: u.UpdatedAt.Format("2017-01-15T01:30:15.01Z"),
// 	})
// }
