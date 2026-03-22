// Package api provides the embedded OpenAPI specification.
package api

import _ "embed"

// OpenAPISpec contains the raw bytes of the OpenAPI 3.1.0 specification (openapi.yaml).
//
//go:embed openapi.yaml
var OpenAPISpec []byte
