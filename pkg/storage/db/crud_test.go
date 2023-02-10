package db_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-user/pkg/storage/db"
)

func setUpDB() db.DB {
	env.Load("app")

	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	storage, err := db.Make(db_host, db_port, db_user, db_pass, db_name, nulls.NullTracer{})
	if err != nil {
		panic(err)
	}
	return storage
}

func Test_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration db.Get test.")
	}
	tests := []struct {
		name    string
		id      string
		want    entity.User
		wantErr bool
	}{
		{
			name: "Test on simple data",
			id:   "test",
			want: entity.User{
				Id:        "test",
				Name:      "testName",
				Email:     "test@test.test",
				Password:  "testPass",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			got, err := db.Get(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Minute)) {
				t.Errorf("DB.Get():\n got = %v\n want = %v\n %v\n", got, tt.want, cmp.Diff(got, tt.want))
				return
			}
		})
	}
}

func Test_GetMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.GetMultiple integration test.")
	}

	user1 := entity.User{
		Id:        "1",
		Name:      "name-1",
		Email:     "email-1",
		Password:  "pass-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := entity.User{
		Id:        "2",
		Name:      "name-2",
		Email:     "email-2",
		Password:  "pass-2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user3 := entity.User{
		Id:        "3",
		Name:      "name-3",
		Email:     "email-3",
		Password:  "pass-3",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	type args struct {
		offset string
		limit  string
	}
	tests := []struct {
		name    string
		args    args
		want    []entity.User
		wantErr bool
	}{
		{
			name: "Test on simple data",
			args: args{
				offset: "0",
				limit:  "3",
			},
			want: []entity.User{user1, user2, user3},
		},
		{
			name: "Test if correctly applies offset on simple data",
			args: args{
				offset: "1",
				limit:  "2",
			},
			want: []entity.User{user2, user3},
		},
		{
			name: "Test if correctly applies limit",
			args: args{
				offset: "0",
				limit:  "2",
			},
			want: []entity.User{user1, user2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			got, err := db.GetMultiple(ctx, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetMultiple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Minute)) {
				t.Errorf("DB.GetMultiple():\n got = %v\n want = %v\n %v\n", got, tt.want, cmp.Diff(got, tt.want))
				return
			}
		})
	}
}

func Test_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Create integration test.")
	}
	tests := []struct {
		name    string
		user    entity.User
		wantErr bool
	}{
		{
			name: "Test if correctly creates a random user",
			user: func() entity.User {
				v := gentest.RandomUser(3, 5, 5)
				v.Id = "9999999"
				return v
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			if err := db.Create(ctx, tt.user); (err != nil) != tt.wantErr {
				t.Errorf("DB.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want := tt.user
			got, err := db.Get(ctx, want.Id)
			if err != nil {
				t.Errorf("Failed to DB.Get() after DB.Create() error = %v", err)
				return
			}

			if !cmp.Equal(got, want, cmpopts.EquateApproxTime(time.Second)) {
				t.Errorf("DB.Create():\n got = %v\n want = %v\n %v\n", got, want, cmp.Diff(got, want))
				return
			}
		})
	}
}

func Test_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Update integration test.")
	}
	tests := []struct {
		name    string
		user    entity.User
		wantErr bool
	}{
		{
			name: "Test if correctly updates a user",
			user: func() entity.User {
				user := gentest.RandomUser(2, 2, 2)
				user.Id = "test"
				return user
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			if err := db.Update(ctx, tt.user); (err != nil) != tt.wantErr {
				t.Errorf("DB.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want := tt.user
			got, err := db.Get(ctx, want.Id)
			if err != nil {
				t.Errorf("Failed to DB.Get() after DB.Update() error = %v", err)
				return
			}

			if !cmp.Equal(got, want, cmpopts.EquateApproxTime(time.Second*5)) {
				t.Errorf("DB.Update():\n got = %v\n want = %v\n %v\n", got, want, cmp.Diff(got, want))
				return
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping db.Delete integration test.")
	}
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Test if correctly deletes a simple user",
			id:      "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			db := setUpDB()

			if err := db.Delete(ctx, tt.id); (err != nil) != tt.wantErr {
				t.Errorf("DB.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err := db.Get(ctx, tt.id)
			if !errors.Is(err, sql.ErrNoRows) {
				t.Errorf("DB.Create():\n gotErr = %T, wantErr = %T, err = %v", err, sql.ErrNoRows, err)
				return
			}
		})
	}
}
