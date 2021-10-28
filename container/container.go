package container

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/storage/fcmserverkeyrepo"
	"github.com/yusufsyaifudin/ngendika/storage/fcmsvcacckeyrepo"
)

// Container is an abstraction layer to be used in use-case to stitch all business logic.
// Use this when you pass into another struct.
type Container interface {
	AppRepo() apprepo.Repo
	FCMServerKeyRepo() fcmserverkeyrepo.Repo
	FCMServiceAccountKeyRepo() fcmsvcacckeyrepo.Repo
}

// DefaultContainerImpl the real implementation of Container
type DefaultContainerImpl struct {
	ctx              context.Context
	cfg              config.Config         `validate:"required"`
	appRepo          apprepo.Repo          `validate:"required"`
	fcmServerKeyRepo fcmserverkeyrepo.Repo `validate:"required"`
	fcmSvcAccKeyRepo fcmsvcacckeyrepo.Repo `validate:"required"`
	closer           []Closer              `validate:"required"` // this to trace what dependencies this closer is related
}

// Ensure that DefaultContainerImpl implements DefaultContainerImpl
var _ Container = (*DefaultContainerImpl)(nil)

// AppRepo return appstore.Repo and return error when connection is closed or nil.
// This should never have caused panic.
func (a *DefaultContainerImpl) AppRepo() apprepo.Repo {
	return a.appRepo
}

func (a *DefaultContainerImpl) FCMServerKeyRepo() fcmserverkeyrepo.Repo {
	return a.fcmServerKeyRepo
}

func (a *DefaultContainerImpl) FCMServiceAccountKeyRepo() fcmsvcacckeyrepo.Repo {
	return a.fcmSvcAccKeyRepo
}

// Close will close all dependencies.
// This using pointer so can update IsClosed value and close the real connection.
func (a *DefaultContainerImpl) Close() error {
	ctx := logger.Inject(context.Background(), logger.Tracer{AppTraceID: "startup"})
	logger.Debug(ctx, "dependencies: trying to close")

	if a.closer == nil {
		logger.Debug(ctx, "dependencies: nothing to close")
		return nil
	}

	return CloseAllSQLConnection(a.ctx, a.closer)
}

// Setup return pointer because it heavily use.
// This will initialize all required dependencies to run.
// This will return DefaultContainerImpl instead Container,
// the reason is when Setup called it must be close in deferred mode, any passed value using interface
// won't let user Close any dependencies during run-time.
func Setup(ctx context.Context, conf config.Config) (*DefaultContainerImpl, error) {
	var (
		appRepo          apprepo.Repo
		FCMServerKeyRepo fcmserverkeyrepo.Repo
		FCMSvcAccKeyRepo fcmsvcacckeyrepo.Repo

		closer = make([]Closer, 0)
	)

	dbSQL, dbSQLCloser, err := SQLConnection(ctx, conf.Database)
	if err != nil {
		return nil, err
	}

	// don't forget add closer
	closer = append(closer, dbSQLCloser...)

	// ===== Prepare App Repository
	appRepoDriver, ok := conf.Database[conf.AppRepo.Database]
	if !ok {
		err = fmt.Errorf("prepare app repository error: unknown database key %s", conf.AppRepo.Database)
		return nil, err
	}

	appRepoSQL := dbSQL[conf.AppRepo.Database]
	switch appRepoDriver.Driver {
	case "postgres":
		appRepo, err = apprepo.Postgres(apprepo.RepoPostgresConfig{
			Connection: appRepoSQL,
		})

	default:
		err = fmt.Errorf("unknown driver %s", appRepoDriver.Driver)
	}

	if err != nil {
		err = fmt.Errorf("prepare app repository error: %w", err)
		return nil, err
	}

	// ==== Prepare FCM Server Key Repository
	FCMServerKeyRepoDriver, ok := conf.Database[conf.FCMServerKeyRepo.Database]
	if !ok {
		err = fmt.Errorf("prepare fcm server key repository error: unknown database key %s", conf.FCMServerKeyRepo.Database)
		return nil, err
	}

	FCMServerKeyRepoSQL := dbSQL[conf.FCMServiceAccountKeyRepo.Database]
	switch FCMServerKeyRepoDriver.Driver {
	case "postgres":
		FCMServerKeyRepo, err = fcmserverkeyrepo.Postgres(fcmserverkeyrepo.RepoPostgresConfig{
			Connection: FCMServerKeyRepoSQL,
		})

	default:
		err = fmt.Errorf("unknown driver %s", FCMServerKeyRepoDriver.Driver)
	}

	if err != nil {
		err = fmt.Errorf("prepare fcm server key repository error: %w", err)
		return nil, err
	}

	// ==== Prepare FCM Service Account Key Repository
	FCMSvcAccKeyRepoDriver, ok := conf.Database[conf.FCMServiceAccountKeyRepo.Database]
	if !ok {
		err = fmt.Errorf("prepare fcm service account key repository error: unknown database key %s", conf.FCMServiceAccountKeyRepo.Database)
		return nil, err
	}

	FCMSvcAccKeyRepoSQL := dbSQL[conf.FCMServiceAccountKeyRepo.Database]
	switch FCMSvcAccKeyRepoDriver.Driver {
	case "postgres":
		FCMSvcAccKeyRepo, err = fcmsvcacckeyrepo.Postgres(fcmsvcacckeyrepo.RepoPostgresConfig{
			Connection: FCMSvcAccKeyRepoSQL,
		})

	default:
		err = fmt.Errorf("unknown driver %s", FCMSvcAccKeyRepoDriver.Driver)
	}

	if err != nil {
		err = fmt.Errorf("prepare fcm service account key repository error: %w", err)
		return nil, err
	}

	dep := &DefaultContainerImpl{
		ctx:              ctx,
		cfg:              conf,
		appRepo:          appRepo,
		fcmServerKeyRepo: FCMServerKeyRepo,
		fcmSvcAccKeyRepo: FCMSvcAccKeyRepo,
		closer:           closer,
	}

	err = validator.New().Struct(dep)
	if err != nil {
		return nil, err
	}

	return dep, nil
}

func SQLConnection(ctx context.Context, dbConfigMap config.Database) (map[string]*sqlx.DB, []Closer, error) {
	// Preparing database connection SQL
	var dbSQL = map[string]*sqlx.DB{}
	var closer = make([]Closer, 0)

	for key, dbConfig := range dbConfigMap {
		key = strings.TrimSpace(strings.ToLower(key))
		if ok, err := regexp.MatchString("", key); err != nil || !ok {
			err = fmt.Errorf("error connecting to database key '%s' must be alphanumeric only", key)

			// close previous opened db if error happen
			if _err := CloseAllSQLConnection(ctx, closer); _err != nil {
				err = fmt.Errorf("%s: close db sql error: %w", err, _err)
			}

			return nil, []Closer{}, err
		}

		switch dbConfig.Driver {
		case "mysql", "postgres":
			sqlxConn, err := sqlx.Connect(dbConfig.Driver, dbConfig.DSN)
			if err != nil {
				err = fmt.Errorf("error connecting to database %s: %w", key, err)

				// close previous opened db if error happen
				if _err := CloseAllSQLConnection(ctx, closer); _err != nil {
					err = fmt.Errorf("%s: close db sql error: %w", err, _err)
				}

				return nil, []Closer{}, err
			}

			// don't forget to register in closer, using unique name to track in the Log
			dbSQL[key] = sqlxConn
			closer = append(closer, NewNamedCloser(key, sqlxConn))
		}
	}

	return dbSQL, closer, nil
}

func CloseAllSQLConnection(ctx context.Context, closers []Closer) error {
	logger.Debug(ctx, "dependencies: trying to close")

	var err error
	for _, closer := range closers {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		} else {
			logger.Debug(ctx, fmt.Sprintf("dependencies: %s success to close", closer.Name()))
		}
	}

	if err != nil {
		logger.Error(ctx, "dependencies: some error occurred when closing dep", logger.KV("error", err))
	}

	return err
}
