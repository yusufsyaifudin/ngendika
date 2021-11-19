package fcmserverkeyrepo

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/storage"
)

const (
	sqlCreateNewFCMServerKey = `
		INSERT INTO fcm_server_keys (id, app_id, server_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;
`

	sqlSelectFCMServerKey = `SELECT * FROM fcm_server_keys WHERE app_id = $1 ORDER BY created_at DESC;`
)

type RepoPostgresConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type RepoPostgres struct {
	Config RepoPostgresConfig
}

var _ Repo = (*RepoPostgres)(nil)

// Postgres return repo interface which implements using PgSQL
func Postgres(conf RepoPostgresConfig) (service *RepoPostgres, err error) {
	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	service = &RepoPostgres{
		Config: conf,
	}
	return
}

func (p *RepoPostgres) CreateFCMServerKey(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error) {
	err = validator.New().Struct(param)
	if err != nil {
		err = fmt.Errorf("%w: %s", storage.ErrValidation, err)
		return
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, sqlCreateNewFCMServerKey,
		param.ID, param.AppID, param.ServerKey, param.CreatedAt,
	)
	return
}

func (p *RepoPostgres) GetFCMServerKeys(ctx context.Context, appID string) (serverKeys []FCMServerKey, err error) {
	err = sqlx.SelectContext(ctx, p.Config.Connection, &serverKeys, sqlSelectFCMServerKey, appID)
	return
}
