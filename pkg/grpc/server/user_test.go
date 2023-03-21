package server

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_userFromPB(t *testing.T) {
	tests := []struct {
		desc string
		arg  *pb.User
		want entity.User
	}{
		{
			desc: "Test on random data",
			arg: &pb.User{
				Id:        "erofjigjbefkdw",
				Name:      "fwqavds",
				Email:     "ad@asda.pl",
				Password:  "poekmfvwes!234",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			},
			want: entity.User{
				Id:        "erofjigjbefkdw",
				Name:      "fwqavds",
				Email:     "ad@asda.pl",
				Password:  "poekmfvwes!234",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := userFromPB(tt.arg); !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Millisecond)) {
				t.Errorf("userFromPB() = %v, want %v", got, tt.want)
			}
		})
	}
}
