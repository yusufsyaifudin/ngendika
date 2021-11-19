package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
)

type SqlDbConnMaker struct {
	ctx    context.Context
	conf   config.Database
	dbSQL  map[string]*sqlx.DB
	closer []Closer
}

func NewSqlDbConnMaker(ctx context.Context, conf config.Database) (*SqlDbConnMaker, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	instance := &SqlDbConnMaker{
		ctx:    ctx,
		conf:   conf,
		dbSQL:  make(map[string]*sqlx.DB),
		closer: make([]Closer, 0),
	}

	err = instance.connect()
	if err != nil {
		// close previous opened connection if error happen
		if _err := instance.CloseAll(); _err != nil {
			err = fmt.Errorf("close db sql error: %w: %s", err, _err)
		}

		return nil, err
	}

	return instance, nil
}

func (i *SqlDbConnMaker) Get(key string) (*sqlx.DB, error) {
	v, ok := i.dbSQL[key]
	if !ok {
		return nil, fmt.Errorf("key %s is not exist on db list", key)
	}

	return v, nil
}

func (i *SqlDbConnMaker) CloseAll() error {
	ctx := i.ctx

	logger.Debug(ctx, "db sql: trying to close")

	var err error
	for _, closer := range i.closer {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		} else {
			logger.Debug(ctx, fmt.Sprintf("db sql: %s success to close", closer.Name()))
		}
	}

	if err != nil {
		logger.Error(ctx, "db sql: some error occurred when closing dep", logger.KV("error", err))
	}

	return err
}

func (i *SqlDbConnMaker) connect() error {
	// Preparing database connection SQL
	ctx := i.ctx

	for key, dbConfig := range i.conf {
		key = strings.TrimSpace(strings.ToLower(key))
		if err := validator.New().Var(key, "required,alphanumeric"); err != nil {
			err = fmt.Errorf("error connecting to database key '%s': %w", key, err)
			return err
		}

		switch dbConfig.Driver {
		case "mysql", "postgres":
			sqlxConn, err := sqlx.ConnectContext(ctx, dbConfig.Driver, dbConfig.DSN)
			if err != nil {
				err = fmt.Errorf("error connecting to database %s: %w", key, err)
				return err
			}

			// don't forget to register in closer, using unique name to track in the Log
			i.dbSQL[key] = sqlxConn
			i.closer = append(i.closer, NewNamedCloser(key, sqlxConn))
		}
	}

	return nil
}
