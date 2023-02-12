package db

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-user/pkg/entity"
)

func Test_datasetFromUser(t *testing.T) {
	tests := []struct {
		name string
		arg  entity.User
		want sqlUser
	}{
		{
			name: "Test on simple user",
			arg: entity.User{
				Id:        "test",
				Name:      "testname",
				Email:     "test@test.test",
				Password:  "testpass",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			want: sqlUser{
				Id:        "test",
				Name:      "testname",
				Email:     "test@test.test",
				Password:  "testpass",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := datasetFromUser(tt.arg); !cmp.Equal(got, tt.want) {
				t.Errorf("datasetFromUser():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_userFromDataset(t *testing.T) {
	tests := []struct {
		name    string
		arg     sqlUser
		want    entity.User
		wantErr bool
	}{
		{
			name: "Test on simple user",
			arg: sqlUser{
				Id:        "test",
				Name:      "testname",
				Email:     "test@test.test",
				Password:  "testpass",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
			want: entity.User{
				Id:        "test",
				Name:      "testname",
				Email:     "test@test.test",
				Password:  "testpass",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Test if returns an error on invalid time",
			arg: sqlUser{
				Id:        "test",
				Name:      "testname",
				Email:     "test@test.test",
				Password:  "testpass",
				CreatedAt: "invalid time",
				UpdatedAt: "invalid time",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.arg.User()
			if (err != nil) != tt.wantErr {
				t.Errorf("userFromDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second)) {
				t.Errorf("userFromDataset():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_usersFromDatasets(t *testing.T) {
	tests := []struct {
		name    string
		arg     []sqlUser
		want    []entity.User
		wantErr bool
	}{
		{
			want: []entity.User{
				{
					Id:        "test",
					Name:      "testname",
					Email:     "test@test.test",
					Password:  "testpass",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			arg: []sqlUser{
				{
					Id:        "test",
					Name:      "testname",
					Email:     "test@test.test",
					Password:  "testpass",
					CreatedAt: time.Now().Format(time.RFC3339),
					UpdatedAt: time.Now().Format(time.RFC3339),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := usersFromDatasets(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("usersFromDatasets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second)) {
				t.Errorf("usersFromDatasets():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
