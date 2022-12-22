package assets

import (
	"embed"
	_ "embed"
)

const ServiceName = "ngendika"

var (
	//go:embed swaggerui/*
	SwaggerUI embed.FS
)
