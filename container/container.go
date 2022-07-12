package container

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/pkg/multidb"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
	"go.uber.org/multierr"
)

// Container is an abstraction layer to be used in use-case to stitch all business logic.
// Use this when you pass into another struct.
type Container interface {
	io.Closer

	AppRepo() (apprepo.Repo, error)
	FCMRepo() (fcmrepo.Repo, error)
}

// DefaultContainerImpl the real implementation of Container
type DefaultContainerImpl struct {
	ctx       context.Context `validate:"required"`
	cfg       *config.Config  `validate:"required,structonly"`
	dbSqlConn multidb.MultiDB `validate:"required"` // all database connection
}

// Ensure that DefaultContainerImpl implements DefaultContainerImpl
var _ Container = (*DefaultContainerImpl)(nil)

// Setup return pointer because it heavily used.
// This will initialize all required dependencies to run.
// This will return DefaultContainerImpl instead Container,
// the reason is when Setup called it must be close in deferred mode, any passed value using interface
// won't let user Close any dependencies during run-time.
func Setup(ctx context.Context, conf *config.Config) (*DefaultContainerImpl, error) {
	dbConfig := multidb.DatabaseResources{}
	for name, conn := range conf.DatabaseResources {
		dbConfig[name] = multidb.DatabaseResource{
			Disable:  conn.Disable,
			Driver:   multidb.Driver(conn.Driver),
			Mysql:    multidb.GoSqlDb(conn.Mysql),
			Postgres: multidb.GoSqlDb(conn.Postgres),
		}
	}

	dbSqlConn, err := multidb.NewSqlDbConnMaker(multidb.SqlDbConnMakerConfig{
		Ctx:    ctx,
		Config: dbConfig,
	})
	if err != nil {
		return nil, err
	}

	dep := &DefaultContainerImpl{
		ctx:       ctx,
		cfg:       conf,
		dbSqlConn: dbSqlConn,
	}

	err = validator.New().Struct(dep)
	if err != nil {
		return nil, err
	}

	return dep, nil
}

// AppRepo return apprepo.Repo and return error when connection is closed or nil.
// This should never have caused panic.
func (a *DefaultContainerImpl) AppRepo() (apprepo.Repo, error) {
	repoConnInfo, ok := a.cfg.DatabaseResources[a.cfg.AppRepo.DBLabel]
	if !ok {
		err := fmt.Errorf("unknown database key %s on appRepo", a.cfg.AppRepo.DBLabel)
		return nil, err
	}

	sqlConn, err := a.dbSqlConn.GetSqlx(multidb.Driver(repoConnInfo.Driver), a.cfg.AppRepo.DBLabel)
	if err != nil {
		return nil, err
	}

	return apprepo.Postgres(apprepo.RepoPostgresConfig{
		Connection: sqlConn,
	})
}

func (a *DefaultContainerImpl) FCMRepo() (fcmrepo.Repo, error) {
	repoConnInfo, ok := a.cfg.DatabaseResources[a.cfg.FCMRepo.DBLabel]
	if !ok {
		err := fmt.Errorf("unknown database key %s on fcmRepo", a.cfg.FCMRepo.DBLabel)
		return nil, err
	}

	sqlConn, err := a.dbSqlConn.GetSqlx(multidb.Driver(repoConnInfo.Driver), a.cfg.FCMRepo.DBLabel)
	if err != nil {
		return nil, err
	}

	return fcmrepo.NewPostgres(fcmrepo.PostgresConfig{
		Connection: sqlConn,
	})
}

// Close will close all dependencies.
func (a *DefaultContainerImpl) Close() error {

	var err error
	if _err := a.dbSqlConn.Close(); _err != nil {
		err = multierr.Append(err, fmt.Errorf("close db error: %w", _err))
	}

	return err
}
