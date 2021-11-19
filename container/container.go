package container

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/storage/fcmserverkeyrepo"
	"github.com/yusufsyaifudin/ngendika/storage/fcmsvcacckeyrepo"
	"go.uber.org/multierr"
)

// Container is an abstraction layer to be used in use-case to stitch all business logic.
// Use this when you pass into another struct.
type Container interface {
	AppRepo() (apprepo.Repo, error)
	FCMServerKeyRepo() (fcmserverkeyrepo.Repo, error)
	FCMServiceAccountKeyRepo() (fcmsvcacckeyrepo.Repo, error)
	GetRedisConn(name string) (redis.UniversalClient, error)
}

// DefaultContainerImpl the real implementation of Container
type DefaultContainerImpl struct {
	ctx       context.Context `validate:"required"`
	cfg       config.Config   `validate:"required"`
	dbSqlConn *SqlDbConnMaker `validate:"required,structonly"` // all database connection
	redisConn *RedisConnMaker `validate:"required,structonly"` // all redis connection
}

// Ensure that DefaultContainerImpl implements DefaultContainerImpl
var _ Container = (*DefaultContainerImpl)(nil)

// Setup return pointer because it heavily used.
// This will initialize all required dependencies to run.
// This will return DefaultContainerImpl instead Container,
// the reason is when Setup called it must be close in deferred mode, any passed value using interface
// won't let user Close any dependencies during run-time.
func Setup(ctx context.Context, conf config.Config) (*DefaultContainerImpl, error) {
	dbSqlConn, err := NewSqlDbConnMaker(ctx, conf.Database)
	if err != nil {
		return nil, err
	}

	redisConn, err := NewRedisConnMaker(ctx, conf.Redis)
	if err != nil {
		return nil, err
	}

	dep := &DefaultContainerImpl{
		ctx:       ctx,
		cfg:       conf,
		dbSqlConn: dbSqlConn,
		redisConn: redisConn,
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
		err := fmt.Errorf("unknown database key %s", a.cfg.AppRepo.Database)
		return nil, err
	}

	switch repoConnInfo.Driver {
	case "mysql", "postgres":
		sqlConn, err := a.dbSqlConn.Get(a.cfg.AppRepo.Database)
		if err != nil {
			return nil, err
		}

		return apprepo.Postgres(apprepo.RepoPostgresConfig{
			Connection: sqlConn,
		})

	default:
		err := fmt.Errorf("unknown driver %s", repoConnInfo.Driver)
		return nil, err
	}
}

func (a *DefaultContainerImpl) FCMServerKeyRepo() (fcmserverkeyrepo.Repo, error) {
	repoConnInfo, ok := a.cfg.Database[a.cfg.FCMServerKeyRepo.Database]
	if !ok {
		err := fmt.Errorf("unknown database key %s", a.cfg.FCMServerKeyRepo.Database)
		return nil, err
	}

	switch repoConnInfo.Driver {
	case "mysql", "postgres":
		sqlConn, err := a.dbSqlConn.Get(a.cfg.FCMServiceAccountKeyRepo.Database)
		if err != nil {
			return nil, err
		}

		return fcmserverkeyrepo.Postgres(fcmserverkeyrepo.RepoPostgresConfig{
			Connection: sqlConn,
		})

	default:
		err := fmt.Errorf("unknown driver %s", repoConnInfo.Driver)
		return nil, err
	}
}

func (a *DefaultContainerImpl) FCMServiceAccountKeyRepo() (fcmsvcacckeyrepo.Repo, error) {
	repoConnInfo, ok := a.cfg.Database[a.cfg.FCMServiceAccountKeyRepo.Database]
	if !ok {
		err := fmt.Errorf(" unknown database key %s", a.cfg.FCMServiceAccountKeyRepo.Database)
		return nil, err
	}

	switch repoConnInfo.Driver {
	case "mysql", "postgres":
		sqlConn, err := a.dbSqlConn.Get(a.cfg.FCMServiceAccountKeyRepo.Database)
		if err != nil {
			return nil, err
		}

		return fcmsvcacckeyrepo.Postgres(fcmsvcacckeyrepo.RepoPostgresConfig{
			Connection: sqlConn,
		})

	default:
		err := fmt.Errorf("unknown driver %s", repoConnInfo.Driver)
		return nil, err
	}
}

func (a *DefaultContainerImpl) GetRedisConn(name string) (redis.UniversalClient, error) {
	return a.redisConn.Get(name)
}

// Close will close all dependencies.
func (a *DefaultContainerImpl) Close() error {

	var err error
	if _err := a.dbSqlConn.CloseAll(); _err != nil {
		err = multierr.Append(err, fmt.Errorf("close db error: %w", _err))
	}

	if _err := a.redisConn.CloseAll(); _err != nil {
		err = multierr.Append(err, fmt.Errorf("close redis error: %w", _err))
	}

	return err
}
