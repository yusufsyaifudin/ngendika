package fcmrepo

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

const (
	SqlCreateNewFCMServerKey = `
		INSERT INTO fcm_server_keys (id, app_id, server_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;
`

	SqlSelectFCMServerKey = `SELECT * FROM fcm_server_keys WHERE app_id = $1 ORDER BY created_at DESC;`
)

type PostgresFCMServerKeyConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type PostgresFCMServerKey struct {
	Config PostgresFCMServerKeyConfig
}

var _ RepoFCMServerKey = (*PostgresFCMServerKey)(nil)

// NewPostgresFCMServerKey return repo interface which implements using PgSQL
// ** table fcm_server_keys
func NewPostgresFCMServerKey(conf PostgresFCMServerKeyConfig) (service *PostgresFCMServerKey, err error) {
	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	service = &PostgresFCMServerKey{
		Config: conf,
	}
	return
}

func (p *PostgresFCMServerKey) Create(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error) {
	err = validator.New().Struct(param)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, SqlCreateNewFCMServerKey,
		param.ID, param.AppID, param.ServerKey, param.CreatedAt,
	)
	return
}

func (p *PostgresFCMServerKey) FetchAll(ctx context.Context, appID string) (serverKeys []FCMServerKey, err error) {
	err = sqlx.SelectContext(ctx, p.Config.Connection, &serverKeys, SqlSelectFCMServerKey, appID)
	return
}
