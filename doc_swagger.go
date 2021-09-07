/*
 * Copyright © 2021 Yusuf
 */

// @title ngendika
// @version 1.0
// @description Welcome to the ngendika HTTP API documentation. You will find documentation for all HTTP APIs here.

// @contact.name Yusuf
// @contact.url http://yusufsyaifudin.github.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// Package main ngendika
//
// Welcome to the ngendika HTTP API documentation. You will find documentation for all HTTP APIs here.
//
//     Schemes: http, https
//     Host: yusufsyaifudin.github.io
//     BasePath: /
//     Version: latest
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Extensions:
//     ---
//     x-request-id: string
//     x-forwarded-proto: string
//     ---
//
// swagger:meta
package main

// Error response
//
// Error responses are sent when an error (e.g. unauthorized, bad request, ...) occurred.
//
// swagger:model genericError
type genericError struct {
	// ErrorCode represents the error status code (00, 01, 02, 03, ...).
	//
	// example: 00
	ErrorCode string `json:"error_code"`

	// ErrorDescription contains further information on the nature of the error.
	//
	// example: Object with FCMKeyID 12345 does not exist
	ErrorDescription string `json:"error_description"`

	// Debug contains debug information. This is usually not available and has to be enabled.
	// For unhandled error, this contain system error that may leak some secret information (such as db name, etc).
	// Hence, this is encourage to disable in production.
	// For another error type, such as validation error, then it will show technical information.
	//
	// example: The database adapter was unable to find the element
	Debug string `json:"debug"`
}

// Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is
// typically 201.
//
// swagger:response emptyResponse
type emptyResponse struct{}
