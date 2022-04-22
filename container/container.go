package container

import (
	"context"
	"fmt"

	"github.com/yusufsyaifudin/ngendika/pkg/multidb"

	_ "github.com/lib/pq"

	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
	"go.uber.org/multierr"
)

// Container is an abstraction layer to be used in use-case to stitch all business logic.
// Use this when you pass into another struct.
type Container interface {
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
	dbSqlConn, err := multidb.NewSqlDbConnMaker(multidb.SqlDbConnMakerConfig{
		Ctx:    ctx,
		Config: multidb.Database(conf.Database),
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

// AppRepo return appstore.Repo and return error when connection is closed or nil.
// This should never have caused panic.
func (a *DefaultContainerImpl) AppRepo() (apprepo.Repo, error) {
	repoConnInfo, ok := a.cfg.Database[a.cfg.AppRepo.Database]
	if !ok {
		err := fmt.Errorf("unknown database key %s on appRepo", a.cfg.AppRepo.Database)
		return nil, err
	}

	sqlConn, err := a.dbSqlConn.Get(repoConnInfo.Driver, a.cfg.AppRepo.Database)
	if err != nil {
		return nil, err
	}

	return apprepo.Postgres(apprepo.RepoPostgresConfig{
		Connection: sqlConn,
	})
}

func (a *DefaultContainerImpl) FCMRepo() (fcmrepo.Repo, error) {
	repoConnInfo, ok := a.cfg.Database[a.cfg.FCMRepo.Database]
	if !ok {
		err := fmt.Errorf("unknown database key %s on fcmRepo", a.cfg.FCMRepo.Database)
		return nil, err
	}

	sqlConn, err := a.dbSqlConn.Get(repoConnInfo.Driver, a.cfg.AppRepo.Database)
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
	if _err := a.dbSqlConn.CloseAll(); _err != nil {
		err = multierr.Append(err, fmt.Errorf("close db error: %w", _err))
	}

	return err
}
