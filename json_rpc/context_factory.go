package json_rpc

import (
	"context"
	"net/http"
)

type InitialContextFactory func(ctx context.Context) (context.Context, context.CancelFunc)
type ContextBuilder func(ctx context.Context, request *RequestBase, rawHttpRequest *http.Request) (context.Context, ServerError)

func NewCompositeContextBuilder(builders []ContextBuilder) ContextBuilder {
	return func(initialCtx context.Context, request *RequestBase, rawHttpRequest *http.Request) (ctx context.Context, err ServerError) {
		ctx = initialCtx
		for i := range builders {
			ctx, err = builders[i](ctx, request, rawHttpRequest)
			if err != nil {
				return
			}
		}
		return
	}
}
