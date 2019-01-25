package json_rpc

import (
	"context"
	"net/http"
)

type ContextFactory func(request *RequestBase, rawHttpRequest *http.Request) (context.Context, ServerError)
type ContextBuilder func(ctx context.Context, request *RequestBase, rawHttpRequest *http.Request) (context.Context, ServerError)

func NewCompositeContextFactory(builders []ContextBuilder) ContextFactory {
	return func(request *RequestBase, rawHttpRequest *http.Request) (ctx context.Context, err ServerError) {
		ctx = context.Background()

		for i := range builders {
			ctx, err = builders[i](ctx, request, rawHttpRequest)
			if err != nil {
				return
			}
		}
		return
	}
}
