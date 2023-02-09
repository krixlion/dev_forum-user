package db

import (
	"context"
	"strconv"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/goqu/v9"
	"github.com/krixlion/goqu/v9/exp"
)

const usersTable = "users"

func (db DB) Get(ctx context.Context, id string) (entity.User, error) {
	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()

	query, args, err := db.queryBuilder.From(usersTable).Where(exp.Ex{usersTable + ".id": id}).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.User{}, err
	}

	var dataset userDataset
	if err := db.conn.GetContext(ctx, &dataset, query, args...); err != nil {
		return entity.User{}, err
	}

	return userFromDataset(dataset), nil
}

func (db DB) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.User, error) {
	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
	defer span.End()

	var o uint64
	var l uint64
	var err error

	if offset != "" {
		o, err = strconv.ParseUint(offset, 10, 32)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return nil, err
		}
	}

	if limit != "" {
		l, err = strconv.ParseUint(limit, 10, 32)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return nil, err
		}
	}

	query, args, err := db.queryBuilder.From(usersTable).Limit(uint(l)).Offset(uint(o)).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	datasets := []userDataset{}
	err = crdb.Execute(func() error {
		return db.conn.SelectContext(ctx, &datasets, query, args...)
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}
	return usersFromDatasets(datasets), nil
}

func (db DB) Create(ctx context.Context, user entity.User) error {
	ctx, span := db.tracer.Start(ctx, "db.Create")
	defer span.End()

	dataset := datasetFromUser(user)

	query, args, err := db.queryBuilder.Insert(usersTable).Rows(dataset).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	err = crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, args...)
		return err
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) Update(ctx context.Context, user entity.User) error {
	ctx, span := db.tracer.Start(ctx, "db.Update")
	defer span.End()

	dataset := datasetFromUser(user)

	query, args, err := db.queryBuilder.Update(usersTable).Set(dataset).Where(goqu.C("id").Eq(dataset.Id)).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	err = crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, args...)
		return err
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := db.tracer.Start(ctx, "db.Delete")
	defer span.End()

	query, _, err := db.queryBuilder.Delete(usersTable).Where(goqu.C("id").Eq(id)).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	err = crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, id)
		return err
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}
