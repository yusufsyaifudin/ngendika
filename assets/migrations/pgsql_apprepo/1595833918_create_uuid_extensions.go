package pgsql_apprepo

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

// CreateUuidExtensions1595833918 is struct to define a migration with ID 1595833918_create_uuid_extensions
type CreateUuidExtensions1595833918 struct{}

// ID return unique identifier for each migration. The prefix is unix time when this migration is created.
func (m CreateUuidExtensions1595833918) ID(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUuidExtensions1595833918.FCMKeyID")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return fmt.Sprintf("%d_%s.sql", 1595833918, "create_uuid_extensions")
}

// SequenceNumber return current time when the migration is created,
// this useful to see the current status of the migration.
func (m CreateUuidExtensions1595833918) SequenceNumber(ctx context.Context) int {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUuidExtensions1595833918.SequenceNumber")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return 1595833918
}

// Up return sql migration for sync database
func (m CreateUuidExtensions1595833918) Up(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUuidExtensions1595833918.Up")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	sql = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
	return
}

// Down return sql migration for rollback database
func (m CreateUuidExtensions1595833918) Down(ctx context.Context) (sql string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUuidExtensions1595833918.Down")
	defer func() {
		ctx.Done()
		span.Finish()
	}()

	return
}
