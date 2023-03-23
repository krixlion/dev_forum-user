package server

import (
	"github.com/krixlion/dev_forum-user/pkg/entity"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
)

func userFromPB(v *pb.User) entity.User {
	return entity.User{
		Id:        v.GetId(),
		Name:      v.GetName(),
		Password:  v.GetPassword(),
		Email:     v.GetEmail(),
		CreatedAt: v.GetCreatedAt().AsTime(),
		UpdatedAt: v.GetUpdatedAt().AsTime(),
	}
}

func mapUserFields(s string) string {
	switch s {
	case "id":
		return "Id"
	case "name":
		return "Name"
	case "password":
		return "Password"
	case "email":
		return "Email"
	case "created_at":
		return "CreatedAt"
	case "updated_at":
		return "UpdatedAt"

	case "updated_at.seconds":
		return "UpdatedAt.Seconds"
	case "updated_at.nanos":
		return "UpdatedAt.Nanos"

	case "created_at.seconds":
		return "CreatedAt.Seconds"
	case "created_at.nanos":
		return "CreatedAt.Nanos"
	default:
		return s
	}
}
