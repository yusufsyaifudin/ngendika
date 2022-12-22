package container

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// ConfigHTTPServer struct for HTTP ConfigTransport configuration
type ConfigHTTPServer struct {
	Port int `yaml:"port"`
}

// ConfigTransport is a configuration for Admin ConfigTransport: HTTP, gRPC or anything
type ConfigTransport struct {
	HTTP ConfigHTTPServer `yaml:"http"`
}

type ConfigGoSqlDb struct {
	Debug bool   `yaml:"debug"`
	DSN   string `yaml:"dsn"` // Data Source Name
}

type ConfigDatabaseResource struct {
	Disable bool   `yaml:"disable"`
	Driver  string `yaml:"driver"` // mysql, postgres, etc

	// per driver configuration
	// Mysql    ConfigGoSqlDb `yaml:"mysql"` TODO: only example if we want to add another driver
	Postgres ConfigGoSqlDb `yaml:"postgres"`
}

// ConfigDatabaseResources redefine config
type ConfigDatabaseResources map[string]ConfigDatabaseResource

type ConfigServiceApp struct {
	DBLabel string `yaml:"dbLabel"`
}

type ConfigServicePushProvider struct {
	DBLabel string `yaml:"dbLabel"`
}

type ConfigServiceMessaging struct {
	DBLabel     string `yaml:"dbLabel"`
	MaxBuffer   int    `yaml:"maxBuffer"`
	MaxParallel int    `yaml:"maxParallel"`
}

type ConfigServices struct {
	App             ConfigServiceApp          `yaml:"app"`
	ServiceProvider ConfigServicePushProvider `yaml:"serviceProvider"`
	Messaging       ConfigServiceMessaging    `yaml:"messaging"`
}

// Config contains application config
type Config struct {
	Transport         ConfigTransport         `yaml:"transport"`
	DatabaseResources ConfigDatabaseResources `yaml:"databaseResources"`
	Services          ConfigServices          `yaml:"services"`
}

// LoadConfig need config file name and pointer to struct to hold the configuration value.
// It only supports YAML file content.
func LoadConfig() (cfg Config, err error) {
	const configFileName = "config.yml"
	fileContent, err := os.ReadFile(configFileName)
	if err != nil {
		err = fmt.Errorf("error read file config %s: %w", configFileName, err)
		return
	}

	dec := yaml.NewDecoder(bytes.NewReader(fileContent))
	dec.KnownFields(false)
	err = dec.Decode(&cfg)
	return
}
