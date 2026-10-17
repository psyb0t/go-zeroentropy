package zeroentropy_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-zeroentropy"
)

//nolint:testableexamples // Needs real credentials, compile check only.
func Example() {
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
