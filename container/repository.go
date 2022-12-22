package container

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnprepo"
	"io"

	"github.com/yusufsyaifudin/ngendika/internal/svc/apprepo"
	"github.com/yusufsyaifudin/ngendika/pkg/multidb"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.uber.org/multierr"
)

// Repositories is an abstraction layer to list down all repositories.
// This only will connect and save the repository.
// To use this, you must select the db label based on config file
type Repositories interface {
	io.Closer

	AppRepo(dbLabel string) (apprepo.Repo, error)
	PNProviderRepo(dbLabel string) (pnprepo.Repo, error)
}

// RepositoryImpl the real implementation of Repositories
type RepositoryImpl struct {
	dbResourceMap ConfigDatabaseResources `validate:"required,structonly"`
	dbSqlConn     multidb.MultiDB         `validate:"required"` // all database connection
}

// Ensure that RepositoryImpl implements RepositoryImpl
var _ Repositories = (*RepositoryImpl)(nil)

// SetupRepositories return pointer because it heavily used.
// This will initialize all required dependencies to run.
// This will return RepositoryImpl instead Repositories,
// the reason is when SetupRepositories called it must be close in deferred mode, any passed value using interface
// won't let user Close any dependencies during run-time.
func SetupRepositories(conf ConfigDatabaseResources) (*RepositoryImpl, error) {
	sqlDbConfig := multidb.DatabaseResources{}
	for name, conn := range conf {
		sqlDbConfig[name] = multidb.DatabaseResource{
			Disable:  conn.Disable,
			Driver:   multidb.Driver(conn.Driver),
			Postgres: multidb.GoSqlDb(conn.Postgres),
		}

	}

	dbSqlConn, err := multidb.NewSqlDbConnMaker(multidb.SqlDbConnMakerConfig{Config: sqlDbConfig})
	if err != nil {
		return nil, err
	}

	dep := &RepositoryImpl{
		dbResourceMap: conf,
		dbSqlConn:     dbSqlConn,
	}

	err = validator.Validate(dep)
	if err != nil {
		return nil, err
	}

	return dep, nil
}

// AppRepo return apprepo.Repo and return error when connection is closed or nil.
// This should never have caused panic.
func (r *RepositoryImpl) AppRepo(dbLabel string) (appRepo apprepo.Repo, err error) {
	repoConnInfo, ok := r.dbResourceMap[dbLabel]
	if !ok {
		err = fmt.Errorf("unknown database key %s on appRepo", dbLabel)
		return
	}

	// for type postgres use sqlx, for type mongo use mongodb
	sqlDriver := repoConnInfo.Driver
	switch sqlDriver {
	case "postgres":
		var sqlConn *sqlx.DB
		sqlConn, err = r.dbSqlConn.GetSqlx(multidb.Postgres, dbLabel)
		if err != nil {
			return
		}

		cfg := apprepo.RepoPostgresConfig{
			Connection: sqlConn,
		}

		appRepo, err = apprepo.Postgres(cfg)
		return

	default:
		err = fmt.Errorf("not supported db driver '%s' on label '%s'", sqlDriver, dbLabel)
		return
	}

}

func (r *RepositoryImpl) PNProviderRepo(dbLabel string) (repo pnprepo.Repo, err error) {
	repoConnInfo, ok := r.dbResourceMap[dbLabel]
	if !ok {
		err = fmt.Errorf("unknown database key %s on fcmRepo", dbLabel)
		return
	}

	// for type postgres use sqlx, for type mongo use mongodb
	sqlDriver := repoConnInfo.Driver
	switch sqlDriver {
	case "postgres":
		var sqlConn *sqlx.DB
		sqlConn, err = r.dbSqlConn.GetSqlx(multidb.Postgres, dbLabel)
		if err != nil {
			return nil, err
		}

		cfg := pnprepo.PostgresConfig{
			Connection: sqlConn,
		}

		repo, err = pnprepo.NewPostgres(cfg)
		return

	default:
		err = fmt.Errorf("not supported db driver '%s' on label '%s'", sqlDriver, dbLabel)
		return
	}
}

// Close will close all dependencies.
func (r *RepositoryImpl) Close() error {
	if r == nil {
		return nil
	}

	if r.dbSqlConn == nil {
		return nil
	}

	var err error
	if _err := r.dbSqlConn.Close(); _err != nil {
		err = multierr.Append(err, fmt.Errorf("close db error: %w", _err))
	}

	return err
}
