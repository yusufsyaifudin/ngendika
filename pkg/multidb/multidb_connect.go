package multidb

import (
	"database/sql"
	"fmt"
	sqldblogger "github.com/simukti/sqldb-logger"
	"strings"

	_ "github.com/lib/pq"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type SqlDbConnMakerConfig struct {
	Config DatabaseResources `validate:"required"`
}

type SqlDbConnMaker struct {
	conf     DatabaseResources
	disabled map[string]struct{} // list of disabled databases, using struct for minimal memory footprint
	dbSQL    map[string]*sqlx.DB // db key name => real connection
	dbDriver map[string]Driver   // db key name => driver name
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
		conf:     conf.Config,
		disabled: make(map[string]struct{}),
		dbSQL:    make(map[string]*sqlx.DB),
		dbDriver: make(map[string]Driver),
		closer:   make([]Closer, 0),
	}

	err = instance.connect()
	if err != nil {
		// close previous opened connection if error happen
		if _err := instance.Close(); _err != nil {
			err = fmt.Errorf("close db sql error: %w: %s", err, _err)
		}

		return nil, err
	}

	return instance, nil
}

func (i *SqlDbConnMaker) GetSqlx(driver Driver, key string) (*sqlx.DB, error) {
	_, exists := i.disabled[key]
	if exists {
		return nil, fmt.Errorf("db with key '%s' is disabled", key)
	}

	dbConnection, ok := i.dbSQL[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' is not exist on db list", key)
	}

	registeredDriver, ok := i.dbDriver[key]
	if ok && driver == registeredDriver {
		return dbConnection, nil
	}

	return nil, fmt.Errorf("db key '%s' not using driver %s", key, driver)
}

func (i *SqlDbConnMaker) Close() error {
	var err error
	for _, c := range i.closer {
		if c == nil {
			continue
		}

		if e := c.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		}
	}

	return err
}

func (i *SqlDbConnMaker) connect() error {
	// Preparing database connection SQL
	for dbLabel, dbConfig := range i.conf {
		dbLabel = strings.TrimSpace(strings.ToLower(dbLabel))
		if err := validator.New().Var(dbLabel, "required,alphanum"); err != nil {
			err = fmt.Errorf("error connecting to database dbLabel '%s': %w", dbLabel, err)
			return err
		}

		if dbConfig.Disable {
			i.disabled[dbLabel] = struct{}{}
			continue
		}

		var (
			sqlxConn *sqlx.DB
			err      error
		)

		switch dbConfig.Driver {
		case Postgres:
			driver := dbConfig.Driver.String()
			dsn := dbConfig.Postgres.DSN

			db, err := sql.Open(driver, dsn)
			if err != nil {
				err = fmt.Errorf("cannot open db connection '%s': %w", dbLabel, err)
				return err
			}

			if dbConfig.Postgres.Debug {
				db = sqldblogger.OpenDriver(dsn, db.Driver(), &QueryLogger{}, sqldblogger.WithConnectionIDFieldname(dbLabel))
			}

			sqlxConn = sqlx.NewDb(db, dbConfig.Driver.String())

		default:
			err = fmt.Errorf("not supported driver '%s'", dbConfig.Driver)
			return err
		}

		if err != nil {
			err = fmt.Errorf("error connecting to database %s: %w", dbLabel, err)
			return err
		}

		// don't forget to register in closer, using unique name to track in the Log
		i.dbSQL[dbLabel] = sqlxConn
		i.dbDriver[dbLabel] = dbConfig.Driver
		i.closer = append(i.closer, newNamedCloser(dbLabel, sqlxConn))
	}

	return nil
}
