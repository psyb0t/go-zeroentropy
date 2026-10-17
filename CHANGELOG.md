# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `zeroentropy`, a REST client generated with oapi-codegen from the ZeroEntropy API spec, version 0.1.0, OpenAPI 3.1.0. It covers all 27 operations: collections, documents, retrieval queries, reranking, embeddings, indexing status, usage and administration.
- Source: [the Stainless-hosted OpenAPI document](https://storage.googleapis.com/stainless-sdk-openapi-specs/zeroentropy%2Fzeroentropy-5ff1b24060b908d169910a7366213eeb423b52d14b03cc16902923eb0ffac71a.yml), resolved from the vendor Python SDK's `.stats.yml`.
- Number fields are `float64` instead of oapi-codegen's default `float32`.
- The two download GETs' empty JSON success schemas use `json.RawMessage` to preserve arbitrary JSON bytes. Binary content remains available through raw responses or `Body`.
- `make generate` downloads the SDK metadata and current spec, downconverts nullable unions, constants and homogeneous tuples for the generator, and regenerates the client. The spec is not committed.
