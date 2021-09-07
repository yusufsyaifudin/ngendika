package pgsql_apprepo

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

// CreateFcmServerKeysTable1624008564 is struct to define a migration with ID 1624008564_create_fcm_server_keys_table
type CreateFcmServerKeysTable1624008564 struct{}

// ID return unique identifier for each migration. The prefix is unix time when this migration is created.
func (m CreateFcmServerKeysTable1624008564) ID(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServerKeysTable1624008564.ID")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return fmt.Sprintf("%d_%s.sql", 1624008564, "create_fcm_server_keys_table")
}

// SequenceNumber return current time when the migration is created,
// this useful to see the current status of the migration.
func (m CreateFcmServerKeysTable1624008564) SequenceNumber(ctx context.Context) int {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServerKeysTable1624008564.SequenceNumber")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return 1624008564
}

// Up return sql migration for sync database
func (m CreateFcmServerKeysTable1624008564) Up(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServerKeysTable1624008564.Up")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `
CREATE TABLE IF NOT EXISTS fcm_server_keys (
	id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
	app_client_id VARCHAR NOT NULL,
	server_key VARCHAR NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

	-- data integrity
	CONSTRAINT fk_fcm_server_keys_app_client_id FOREIGN KEY(app_client_id) REFERENCES apps(client_id) ON DELETE CASCADE
);

-- for faster query WHERE app_client_id = ? ORDER BY created_at DESC
CREATE INDEX idx_fcm_server_keys_app_id_created_at ON fcm_server_keys (app_client_id, created_at DESC);
`

	return
}

// Down return sql migration for rollback database
func (m CreateFcmServerKeysTable1624008564) Down(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServerKeysTable1624008564.Down")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `DROP TABLE IF EXISTS fcm_server_keys;`
	return
}
