package migration

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	migrate "github.com/rubenv/sql-migrate"
)

type SQLImmigrationConfig struct {
	Dialect        string    `validate:"required,oneof=mysql postgres"`
	DB             *sql.DB   `validate:"required"`
	MigrationTable string    `validate:"required"`
	Migrations     []Migrate `validate:"required,unique"`
}

type SQLImmigration struct {
	config SQLImmigrationConfig
	source migrate.MigrationSource
}

func (p SQLImmigration) Up() error {
	migrate.SetTable(p.config.MigrationTable)

	_, err := migrate.Exec(
		p.config.DB,
		p.config.Dialect,
		p.source,
		migrate.Up,
	)

	return err
}

func (p *SQLImmigration) Down() error {
	migrate.SetTable(p.config.MigrationTable)

	_, err := migrate.Exec(
		p.config.DB,
		p.config.Dialect,
		p.source,
		migrate.Down,
	)

	return err
}

func NewSQLImmigration(ctx context.Context, config SQLImmigrationConfig) (Immigration, error) {
	err := validator.New().Struct(config)
	if err != nil {
		return nil, err
	}

	mig := make([]*migrate.Migration, 0)
	for _, m := range config.Migrations {
		sqlUp, err := m.Up(ctx)
		if err != nil {
			return nil, err
		}

		sqlDown, err := m.Down(ctx)
		if err != nil {
			return nil, err
		}

		mig = append(mig,
			&migrate.Migration{
				Id:   m.ID(ctx),
				Up:   []string{sqlUp},
				Down: []string{sqlDown},
			})
	}

	m := &migrate.MemoryMigrationSource{
		Migrations: mig,
	}

	return &SQLImmigration{
		config: config,
		source: m,
	}, nil
}
