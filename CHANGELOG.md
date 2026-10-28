# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.2.0, 2026-10-28

- The client is generated from the spec ZeroEntropy's API publishes at `https://api.zeroentropy.dev/v1/openapi.json`. The Stainless-hosted copy of the spec it used before is no longer publicly readable, so scheduled regeneration had stopped.
- Added three operations: get organization, update organization and embed usage. Nothing was removed.
- `scripts/patch-spec.sh` fails when a download response schema it rewrites to `json.RawMessage` is no longer empty, instead of skipping the change silently.
- Updated Semgrep in the development image to 1.182.1 and refreshed its locked dependencies. certifi 2026.10.23 drops the emSign root CAs whose intermediate keys were compromised (GHSA-6q48-6c7w-9gw2).
- Dependabot no longer opens pull requests for `mcp` or the OpenTelemetry packages in `requirements-semgrep.txt`. Semgrep pins their versions itself, so bumping one of them on its own cannot install.
- Updated kin-openapi, which the code generator uses, from 0.142.0 to 0.149.0. The newer version fixes several panics on malformed specs reported in GitHub security advisories.
- The security scan merges its two gitleaks runs into one SARIF run. GitHub code scanning rejects a report with two runs from the same tool, so no scan results had been uploaded.

## v0.1.2, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.
- Raised the minimum Go version in `go.mod` from 1.26.9 to 1.26.10, which fixes the `crypto/x509` vulnerability GO-2026-6809 found by govulncheck.
- The issue-pull workflow runs at hours derived from the repository name instead of the same four hours as every other repository.

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
