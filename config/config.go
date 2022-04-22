package config

import "github.com/yusufsyaifudin/ngendika/pkg/multidb"

// HTTPServer struct for HTTP Transport configuration
type HTTPServer struct {
	Port int `mapstructure:"port" flag:"port"`
}

// Transport is a configuration for Admin Transport: HTTP, gRPC or anything
type Transport struct {
	HTTP HTTPServer `yaml:"http" flag:"http"`
}

// Database redefine config
type Database multidb.Database

type MsgService struct {
	MaxParallel int `yaml:"maxParallel" flag:"maxParallel"`
}

// Config contains application config
type Config struct {
	Transport Transport `yaml:"transport" flag:"transport"`
	Database  Database  `yaml:"database" flag:"database"`

	AppRepo struct {
		Database string `yaml:"database"`
	} `yaml:"appRepo" flag:"appRepo"`

	FCMRepo struct {
		Database string `yaml:"database"`
	} `yaml:"fcmRepo" flag:"fcmRepo"`

	MsgRepo struct {
		Database string `yaml:"database"`
	} `yaml:"msgRepo" flag:"msgRepo"`

	MsgService MsgService `yaml:"msgService"`
}
