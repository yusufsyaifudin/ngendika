package apprepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

const (
	sqlCreateApp = `INSERT INTO apps (client_id, name, enabled, created_at) VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlGetAppByClientID = `SELECT * FROM apps WHERE LOWER(client_id) = $1 LIMIT 1;`
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

func (p *RepoPostgres) CreateApp(ctx context.Context, in InputCreateApp) (out OutCreateApp, err error) {
	err = validator.New().Struct(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	app := in.App
	app.ClientID = strings.TrimSpace(strings.ToLower(app.ClientID))

	insertedApp := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &insertedApp, sqlCreateApp,
		app.ClientID, app.Name, app.Enabled, app.CreatedAt,
	)

	if err != nil {
		return
	}

	out = OutCreateApp{
		App: insertedApp,
	}
	return
}

func (p *RepoPostgres) GetApp(ctx context.Context, in InputGetApp) (out OutGetApp, err error) {
	err = validator.New().Struct(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	appData := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &appData, sqlGetAppByClientID, in.ClientID)
	if err != nil {
		return
	}

	if in.Enabled != nil && *in.Enabled != appData.Enabled {
		err = fmt.Errorf(
			"app id %s is in enabled=%t state, you request enabled=%+v",
			in.ClientID, appData.Enabled, in.Enabled,
		)
		return
	}

	out = OutGetApp{
		App: appData,
	}
	return
}
