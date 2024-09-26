package cockroach

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/storage"
)

const usersTable = "users"

func (db CockroachDB) Get(ctx context.Context, params filter.Filter) (_ entity.User, err error) {
	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	exps, err := filterToSqlExp(params)
	if err != nil {
		return entity.User{}, err
	}

	query, args, err := db.queryBuilder.From(usersTable).Where(exps...).Prepared(true).ToSQL()
	if err != nil {
		return entity.User{}, err
	}

	var dataset userDataset
	if err := db.conn.GetContext(ctx, &dataset, query, args...); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, storage.ErrNotFound
		}
		return entity.User{}, err
	}

	return dataset.User()
}

func (db CockroachDB) GetMultiple(ctx context.Context, offset, limit string, params filter.Filter) (_ []entity.User, err error) {
	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	o, err := str.ConvertToUint(offset)
	if err != nil {
		return nil, err
	}

	l, err := str.ConvertToUint(limit)
	if err != nil {
		return nil, err
	}

	exps, err := filterToSqlExp(params)
	if err != nil {
		return nil, err
	}

	colExp := exp.NewColumnListExpression("name")
	orderExp := exp.NewOrderedExpression(colExp, exp.DescSortDir, exp.NullsLastSortType)
	mainExp := db.queryBuilder.From(usersTable).Order(orderExp).Limit(uint(l)).Offset(uint(o)).Where(exps...).Prepared(true)
	query, args, err := mainExp.ToSQL()
	if err != nil {
		return nil, err
	}

	datasets := []userDataset{}
	if err := crdb.Execute(func() error { return db.conn.SelectContext(ctx, &datasets, query, args...) }); err != nil {
		return nil, err
	}

	return usersFromDatasets(datasets)
}

func (db CockroachDB) Create(ctx context.Context, user entity.User) (err error) {
	ctx, span := db.tracer.Start(ctx, "db.Create")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	dataset := datasetFromUser(user)

	query, args, err := db.queryBuilder.Insert(usersTable).Rows(dataset).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	return crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, args...)
		return err
	})
}

func (db CockroachDB) Update(ctx context.Context, user entity.User) (err error) {
	ctx, span := db.tracer.Start(ctx, "db.Update")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	dataset := datasetFromUser(user)

	query, args, err := db.queryBuilder.Update(usersTable).Set(dataset).Where(goqu.C("id").Eq(dataset.Id)).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	return crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, args...)
		return err
	})
}

func (db CockroachDB) Delete(ctx context.Context, id string) (err error) {
	ctx, span := db.tracer.Start(ctx, "db.Delete")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	query, _, err := db.queryBuilder.Delete(usersTable).Where(goqu.C("id").Eq(id)).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	return crdb.Execute(func() error {
		_, err := db.conn.ExecContext(ctx, query, id)
		return err
	})
}
