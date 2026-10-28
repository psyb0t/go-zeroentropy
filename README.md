# go-zeroentropy

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-zeroentropy.svg)](https://pkg.go.dev/github.com/psyb0t/go-zeroentropy)
[![CI](https://github.com/psyb0t/go-zeroentropy/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-zeroentropy/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-zeroentropy/badges/version.svg)](https://github.com/psyb0t/go-zeroentropy/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-zeroentropy/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-zeroentropy/badges/importers.svg)](https://github.com/psyb0t/go-zeroentropy/blob/badges/importers.md)

Unofficial Go client for the [ZeroEntropy API](https://docs.zeroentropy.dev), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the public OpenAPI document the ZeroEntropy API serves. I am not affiliated with ZeroEntropy.

The package is `zeroentropy`, in the repository root. Everything in `client.gen.go` is generated. Don't edit it by hand, regenerate it.

## Install

```sh
go get github.com/psyb0t/go-zeroentropy
```

## Usage

Every endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200` and the other status fields. The full API, with the spec's descriptions as doc comments, is in the [Go reference](https://pkg.go.dev/github.com/psyb0t/go-zeroentropy) and in `go doc github.com/psyb0t/go-zeroentropy`.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-zeroentropy"
)

func main() {
	auth := zeroentropy.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"Authorization",
				"Bearer "+os.Getenv("ZEROENTROPY_API_KEY"),
			)

			return nil
		},
	)

	client, err := zeroentropy.NewClientWithResponses(
		"https://api.zeroentropy.dev/v1",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.
		GetCollectionListCollectionsGetCollectionListPostWithResponse(
			context.Background(),
			zeroentropy.EmptyRequest{},
		)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles. Check `StatusCode()` and that `JSON200` is non-nil before using a response; HTTP errors can return a nil Go error.

Method names retain the spec's operation IDs. The full spec covers 30 operations: collections, documents, document/page/snippet retrieval, reranking, embeddings, indexing status, usage and administration. Most reads use POST with a JSON body.

## Auth

The spec's `HTTPBearer` security scheme uses `Authorization: Bearer <your-api-key>`. Get a key from the [dashboard](https://dashboard.zeroentropy.dev/) and set `ZEROENTROPY_API_KEY` for the example. `WithRequestEditorFn` attaches the header; the client does not load credentials itself.

Choose the base URL for the region holding your data. Include `/v1`:

| Region | Base URL |
| --- | --- |
| Default | `https://api.zeroentropy.dev/v1` |
| EU datacenters | `https://eu-api.zeroentropy.dev/v1` |

The document and page-image download GETs use a `token` query parameter. Their descriptions specify binary content, while their response schemas advertise JSON. Use the raw client methods for binary downloads and keep token URLs private.

## Regenerating

The spec is not committed. `make generate` downloads the current OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps, so `make generate` is just `go generate ./...`:

- `scripts/fetch-spec.sh` downloads the spec the API serves at `https://api.zeroentropy.dev/v1/openapi.json` to `api/openapi.json`.
- `scripts/patch-spec.sh` writes `api/openapi.fixed.json`. It downconverts OpenAPI 3.1 nullable unions to nullable schemas, constants to singleton enums, and the homogeneous page-span tuple to array items with its length bounds. Nullable references use `allOf` so the loader retains nullability. The two download GETs' empty JSON success schemas map to `json.RawMessage`, preserving JSON bytes without decoding arbitrary objects. The resulting document declares OpenAPI 3.0.3.
- oapi-codegen runs with `config.yaml`. The config maps `number` to `float64` to retain precision in scores and embeddings. All operations are generated, including those marked hidden in the spec.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release.

## Agent integrations

The [skill](.agents/skills/go-zeroentropy) teaches a coding agent how to use go-zeroentropy when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-zeroentropy@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-zeroentropy@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The skill is published to ClawHub on every release:

```bash
openclaw skills install @psyb0t/go-zeroentropy
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

## License

[MIT](LICENSE). The generated code comes from ZeroEntropy's published API specification. The specification itself is not redistributed here and provides no license or terms-of-service metadata.
