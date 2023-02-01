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
	CreatedAt time.Time `db:"created_at" goqu:"skipupdate,omitempty"`
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

// Add its own MarshalJSON implementation so that timestamps can be formatted as needed.
// func (u *User) MarshalJSON() ([]byte, error) {
// 	type Alias User
// 	return json.Marshal(&struct {
// 		*Alias
// 		CreatedAt string `json:"created_at,omitempty"`
// 		UpdatedAt string `json:"updated_at,omitempty"`
// 	}{
// 		Alias:     (*Alias)(u),
// 		CreatedAt: u.CreatedAt.Format(time.RFC3339),
// 		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
// 	})
// }
