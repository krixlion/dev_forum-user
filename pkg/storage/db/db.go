package db

import (
	"fmt"

	"github.com/krixlion/goqu/v9"
	_ "github.com/krixlion/goqu/v9/dialect/postgres"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func formatConnString(host, port, user, password, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

type DB struct {
	conn         *sqlx.DB
	queryBuilder goqu.DialectWrapper
}

func Make(host, port, user, password, dbname string) (DB, error) {
	conn, err := sqlx.Open("postgres", formatConnString(host, port, user, password, dbname))
	if err != nil {
		return DB{}, err
	}
	queryBuilder := goqu.Dialect("postgres")

	return DB{
		conn:         conn,
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
