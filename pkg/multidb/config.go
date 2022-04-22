package multidb

type DBConnection struct {
	Disable bool   `yaml:"disable" flag:"disable"`
	Debug   bool   `yaml:"debug" flag:"debug"`
	Driver  string `yaml:"driver" flag:"driver"`
	DSN     string `yaml:"dsn" flag:"dsn"` // Data Source Name
}

type Database map[string]DBConnection
