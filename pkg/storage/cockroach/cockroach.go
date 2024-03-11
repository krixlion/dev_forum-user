package cockroach

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"go.nhat.io/otelsql"
	"go.opentelemetry.io/otel/trace"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const Driver = "postgres"

var _ storage.Storage = (*CockroachDB)(nil)

func formatConnString(host, port, user, password, dbname string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

type CockroachDB struct {
	conn         *sqlx.DB
	queryBuilder goqu.DialectWrapper
	tracer       trace.Tracer
}

func (db CockroachDB) Conn() *sql.DB {
	return db.conn.DB
}

func Make(host, port, user, password, dbName string, tracer trace.Tracer) (CockroachDB, error) {
	driverName, err := otelsql.Register(Driver,
		otelsql.AllowRoot(),
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.TracePing(),
	)
	if err != nil {
		return CockroachDB{}, err
	}

	db, err := sql.Open(driverName, formatConnString(host, port, user, password, dbName))
	if err != nil {
		return CockroachDB{}, err
	}
	queryBuilder := goqu.Dialect(Driver)

	return CockroachDB{
		conn:         sqlx.NewDb(db, Driver),
		queryBuilder: queryBuilder,
		tracer:       tracer,
	}, nil
}

func (db CockroachDB) Close() error {
	return db.conn.Close()
}
