package apprepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/storage"
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

func (p *RepoPostgres) CreateApp(ctx context.Context, app App) (insertedApp App, err error) {
	err = validator.New().Struct(app)
	if err != nil {
		err = fmt.Errorf("%w: %s", storage.ErrValidation, err)
		return
	}

	app.ClientID = strings.TrimSpace(strings.ToLower(app.ClientID))

	err = sqlx.GetContext(ctx, p.Config.Connection, &insertedApp, sqlCreateApp,
		app.ClientID, app.Name, app.Enabled, app.CreatedAt,
	)
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
