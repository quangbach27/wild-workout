package adapter

import (
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewPostgresSQLConnection() (*sqlx.DB, error) {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DATABASE")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to PostgreSQL")
	}
	return db, nil
}

func MustNewPostgresSQLConnection() *sqlx.DB {
	db, err := NewPostgresSQLConnection()
	if err != nil {
		panic("connect to db failed.")
	}
	return db
}
