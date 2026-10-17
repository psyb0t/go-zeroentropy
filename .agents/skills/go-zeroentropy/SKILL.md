---
name: go-zeroentropy
description: "Use the unofficial ZeroEntropy Go client when writing Go code for collections, documents, retrieval queries, reranking, embeddings, indexing status, usage or administration, or when a project already imports github.com/psyb0t/go-zeroentropy. Package zeroentropy provides typed methods generated from the full public OpenAPI spec."
homepage: https://github.com/psyb0t/go-zeroentropy
user-invocable: true
metadata:
  openclaw:
    emoji: "🔎"
    requires:
      bins:
        - go
      env:
        - ZEROENTROPY_API_KEY
permissions:
  filesystem: "Guides edits to Go source in the current project. The library reads and writes no files."
  shell: "The Go toolchain (`go get`, `go build`, `go test`) to add and compile the dependency."
  network: "`go get` fetches the module from the Go module proxy. Runtime requests use api.zeroentropy.dev or eu-api.zeroentropy.dev."
---

# go-zeroentropy

Unofficial, typed Go client for the ZeroEntropy API, in package `zeroentropy` at the module root.

Read [references/setup.md](references/setup.md) before choosing methods or writing requests. It lists authentication, all operations, pagination and response pitfalls.

## Security and safety

- Read the API key from `ZEROENTROPY_API_KEY`. The request editor sets `Authorization: Bearer <key>`. Never hardcode, commit or log the key, authorization header or document download tokens.
- Use `https://api.zeroentropy.dev/v1` or `https://eu-api.zeroentropy.dev/v1`. Select the region that holds the user's data. Keep `/v1` in the base URL.
- Creating or deleting collections, adding, updating or deleting documents, and changing organizations or API keys act on the account. Call them only when the user requested that action. Queries and model calls can incur usage charges.
- Use an `httptest` server for tests. Do not send test documents or keys to the live service.

## When to use

- Go code needs ZeroEntropy retrieval, reranking or embeddings.
- Go code needs collections, documents, indexing status, usage or administration.
- The project already imports `github.com/psyb0t/go-zeroentropy`.

## When not to use

- The code is not Go. This module ships only a Go package.
- The task requires a client maintained by ZeroEntropy. This client is unofficial and unaffiliated.
- The task changes the generated client itself. Read the repository's `AGENTS.md` instead.

## Usage

Install `github.com/psyb0t/go-zeroentropy`, create a `ClientWithResponses` with `WithRequestEditorFn`, and call `GetCollectionListCollectionsGetCollectionListPostWithResponse` with `EmptyRequest{}` to list collections. The complete example is in the repository's `example_test.go` and README.

The error return covers transport and decoding failures. HTTP errors require inspecting `StatusCode()` and `Body`; a nil Go error does not prove success. Read `JSON200.CollectionNames` only after checking `JSON200` for nil. Some successful calls use another status, such as 201 for a newly added document.

Method names retain the operation IDs, including their resource and HTTP verb suffixes. Most reads use POST. Scores and numeric embeddings use `float64`; nullable or optional fields can be pointers, while required fields can be values. Check the generated type before constructing a request.

Finish by compiling and testing the consumer's code, checking the selected host and ensuring every account mutation matches the user's request.
