package apprepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

const (
	sqlCreateApp        = `INSERT INTO apps (id, client_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING *;`
	sqlGetAppByClientID = `SELECT * FROM apps WHERE LOWER(client_id) = $1 AND deleted_at = 0 LIMIT 1;`

	sqlUpsertApp = `
		INSERT INTO apps (id, client_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (LOWER(client_id), deleted_at) 
		DO UPDATE SET 
		    name = EXCLUDED.name,
		    updated_at = EXCLUDED.updated_at
		WHERE LOWER(apps.client_id) = $2 AND apps.deleted_at = 0
		RETURNING *;
`

	sqlListAppsCount        = `SELECT COUNT(*) as total FROM apps WHERE deleted_at = 0;`
	sqlListAppsWithoutRange = `SELECT * FROM apps WHERE deleted_at = 0 ORDER BY id ASC LIMIT $1;`
	sqlListAppsWithRange    = `SELECT * FROM apps WHERE (id > $1 AND id < $2) AND deleted_at = 0 ORDER BY id ASC LIMIT $3;`
	sqlListAppsAfterID      = `SELECT * FROM apps WHERE id > $1 AND deleted_at = 0 ORDER BY id ASC LIMIT $2;`

	// sorting to ASC, this to ensure for example we have limit 5 and before_id 12
	// we may have [11, 10, 9, 8, 7] from database (DESC)
	// to make it consistent, we reverse to ASC order [7, 8, 9, 10, 11]
	sqlListAppsBeforeID = `SELECT * FROM (SELECT * FROM apps WHERE id < $1 AND deleted_at = 0 ORDER BY id DESC LIMIT $2) AS tmp ORDER BY tmp.id ASC;`
	sqlSoftDeleteApp    = `UPDATE apps SET deleted_at = $1 WHERE id = (SELECT id FROM apps WHERE LOWER(apps.client_id) = $2 AND apps.deleted_at = 0 LIMIT 1) RETURNING *;`
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
	err = validator.Validate(conf)
	if err != nil {
		return nil, err
	}

	service = &RepoPostgres{
		Config: conf,
	}
	return
}

func (p *RepoPostgres) Create(ctx context.Context, in InputCreate) (out OutCreate, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	app := in.App
	app.ClientID = strings.TrimSpace(strings.ToLower(app.ClientID))

	insertedApp := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &insertedApp, sqlCreateApp,
		in.App.ID, app.ClientID, app.Name, app.CreatedAt, app.UpdatedAt,
	)

	if err != nil {
		return
	}

	out = OutCreate{
		App: insertedApp,
	}
	return
}

func (p *RepoPostgres) Upsert(ctx context.Context, in InputUpsert) (out OutUpsert, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	app := in.App
	app.ClientID = strings.TrimSpace(strings.ToLower(app.ClientID))

	insertedApp := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &insertedApp, sqlUpsertApp,
		in.App.ID, app.ClientID, app.Name, app.CreatedAt, app.UpdatedAt,
	)

	if err != nil {
		return
	}

	out = OutUpsert{
		App: insertedApp,
	}
	return
}

func (p *RepoPostgres) GetByClientID(ctx context.Context, in InputGetByClientID) (out OutGetByClientID, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "apprepo.GetByClientID")
	defer span.End()

	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	appData := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &appData, sqlGetAppByClientID, in.ClientID)
	if err != nil {
		return
	}

	out = OutGetByClientID{
		App: appData,
	}
	return
}

// List all query is exclusive, means that before_id and after_id will not be in the result
func (p *RepoPostgres) List(ctx context.Context, in InputList) (out OutList, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	count := struct {
		Total int64 `db:"total"`
	}{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &count, sqlListAppsCount)
	if err != nil {
		err = fmt.Errorf("cannot count list of apps: %w", err)
		return
	}

	if count.Total <= 0 {
		return
	}

	appData := make([]App, 0)

	switch {
	case in.BeforeID == 0 && in.AfterID == 0:
		err = sqlx.SelectContext(ctx, p.Config.Connection, &appData, sqlListAppsWithoutRange, in.Limit)

	case in.BeforeID == 0 && in.AfterID != 0:
		// in: before_id = 0 and after_id = 100
		// sql: (id > 100)
		err = sqlx.SelectContext(ctx, p.Config.Connection, &appData, sqlListAppsAfterID, in.AfterID, in.Limit)

	case in.BeforeID != 0 && in.AfterID == 0:
		// in: before_id = 100 and after_id = 0
		// sql: (id < 100)
		err = sqlx.SelectContext(ctx, p.Config.Connection, &appData, sqlListAppsBeforeID, in.BeforeID, in.Limit)

	case in.AfterID > in.BeforeID:
		// in: before_id = 10 and after_id = 12
		// cannot do this, because:
		// before id 10 we have 9, 8, 7, ... 0
		// after id 12 we have 13, 14, 15, ... inf+
		// So, the query is error!
		prevData := in.BeforeID
		previous := make([]string, 0)

		nextData := in.AfterID
		next := make([]string, 0)
		for i := 0; i < 3; i++ {
			prevData -= 1
			previous = append(previous, fmt.Sprint(prevData))

			nextData += 1
			next = append(next, fmt.Sprint(nextData))
		}

		previous = append(previous, "-inf")
		next = append(next, "+inf")

		err = fmt.Errorf(
			"cannot do range query: after_id %d we may have %s, and before_id %d we may have %s. "+
				"From two slices, we may never find the subset",
			in.AfterID, next,
			in.BeforeID, previous,
		)
	default:
		// in: before_id = 12 and after_id = 10
		// sql: (id > 10 AND id < 12)
		err = sqlx.SelectContext(ctx, p.Config.Connection, &appData, sqlListAppsWithRange, in.AfterID, in.BeforeID, in.Limit)
	}

	if err != nil {
		err = fmt.Errorf("cannot get list of apps: %w", err)
		return
	}

	out = OutList{
		Total: count.Total,
		Apps:  appData,
	}

	return
}

func (p *RepoPostgres) DelByClientID(ctx context.Context, in InputDelByClientID) (out OutDelByClientID, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	appData := App{}
	err = sqlx.GetContext(ctx, p.Config.Connection, &appData, sqlSoftDeleteApp, in.DeletedAt, in.ClientID)
	if errors.Is(err, sql.ErrNoRows) {
		out = OutDelByClientID{
			Success: false,
		}

		err = nil // discard error
		return
	}

	if err != nil {
		return
	}

	out = OutDelByClientID{
		Success: appData.ClientID == in.ClientID && appData.DeletedAt == in.DeletedAt,
	}
	return
}
