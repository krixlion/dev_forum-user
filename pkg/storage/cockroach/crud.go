package cockroach

import (
	"context"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-user/pkg/entity"
)

const usersTable = "users"

func (db CockroachDB) Get(ctx context.Context, params filter.Filter) (entity.User, error) {
	ctx, span := db.tracer.Start(ctx, "db.Get")
	defer span.End()

	exps, err := filterToSqlExp(params)
	if err != nil {
		return entity.User{}, err
	}

	query, args, err := db.queryBuilder.From(usersTable).Where(exps...).Prepared(true).ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.User{}, err
	}

	var dataset userDataset
	if err := db.conn.GetContext(ctx, &dataset, query, args...); err != nil {
		return entity.User{}, err
	}

	user, err := dataset.User()
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (db CockroachDB) GetMultiple(ctx context.Context, offset, limit string, params filter.Filter) ([]entity.User, error) {
	ctx, span := db.tracer.Start(ctx, "db.GetMultiple")
	defer span.End()

	o, err := str.ConvertToUint(offset)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	l, err := str.ConvertToUint(limit)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	exps, err := filterToSqlExp(params)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	colExp := exp.NewColumnListExpression("name")
	orderExp := exp.NewOrderedExpression(colExp, exp.DescSortDir, exp.NullsLastSortType)
	mainExp := db.queryBuilder.From(usersTable).Order(orderExp).Limit(uint(l)).Offset(uint(o)).Where(exps...).Prepared(true)
	query, args, err := mainExp.ToSQL()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	datasets := []userDataset{}
	if err := crdb.Execute(func() error { return db.conn.SelectContext(ctx, &datasets, query, args...) }); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	users, err := usersFromDatasets(datasets)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	return users, nil
}

func (db CockroachDB) Create(ctx context.Context, user entity.User) error {
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

func (db CockroachDB) Update(ctx context.Context, user entity.User) error {
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

func (db CockroachDB) Delete(ctx context.Context, id string) error {
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
