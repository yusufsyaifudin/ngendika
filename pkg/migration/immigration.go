package migration

import "context"

// Immigration is an interface to run migration Up or Down.
// Up means running all queries to latest version,
// Down means rollback to Down queries.
type Immigration interface {
	Up() error
	Down() error
}

// Migrate is a migration data to run.
type Migrate interface {
	// ID return unique identifier for each migration. The prefix must be number
	ID(ctx context.Context) string

	// SequenceNumber must be unique, this useful to see the current status of the migration.
	SequenceNumber(ctx context.Context) int

	// Up return sql migration for sync database
	Up(ctx context.Context) (sql string, err error)

	// Down return sql migration for rollback database
	Down(ctx context.Context) (sql string, err error)
}
