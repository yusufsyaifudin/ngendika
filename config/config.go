package config

// HTTPServer struct for HTTP Transport configuration
type HTTPServer struct {
	Port int `yaml:"port"`
}

// Transport is a configuration for Admin Transport: HTTP, gRPC or anything
type Transport struct {
	HTTP HTTPServer `yaml:"http"`
}

type GoSqlDb struct {
	Debug bool   `yaml:"debug"`
	DSN   string `yaml:"dsn"` // Data Source Name
}

type DatabaseResource struct {
	Disable bool   `yaml:"disable"`
	Driver  string `yaml:"driver"` // mysql, postgres, etc

	// per driver configuration
	Mysql    GoSqlDb `yaml:"mysql"`
	Postgres GoSqlDb `yaml:"postgres"`
}

// DatabaseResources redefine config
type DatabaseResources map[string]DatabaseResource

type MsgService struct {
	MaxParallel int `yaml:"maxParallel"`
}

// Config contains application config
type Config struct {
	Transport Transport `yaml:"transport"`

	DatabaseResources DatabaseResources `yaml:"databaseResources"`

	AppRepo struct {
		DBLabel string `yaml:"dbLabel"`
	} `yaml:"appRepo" flag:"appRepo"`

	FCMRepo struct {
		DBLabel string `yaml:"dbLabel"`
	} `yaml:"fcmRepo" flag:"fcmRepo"`

	MsgRepo struct {
		DBLabel string `yaml:"dbLabel"`
	} `yaml:"msgRepo" flag:"msgRepo"`

	MsgService MsgService `yaml:"msgService"`
}
