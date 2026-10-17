// Package zeroentropy is a Go client for the ZeroEntropy retrieval API,
// generated with oapi-codegen from the OpenAPI document ZeroEntropy publishes.
package zeroentropy

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/patch-spec.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.fixed.json
