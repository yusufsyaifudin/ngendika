package migrate

import (
	"fmt"
	"strings"

	"github.com/satori/uuid"
	"github.com/spf13/cobra"
	"github.com/yusufsyaifudin/ngendika/assets/migrations/pgsql_apprepo"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/container"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/migration"
	"gopkg.in/yaml.v2"
)

func Execute() *cobra.Command {
	var migrate = &cobra.Command{
		Use:   "migrate",
		Short: "Will migrate database",
		Long:  "",
	}

	// add command "ngendika -c config.yaml migrate appRepo up"
	migrate.AddCommand(appRepo())

	return migrate
}

func appRepo() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "appRepo",
		Short: "Will migrate database to save application and it's data (FCM, APNS cert, etc).",
		Long:  "appRepo will save application state (enable/disable), FCM Service Account Key and Server Key, APNS certificates, and all value related to the Client's data.",
	}

	cmd.AddCommand(appRepoMig("up"))
	cmd.AddCommand(appRepoMig("down"))
	cmd.AddCommand(appRepoMig("print"))

	return cmd
}

func appRepoMig(direction string) *cobra.Command {
	switch strings.ToLower(direction) {
	case "up":
		return &cobra.Command{
			Use:   "up",
			Short: "Sync up all migration.",
			RunE:  appRepoMigration("up"),
		}
	case "down":
		return &cobra.Command{
			Use:   "down",
			Short: "Reset all migration.",
			RunE:  appRepoMigration("down"),
		}
	case "print":
		return &cobra.Command{
			Use:   "print",
			Short: "Print all migration.",
			RunE:  appRepoMigration("print"),
		}
	default:
		return &cobra.Command{
			Use:   "down",
			Short: "Reset all migration.",
			RunE:  appRepoMigration(direction),
		}
	}
}

// appRepoMigration --
func appRepoMigration(direction string) func(*cobra.Command, []string) error {

	return func(cmd *cobra.Command, args []string) error {
		ctx := logger.Inject(cmd.Context(), logger.Tracer{
			RemoteAddr: "system",
			AppTraceID: uuid.NewV4().String(),
		})

		conf := config.Config{}
		zapLog, err := config.Setup(cmd, args, &conf)
		if err != nil {
			return fmt.Errorf("loading config error: %w", err)
		}

		defer func() {
			if err != nil {
				str, _ := yaml.Marshal(conf)
				fmt.Println(string(str))
			}
		}()

		// set global logger
		logger.SetGlobalLogger(logger.NewZap(zapLog))

		// empty migration
		var migrations = make([]migration.Migrate, 0)

		const (
			migrationTable = "migration_records_app_repo"
		)

		dbSQL, closer, err := container.SQLConnection(ctx, conf.Database)
		if err != nil {
			return err
		}

		defer func() {
			if _err := container.CloseAllSQLConnection(ctx, closer); _err != nil {
				logger.Error(ctx, "error close db", logger.KV("error", _err))
				return
			}
		}()

		appRepoDB, ok := conf.Database[conf.AppRepo.Database]
		if !ok {
			err = fmt.Errorf("prepare app repository error: unknown database key %s", conf.AppRepo.Database)
			return err
		}

		appRepoSQL, ok := dbSQL[conf.AppRepo.Database]
		if !ok {
			return fmt.Errorf("unknown database for %s", conf.AppRepo.Database)
		}

		switch appRepoDB.Driver {
		case "postgres":
			migrations = []migration.Migrate{
				new(pgsql_apprepo.CreateUuidExtensions1595833918),
				new(pgsql_apprepo.CreateAppsTable1595833942),
				new(pgsql_apprepo.CreateFcmServiceAccountKeysTable1600239931),
				new(pgsql_apprepo.CreateFcmServerKeysTable1624008564),
			}

		case "mongo": // TODO, for now it just example
			logger.Info(ctx, "mongo does not need to migrate")
			return nil

		default:
			return fmt.Errorf("unknwon dialect %s", appRepoDB.Driver)
		}

		err = appRepoSQL.Ping()
		if err != nil {
			return fmt.Errorf("ping db error: %w", err)
		}

		logger.Info(ctx, "trying to migrate")
		mig, err := migration.NewSQLImmigration(ctx, migration.SQLImmigrationConfig{
			Dialect:        appRepoDB.Driver,
			DB:             appRepoSQL.DB,
			MigrationTable: migrationTable,
			Migrations:     migrations,
		})

		if err != nil {
			return fmt.Errorf("prepare immigration error: %w", err)
		}

		switch direction {
		case "up":
			err = mig.Up()
		case "down":
			err = mig.Down()
		case "print":
			for _, mig := range migrations {
				fmt.Println(mig.ID(ctx))
				fmt.Println()
				fmt.Println(`
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied`)
				up, _ := mig.Up(ctx)
				fmt.Println(up)
				fmt.Println()
				fmt.Println(`
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back`)
				down, _ := mig.Down(ctx)
				fmt.Println(down)
			}

		default:
			return fmt.Errorf("unknown sub command direction: '%s'", direction)
		}

		if err != nil {
			return fmt.Errorf("query db error: %w", err)
		}

		logger.Info(ctx, "success migrate")
		return nil
	}
}
