package multidb

import "github.com/jmoiron/sqlx"

type MultiDB interface {
	Get(driver, key string) (*sqlx.DB, error)
	CloseAll() error
}
