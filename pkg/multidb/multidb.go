package multidb

import (
	"github.com/jmoiron/sqlx"
	"io"
)

type MultiDB interface {
	GetSqlx(driver Driver, key string) (*sqlx.DB, error)
	io.Closer
}
