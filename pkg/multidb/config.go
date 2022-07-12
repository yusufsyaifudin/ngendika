package multidb

type Driver string

func (d Driver) String() string {
	return string(d)
}

const (
	Mysql    Driver = "mysql"
	Postgres Driver = "postgres"
)

type GoSqlDb struct {
	Debug bool
	DSN   string // Data Source Name
}

type DatabaseResource struct {
	Disable bool
	Driver  Driver // mysql, postgres, etc

	// per driver configuration
	Mysql    GoSqlDb
	Postgres GoSqlDb
}

type DatabaseResources map[string]DatabaseResource
