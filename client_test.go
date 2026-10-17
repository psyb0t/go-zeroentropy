package zeroentropy_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-zeroentropy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAuthorization = "Bearer EXAMPLE-DO-NOT-USE"
	testCollection    = "manuals"
	testDocumentPath  = "guide.txt"
	testContentType   = "application/json"
	testQuery         = "installation"
	testScore         = 0.987654321012345
	testDownloadToken = "EXAMPLE-DO-NOT-USE+/=&? token"
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *zeroentropy.ClientWithResponses {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	auth := zeroentropy.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", testAuthorization)

			return nil
		},
	)

	client, err := zeroentropy.NewClientWithResponses(srv.URL+"/v1", auth)
	require.NoError(t, err)

	return client
}

func assertRequest(t *testing.T, r *http.Request, path string) {
	t.Helper()

	assert.Equal(t, http.MethodPost, r.Method)
	assert.Equal(t, path, r.URL.Path)
	assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
	assert.Equal(t, testContentType, r.Header.Get("Content-Type"))
}

func writeResponse(
	t *testing.T,
	w http.ResponseWriter,
	status int,
	body string,
) {
	t.Helper()

	w.Header().Set("Content-Type", testContentType)
	w.WriteHeader(status)
	_, err := io.WriteString(w, body)
	assert.NoError(t, err)
}

func TestGetCollectionList(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "/v1/collections/get-collection-list")

		var body zeroentropy.EmptyRequest

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		writeResponse(t, w, http.StatusOK,
			`{"collection_names":["manuals","notes"]}`)
	})

	res, err := client.
		GetCollectionListCollectionsGetCollectionListPostWithResponse(
			context.Background(),
			zeroentropy.EmptyRequest{},
		)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	assert.Equal(t, []string{testCollection, "notes"},
		res.JSON200.CollectionNames)
}

func TestTopDocumentsWithNullQuery(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "/v1/queries/top-documents")

		var body map[string]json.RawMessage

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Contains(t, body, "query")
		assert.JSONEq(t, `null`, string(body["query"]))
		assert.JSONEq(t, `"manuals"`, string(body["collection_name"]))
		writeResponse(t, w, http.StatusOK, `{
			"results":[{
				"path":"guide.txt","score":0.987654321012345,
				"metadata":null,"file_url":"https://example.invalid/file"
			}]
		}`)
	})

	res, err := client.TopDocumentsQueriesTopDocumentsPostWithResponse(
		context.Background(),
		zeroentropy.TopDocumentsRequest{CollectionName: testCollection, K: 1},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.Len(t, res.JSON200.Results, 1)

	result := res.JSON200.Results[0]
	assert.Equal(t, testDocumentPath, result.Path)
	assert.InDelta(t, testScore, result.Score, 0)
	assert.Nil(t, result.Metadata)
}

func TestTopSnippetsPageSpan(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "/v1/queries/top-snippets")

		var body zeroentropy.TopSnippetsRequest

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, testCollection, body.CollectionName)
		assert.Equal(t, testQuery, body.Query)
		writeResponse(t, w, http.StatusOK, `{
			"results":[{
				"path":"guide.txt","start_index":0,"end_index":12,
				"page_span":[0,2],"content":"installation",
				"score":0.987654321012345
			}],"document_results":[]
		}`)
	})

	res, err := client.TopSnippetsQueriesTopSnippetsPostWithResponse(
		context.Background(),
		zeroentropy.TopSnippetsRequest{
			CollectionName: testCollection,
			Query:          testQuery,
			K:              1,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.Len(t, res.JSON200.Results, 1)

	result := res.JSON200.Results[0]
	assert.Equal(t, []int{0, 2}, result.PageSpan)
	assert.InDelta(t, testScore, result.Score, 0)
	assert.Equal(t, testQuery, result.Content)
}

func TestAddTextDocument(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "/v1/documents/add-document")

		var body zeroentropy.AddDocumentRequest

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, testCollection, body.CollectionName)
		assert.Equal(t, testDocumentPath, body.Path)

		content, err := body.Content.AsAPITextDocument()
		assert.NoError(t, err)
		assert.Equal(t, zeroentropy.APITextDocumentTypeText, content.Type)
		assert.Equal(t, testQuery, content.Text)
		writeResponse(t, w, http.StatusOK, `{"message":"Success!"}`)
	})

	var content zeroentropy.AddDocumentRequest_Content

	require.NoError(t, content.FromAPITextDocument(zeroentropy.APITextDocument{
		Type: zeroentropy.APITextDocumentTypeText,
		Text: testQuery,
	}))

	res, err := client.AddDocumentDocumentsAddDocumentPostWithResponse(
		context.Background(),
		zeroentropy.AddDocumentRequest{
			CollectionName: testCollection,
			Path:           testDocumentPath,
			Content:        content,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.NotNil(t, res.JSON200.Message)
	assert.Equal(t, "Success!", *res.JSON200.Message)
}

func TestHTTPErrorResponses(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status int
		body   string
	}{
		{
			name:   "unauthorized",
			status: http.StatusUnauthorized,
			body:   `{"detail":"Invalid API key"}`,
		},
		{
			name:   "validation",
			status: http.StatusUnprocessableEntity,
			body: `{"detail":[{
				"loc":["body",0],"msg":"Field required","type":"missing"
			}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t,
				func(w http.ResponseWriter, r *http.Request) {
					assertRequest(t, r,
						"/v1/collections/get-collection-list")
					writeResponse(t, w, tc.status, tc.body)
				})

			res, err := client.
				GetCollectionListCollectionsGetCollectionListPostWithResponse(
					context.Background(),
					zeroentropy.EmptyRequest{},
				)
			require.NoError(t, err)
			assert.Equal(t, tc.status, res.StatusCode())
			assert.JSONEq(t, tc.body, string(res.Body))
			assert.Nil(t, res.JSON200)

			if tc.status == http.StatusUnprocessableEntity {
				require.NotNil(t, res.JSON422)
				require.NotNil(t, res.JSON422.Detail)
				require.Len(t, *res.JSON422.Detail, 1)
				assert.Equal(t, "Field required",
					(*res.JSON422.Detail)[0].Msg)
			}
		})
	}
}

func TestMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeResponse(t, w, http.StatusOK, `{"collection_names":!}`)
	})

	res, err := client.
		GetCollectionListCollectionsGetCollectionListPostWithResponse(
			context.Background(),
			zeroentropy.EmptyRequest{},
		)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
	assert.Nil(t, res)
}

func TestCanceledContext(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("canceled request reached the server")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := client.
		GetCollectionListCollectionsGetCollectionListPostWithResponse(
			ctx,
			zeroentropy.EmptyRequest{},
		)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, res)
}

func TestDocumentListPagination(t *testing.T) {
	t.Parallel()

	const pageSize = 2

	paths := []string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt"}
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "/v1/documents/get-document-info-list")

		var body zeroentropy.GetDocumentInfoListRequest

		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			return
		}

		assert.Equal(t, testCollection, body.CollectionName)
		assert.Equal(t, new(pageSize), body.Limit)

		documents := make([]zeroentropy.DocumentMetadataResponse, 0)

		for _, path := range paths {
			if body.PathGt != nil && path <= *body.PathGt {
				continue
			}

			documents = append(documents, zeroentropy.DocumentMetadataResponse{
				CollectionName: testCollection,
				Path:           path,
				IndexStatus:    zeroentropy.DocumentIndexStatusIndexed,
				Metadata:       zeroentropy.DocumentMetadataJson{},
			})
			if len(documents) == pageSize {
				break
			}
		}

		bodyBytes, err := json.Marshal(zeroentropy.GetDocumentInfoListResponse{
			Documents: documents,
		})
		assert.NoError(t, err)
		writeResponse(t, w, http.StatusOK, string(bodyBytes))
	})

	var cursor *string

	actualPaths := make([]string, 0, len(paths))

	for _, count := range []int{pageSize, pageSize, 1, 0} {
		res, err := client.
			GetDocumentInfoListDocumentsGetDocumentInfoListPostWithResponse(
				context.Background(),
				zeroentropy.GetDocumentInfoListRequest{
					CollectionName: testCollection,
					Limit:          new(pageSize),
					PathGt:         cursor,
				},
			)
		require.NoError(t, err)
		require.NotNil(t, res.JSON200)
		require.Len(t, res.JSON200.Documents, count)

		for _, document := range res.JSON200.Documents {
			actualPaths = append(actualPaths, document.Path)
			cursor = new(document.Path)
			assert.Nil(t, document.NumPages)
		}
	}

	assert.Equal(t, paths, actualPaths)
}

func TestRerankNullableOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		topN *int
		want string
	}{
		{
			name: "omitted",
			want: `{"documents":["text"],"model":"model-X","query":"query"}`,
		},
		{
			name: "set",
			topN: new(1),
			want: `{
				"documents":["text"],"model":"model-X","query":"query",
				"top_n":1
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(zeroentropy.RerankRequest{
				Documents: []string{"text"},
				Model:     "model-X",
				Query:     "query",
				TopN:      tc.topN,
			})
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(body))
		})
	}
}

func TestDownloadDocumentJSON(t *testing.T) {
	t.Parallel()

	const responseBody = `{"size":9007199254740993,"ok":true}`

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/documents/get-document", r.URL.Path)
		assert.Equal(t, testDownloadToken, r.URL.Query().Get("token"))
		assert.Len(t, r.URL.Query(), 1)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		writeResponse(t, w, http.StatusOK, responseBody)
	})

	res, err := client.GetDocumentDocumentsGetDocumentGetWithResponse(
		context.Background(),
		&zeroentropy.GetDocumentDocumentsGetDocumentGetParams{
			Token: testDownloadToken,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)

	assert.Equal(t, json.RawMessage(responseBody), *res.JSON200)
	assert.Equal(t, []byte(responseBody), res.Body)
}

func TestDownloadPageImage(t *testing.T) {
	t.Parallel()

	const (
		imageContentType = "image/jpeg"
		imageBody        = "\xff\xd8EXAMPLE\xff\xd9"
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/documents/get-page-image", r.URL.Path)
		assert.Equal(t, testDownloadToken, r.URL.Query().Get("token"))
		assert.Len(t, r.URL.Query(), 1)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", imageContentType)
		_, err := io.WriteString(w, imageBody)
		assert.NoError(t, err)
	})

	res, err := client.GetPageImageDocumentsGetPageImageGetWithResponse(
		context.Background(),
		&zeroentropy.GetPageImageDocumentsGetPageImageGetParams{
			Token: testDownloadToken,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, imageContentType, res.ContentType())
	assert.Equal(t, []byte(imageBody), res.Body)
	assert.Nil(t, res.JSON200)
}
