package fcmrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

const (
	SqlCreateNewFCMServiceAccKey = `
		INSERT INTO fcm_service_account_keys (
			id, app_id, service_account_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;
`

	SqlSelectFCMServiceAccKey = `
		SELECT * FROM fcm_service_account_keys WHERE app_id = $1 ORDER BY created_at DESC;
`
)

type PostgresFCMServiceAccountKeyConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type PostgresFCMServiceAccountKey struct {
	Config PostgresFCMServiceAccountKeyConfig
}

var _ RepoFCMServiceAccountKey = (*PostgresFCMServiceAccountKey)(nil)

// NewPostgresFCMServiceAccountKey return repo interface which implements using PgSQL
// ** table fcm_service_account_keys
func NewPostgresFCMServiceAccountKey(conf PostgresFCMServiceAccountKeyConfig) (service *PostgresFCMServiceAccountKey, err error) {
	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	service = &PostgresFCMServiceAccountKey{
		Config: conf,
	}
	return
}

func (p *PostgresFCMServiceAccountKey) Create(ctx context.Context, param FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error) {
	err = validator.New().Struct(param)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	serviceAccountKey, err := json.Marshal(param.ServiceAccountKey)
	if err != nil {
		err = fmt.Errorf("marshalling fcm service account key error: %w", err)
		return
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, SqlCreateNewFCMServiceAccKey,
		param.ID, param.AppID, string(serviceAccountKey), param.CreatedAt,
	)

	return
}

func (p *PostgresFCMServiceAccountKey) FetchAll(ctx context.Context, appID string) (fcmServiceAccountKeys []FCMServiceAccountKey, err error) {
	err = sqlx.SelectContext(ctx, p.Config.Connection, &fcmServiceAccountKeys, SqlSelectFCMServiceAccKey, appID)
	return
}
