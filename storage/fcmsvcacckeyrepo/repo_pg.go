package fcmsvcacckeyrepo

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/storage"
)

const (
	sqlCreateNewFCMServiceAccKey = `
		INSERT INTO fcm_service_account_keys (
			id, app_id, service_account_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;
`

	sqlSelectFCMServiceAccKey = `
		SELECT * FROM fcm_service_account_keys WHERE app_id = $1 ORDER BY created_at DESC;
`
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

func (p *RepoPostgres) CreateFCMServiceAccountKey(ctx context.Context, param FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error) {
	err = validator.New().Struct(param)
	if err != nil {
		err = fmt.Errorf("%w: %s", storage.ErrValidation, err)
		return
	}

	serviceAccountKey, err := json.Marshal(param.ServiceAccountKey)
	if err != nil {
		err = fmt.Errorf("marshalling fcm service account key error: %w", err)
		return
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, sqlCreateNewFCMServiceAccKey,
		param.ID, param.AppID, string(serviceAccountKey), param.CreatedAt,
	)

	return
}

func (p *RepoPostgres) GetFCMSvcAccKeys(ctx context.Context, appID string) (fcmServiceAccountKeys []FCMServiceAccountKey, err error) {
	err = sqlx.SelectContext(ctx, p.Config.Connection, &fcmServiceAccountKeys, sqlSelectFCMServiceAccKey, appID)
	return
}
