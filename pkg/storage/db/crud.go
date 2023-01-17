package db

import (
	"context"
	"strconv"

	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/goqu/v9"
	"github.com/krixlion/goqu/v9/exp"
)

func (db DB) Get(ctx context.Context, id string) (entity.User, error) {
	query, args, _ := db.queryBuilder.From("user").Where(exp.Ex{"user.id": goqu.I(id)}).Prepared(true).ToSQL()
	user := entity.User{}
	if err := db.conn.GetContext(ctx, &user, query, args...); err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (db DB) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.User, error) {
	o, err := strconv.ParseUint(offset, 10, 32)
	if err != nil {
		return nil, err
	}
	l, err := strconv.ParseUint(limit, 10, 32)
	if err != nil {
		return nil, err
	}

	query, args, _ := db.queryBuilder.From("user").Limit(uint(l)).Offset(uint(o)).Prepared(true).ToSQL()

	users := []entity.User{}
	if err := db.conn.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, err
	}
	return users, nil
}

func (db DB) Create(ctx context.Context, user entity.User) error {
	hash, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash

	query, _, _ := db.queryBuilder.Insert("user").Rows(user).Prepared(true).ToSQL()

	if _, err = db.conn.NamedExecContext(ctx, query, user); err != nil {
		return err
	}

	return nil
}

func (db DB) Update(ctx context.Context, user entity.User) error {
	query, _, _ := db.queryBuilder.Update("user").Set(user).Where(exp.Ex{"user.id": goqu.I(user.Id)}).Prepared(true).ToSQL()

	if _, err := db.conn.NamedExecContext(ctx, query, user); err != nil {
		return err
	}
	return nil
}

func (db DB) Delete(ctx context.Context, id string) error {
	query, _, _ := db.queryBuilder.Delete("user").Where(exp.Ex{"user.id": goqu.I(id)}).Prepared(true).ToSQL()

	if _, err := db.conn.NamedExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}
