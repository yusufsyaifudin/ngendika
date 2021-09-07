package pgsql_apprepo

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

// CreateAppsTable1595833942 is struct to define a migration with ID 1595833942_create_apps_table
type CreateAppsTable1595833942 struct{}

// ID return unique identifier for each migration. The prefix is unix time when this migration is created.
func (m CreateAppsTable1595833942) ID(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateAppsTable1595833942.FCMKeyID")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return fmt.Sprintf("%d_%s.sql", 1595833942, "create_apps_table")
}

// SequenceNumber return current time when the migration is created,
// this useful to see the current status of the migration.
func (m CreateAppsTable1595833942) SequenceNumber(ctx context.Context) int {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateAppsTable1595833942.SequenceNumber")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return 1595833942
}

// Up return sql migration for sync database
func (m CreateAppsTable1595833942) Up(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateAppsTable1595833942.Up")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `
CREATE TABLE IF NOT EXISTS apps (
	client_id VARCHAR NOT NULL PRIMARY KEY,
	name VARCHAR NOT NULL DEFAULT '',
	enabled BOOL NOT NULL DEFAULT true,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX unique_idx_apps_client_id ON apps (LOWER(client_id));
`

	return
}

// Down return sql migration for rollback database
func (m CreateAppsTable1595833942) Down(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateAppsTable1595833942.Down")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `DROP TABLE IF EXISTS apps;`
	return
}
