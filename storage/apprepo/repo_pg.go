package apprepo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/storage"
)

type RepoPostgresConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type RepoPostgres struct {
	Config RepoPostgresConfig
}

func (p *RepoPostgres) CreateApp(ctx context.Context, app App) (insertedApp App, err error) {
	clientID := strings.ToLower(strings.TrimSpace(app.ClientID))
	if clientID == "" {
		err = storage.ErrAppWrongClientID
		return
	}

	name := strings.TrimSpace(app.Name)
	if name == "" {
		err = storage.ErrAppWrongName
		return
	}

	enabled := app.Enabled
	createdAt := app.CreatedAt
	if createdAt.IsZero() || createdAt.Unix() <= 0 {
		createdAt = time.Now().UTC()
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &insertedApp, sqlCreateApp, clientID, name, enabled, createdAt)
	return
}

func (p *RepoPostgres) GetAppByClientID(ctx context.Context, clientID string) (appData App, err error) {
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	if clientID == "" {
		return appData, storage.ErrAppWrongClientID
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &appData, sqlGetAppByClientID, clientID)
	return
}

func (p *RepoPostgres) CreateFCMServiceAccountKey(ctx context.Context, cert FCMServiceAccountKey) (inserted FCMServiceAccountKey, err error) {
	clientID := strings.ToLower(strings.TrimSpace(cert.AppClientID))
	if clientID == "" {
		return inserted, storage.ErrAppWrongClientID
	}

	serviceAccountKey, err := json.Marshal(cert.ServiceAccountKey)
	if err != nil {
		err = fmt.Errorf("marshalling fcm service account key error: %w", err)
		return
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, sqlCreateNewFCMServiceAccKey,
		cert.ID, clientID, string(serviceAccountKey), cert.CreatedAt,
	)

	return
}

func (p *RepoPostgres) GetFCMServiceAccountKeys(ctx context.Context, clientID string) (fcmKeys []FCMServiceAccountKey, err error) {
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	if clientID == "" {
		return nil, storage.ErrAppWrongClientID
	}

	err = sqlx.SelectContext(ctx, p.Config.Connection, &fcmKeys, sqlSelectFCMServiceAccKey, clientID)
	return
}

func (p *RepoPostgres) CreateFCMServerKey(ctx context.Context, param FCMServerKey) (inserted FCMServerKey, err error) {
	clientID := strings.ToLower(strings.TrimSpace(param.AppClientID))
	if clientID == "" {
		return inserted, storage.ErrAppWrongClientID
	}

	serverKey := strings.ToLower(strings.TrimSpace(param.ServerKey))
	if serverKey == "" {
		return inserted, fmt.Errorf("%w: required server key", storage.ErrValidation)
	}

	err = sqlx.GetContext(ctx, p.Config.Connection, &inserted, sqlCreateNewFCMServerKey,
		param.ID, param.AppClientID, param.ServerKey, param.CreatedAt,
	)
	return
}

func (p *RepoPostgres) GetFCMServerKeys(ctx context.Context, clientID string) (serverKeys []FCMServerKey, err error) {
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	if clientID == "" {
		return nil, storage.ErrAppWrongClientID
	}

	err = sqlx.SelectContext(ctx, p.Config.Connection, &serverKeys, sqlSelectFCMServerKey, clientID)
	return
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
