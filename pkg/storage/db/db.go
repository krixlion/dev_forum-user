package db

import (
	"database/sql"
	"fmt"

	"github.com/krixlion/goqu/v9"
	_ "github.com/krixlion/goqu/v9/dialect/postgres"
	"go.nhat.io/otelsql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

const Driver = "postgres"

func formatConnString(host, port, user, password, dbname string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

type DB struct {
	conn         *sqlx.DB
	queryBuilder goqu.DialectWrapper
}

func (db DB) Conn() *sql.DB {
	return db.conn.DB
}

func Make(host, port, user, password, dbname string) (DB, error) {
	driverName, err := otelsql.Register(Driver,
		otelsql.AllowRoot(),
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.TracePing(),
	)
	if err != nil {
		return DB{}, err
	}

	db, err := sql.Open(driverName, formatConnString(host, port, user, password, dbname))
	if err != nil {
		return DB{}, err
	}
	queryBuilder := goqu.Dialect(Driver)

	return DB{
		conn:         sqlx.NewDb(db, Driver),
		queryBuilder: queryBuilder,
	}, nil
}

func (db DB) Close() error {
	return db.conn.Close()
}

func hashPassword(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
