package pgsql_apprepo

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

// CreateFcmServiceAccountKeysTable1600239931 is struct to define a migration with ID 1600239931_create_fcm_service_account_keys_table
type CreateFcmServiceAccountKeysTable1600239931 struct{}

// ID return unique identifier for each migration. The prefix is unix time when this migration is created.
func (m CreateFcmServiceAccountKeysTable1600239931) ID(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServiceAccountKeysTable1600239931.FCMKeyID")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return fmt.Sprintf("%d_%s.sql", 1600239931, "create_fcm_service_account_keys_table")
}

// SequenceNumber return current time when the migration is created,
// this useful to see the current status of the migration.
func (m CreateFcmServiceAccountKeysTable1600239931) SequenceNumber(ctx context.Context) int {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServiceAccountKeysTable1600239931.SequenceNumber")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return 1600239931
}

// Up return sql migration for sync database
func (m CreateFcmServiceAccountKeysTable1600239931) Up(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServiceAccountKeysTable1600239931.Up")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `
CREATE TABLE IF NOT EXISTS fcm_service_account_keys (
	id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
	app_id UUID NOT NULL,
	service_account_key JSONB NOT NULL DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- for faster query WHERE app_client_id = ? ORDER BY created_at DESC
CREATE INDEX idx_fcm_service_account_keys_app_id_created_at ON fcm_service_account_keys (app_id, created_at DESC);
`
	return
}

// Down return sql migration for rollback database
func (m CreateFcmServiceAccountKeysTable1600239931) Down(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateFcmServiceAccountKeysTable1600239931.Down")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `DROP TABLE IF EXISTS fcm_service_account_keys;`
	return
}
