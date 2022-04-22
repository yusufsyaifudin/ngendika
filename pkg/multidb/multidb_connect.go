package multidb

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ylog"
)

type SqlDbConnMakerConfig struct {
	Ctx    context.Context `validate:"required"`
	Config Database        `validate:"required"`
}

type SqlDbConnMaker struct {
	ctx      context.Context
	conf     Database
	disabled map[string]struct{} // list of disabled databases, using struct for minimal memory footprint
	dbSQL    map[string]*sqlx.DB // db key name => real connection
	dbDriver map[string]string   // db key name => driver name
	closer   []Closer
}

var _ MultiDB = (*SqlDbConnMaker)(nil)

func NewSqlDbConnMaker(conf SqlDbConnMakerConfig) (*SqlDbConnMaker, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		err = fmt.Errorf("sql db connection maker failed: %w", err)
		return nil, err
	}

	instance := &SqlDbConnMaker{
		ctx:      conf.Ctx,
		conf:     conf.Config,
		disabled: make(map[string]struct{}),
		dbSQL:    make(map[string]*sqlx.DB),
		dbDriver: make(map[string]string),
		closer:   make([]Closer, 0),
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

func (i *SqlDbConnMaker) Get(driver, key string) (*sqlx.DB, error) {
	_, exists := i.disabled[key]
	if exists {
		return nil, fmt.Errorf("db with key '%s' is disabled", key)
	}

	v, ok := i.dbSQL[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' is not exist on db list", key)
	}

	registeredDriver, ok := i.dbDriver[key]
	if ok && driver == registeredDriver {
		return v, nil
	}

	return nil, fmt.Errorf("db key '%s' not using driver %s", key, driver)
}

func (i *SqlDbConnMaker) CloseAll() error {
	ctx := i.ctx

	ylog.Debug(ctx, "db sql: trying to close")

	var err error
	for _, closer := range i.closer {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		} else {
			ylog.Debug(ctx, fmt.Sprintf("db sql: %s success to close", closer.Name()))
		}
	}

	if err != nil {
		ylog.Error(ctx, "db sql: some error occurred when closing dep", ylog.KV("error", err))
	}

	return err
}

func (i *SqlDbConnMaker) connect() error {
	// Preparing database connection SQL
	ctx := i.ctx

	for key, dbConfig := range i.conf {
		key = strings.TrimSpace(strings.ToLower(key))
		if err := validator.New().Var(key, "required,alphanum"); err != nil {
			err = fmt.Errorf("error connecting to database key '%s': %w", key, err)
			return err
		}

		if dbConfig.Disable {
			i.disabled[key] = struct{}{}
			continue
		}

		dbConfig.Driver = strings.ToLower(strings.TrimSpace(dbConfig.Driver))
		switch dbConfig.Driver {
		case "mysql", "postgres":
			sqlxConn, err := sqlx.ConnectContext(ctx, dbConfig.Driver, dbConfig.DSN)
			if err != nil {
				err = fmt.Errorf("error connecting to database %s: %w", key, err)
				return err
			}

			// don't forget to register in closer, using unique name to track in the Log
			i.dbSQL[key] = sqlxConn
			i.dbDriver[key] = dbConfig.Driver
			i.closer = append(i.closer, newNamedCloser(key, sqlxConn))
		}
	}

	return nil
}
