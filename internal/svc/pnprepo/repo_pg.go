package pnprepo

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.opentelemetry.io/otel/trace"
)

const (
	SqlInsert = `
INSERT INTO push_providers (id, app_id, provider, label, credential_json, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7) 
RETURNING *;
`

	// SqlGetByLabels use with sqlx.In so it mush using quote rather than dollar
	SqlGetByLabels = `SELECT * FROM push_providers WHERE app_id = ? AND provider = ? AND label IN (?);`
)

func CreatePartitionSQL(id int64) string {
	return fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS push_provider_app_%d PARTITION OF push_providers FOR VALUES IN (%d);",
		id, id,
	)
}

type PostgresConfig struct {
	Connection sqlx.QueryerContext `validate:"required"`
}

type Postgres struct {
	Config PostgresConfig
}

var _ Repo = (*Postgres)(nil)

func NewPostgres(cfg PostgresConfig) (repo *Postgres, err error) {
	err = validator.Validate(cfg)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	repo = &Postgres{
		Config: cfg,
	}

	return
}

func (p *Postgres) Insert(ctx context.Context, in InputInsert) (out OutInsert, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	sqlCreatePartition := CreatePartitionSQL(in.PnProvider.AppID)
	_, err = p.Config.Connection.QueryContext(ctx, sqlCreatePartition)
	if err != nil {
		err = fmt.Errorf("cannot create partition for app id '%d' error: %w", in.PnProvider.AppID, err)
		return
	}

	args := []interface{}{
		in.PnProvider.ID,
		in.PnProvider.AppID,
		in.PnProvider.Provider,
		in.PnProvider.Label,
		in.PnProvider.CredentialJSON,
		in.PnProvider.CreatedAt,
		in.PnProvider.UpdatedAt,
	}

	var svcProvider PushNotificationProvider
	err = sqlx.GetContext(ctx, p.Config.Connection, &svcProvider, SqlInsert, args...)
	if err != nil {
		err = fmt.Errorf("insert db error: %w", err)
		return
	}

	out = OutInsert{
		PnProvider: svcProvider,
	}

	return
}

func (p *Postgres) GetByLabels(ctx context.Context, in InGetByLabels) (out OutGetByLabels, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "pnprepo.GetByLabels")
	defer span.End()

	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	query, args, err := sqlx.In(SqlGetByLabels, in.AppID, in.Provider, in.Labels)
	if err != nil {
		err = fmt.Errorf("cannot generate sql query: %w", err)
		return
	}

	// query is rebind using $ because we use postrges here
	query = sqlx.Rebind(sqlx.DOLLAR, query)

	var svcProviders []PushNotificationProvider
	err = sqlx.SelectContext(ctx, p.Config.Connection, &svcProviders, query, args...)
	if err != nil {
		err = fmt.Errorf("cannot get config by labels: %w", err)
		return
	}

	out = OutGetByLabels{
		PnProvider: svcProviders,
	}

	return
}
