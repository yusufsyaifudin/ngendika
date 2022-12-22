package multidb

type Driver string

func (d Driver) String() string {
	return string(d)
}

const (
	Postgres Driver = "postgres"
)

type GoSqlDb struct {
	Debug bool
	DSN   string // Data Source Name
}

type DatabaseResource struct {
	Disable bool
	Driver  Driver // postgres, etc

	// per driver configuration
	Postgres GoSqlDb
}

type DatabaseResources map[string]DatabaseResource
