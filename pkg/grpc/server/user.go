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
