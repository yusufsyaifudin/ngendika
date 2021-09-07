package config

// HTTPServer struct for HTTP Transport configuration
type HTTPServer struct {
	Port int `mapstructure:"port" flag:"port"`
}

// Transport is a configuration for Admin Transport: HTTP, gRPC or anything
type Transport struct {
	HTTP HTTPServer `yaml:"http" flag:"http"`
}

type DBConnection struct {
	Disable bool   `yaml:"disable" flag:"disable"`
	Debug   bool   `yaml:"debug" flag:"debug"`
	Driver  string `yaml:"driver" flag:"driver"`
	DSN     string `yaml:"dsn" flag:"dsn"`
}

type Database map[string]DBConnection

type RedisConnection struct {
	Mode       string   `yaml:"mode" validate:"required,oneof=single sentinel cluster"`
	Address    []string `yaml:"address" validate:"required,unique,dive,required"`
	Username   string   `yaml:"username" validate:"-"`
	Password   string   `yaml:"password" validate:"-"`
	DB         int      `yaml:"db" validate:"-"`
	MasterName string   `yaml:"master_name" validate:"required_if=Mode sentinel"`
}

type Redis map[string]RedisConnection

type Worker struct {
	QueueType       string `yaml:"queueType" flag:"queueType" validate:"required"`
	QueueIdentifier string `yaml:"queueIdentifier" flag:"queueIdentifier" validate:"required_if=QueueType redis"`
	Num             int    `yaml:"num" flag:"num" usage:"Number of worker"`
}

// Config contains application config
type Config struct {
	Transport Transport `yaml:"transport" flag:"transport"`
	Database  Database  `yaml:"database" flag:"database"`
	Redis     Redis     `yaml:"redis" flag:"redis"`

	AppRepo struct {
		Database string `yaml:"database"`
		Cache    bool   `yaml:"cache"`
	} `yaml:"appRepo" flag:"appRepo"`

	MsgRepo struct {
		Database string `yaml:"database"`
		Cache    bool   `yaml:"cache"`
	} `yaml:"msgRepo" flag:"msgRepo"`

	MsgService struct {
		QueueDisable    bool   `yaml:"queueDisable" flag:"queueDisable"`
		QueueType       string `yaml:"queueType" flag:"queueType" validate:"required"`
		QueueIdentifier string `yaml:"queueIdentifier" flag:"queueIdentifier" validate:"required_if=QueueType redis"`
	} `yaml:"msgService"`

	Worker Worker `yaml:"worker" flag:"worker"`
}
