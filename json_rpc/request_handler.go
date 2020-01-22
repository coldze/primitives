package json_rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/coldze/primitives/custom_error"
)

const (
	JSON_RPC_VERSION = "2.0"
)

type RequestInfo struct {
	Headers http.Header
	Cookies []*http.Cookie
	Data    interface{}
}

type ResponseInfo struct {
	Headers http.Header
	Data    interface{}
}

type RequestHandler func(ctx context.Context, request *RequestInfo) (*ResponseInfo, ServerError)
type RequestParamsFactory func() interface{}
type HeadersFromContext func(ctx context.Context) http.Header

type HandlingInfo struct {
	Handle         RequestHandler
	NewParams      RequestParamsFactory
	ComposeContext ContextBuilder
	GetHeaders     HeadersFromContext
}

type UnknownErrorData struct {
	CallStack     string      `json:"call_stack"`
	OriginalError interface{} `json:"original_error"`
}

func applyHeaders(w http.Header, headers http.Header) {
	for k, hs := range headers {
		for i := range hs {
			w.Add(k, hs[i])
		}
	}
}

func dummyContextFactory(ctx context.Context, request *RequestBase, r *http.Request) (context.Context, ServerError) {
	return ctx, nil
}

func dummyContextExpert(ctx context.Context) http.Header {
	return nil
}

type RawRequestHandler func(ctx context.Context, r *http.Request) (*ResponseInfo, string)

func NewJsonRPCHandle(handlers map[string]HandlingInfo, getDecoder func(data io.Reader) *json.Decoder) RawRequestHandler {
	methodHandlers := map[string]HandlingInfo{}
	for k, v := range handlers {
		updated := v
		if updated.ComposeContext == nil {
			updated.ComposeContext = dummyContextFactory
		}
		if updated.GetHeaders == nil {
			updated.GetHeaders = dummyContextExpert
		}
		methodHandlers[k] = updated
	}
	return func(srcCtx context.Context, r *http.Request) (*ResponseInfo, string) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ThrowError(0, 0, "Failed to read request body.", err)
		}
		incomingRequest := RequestBase{}
		dec := getDecoder(bytes.NewReader(data))
		err = dec.Decode(&incomingRequest)
		if err != nil {
			ThrowError(0, 1, "Failed to parse request body.", err)
		}

		if incomingRequest.Version != JSON_RPC_VERSION {
			ThrowError(0, 2, "Unsupported JSON RPC version", errors.New("Expected version = 2.0"))
		}

		handler, ok := methodHandlers[incomingRequest.Method]
		if !ok {
			ThrowError(0, 3, "Unsupported method: "+incomingRequest.Method, nil)
		}

		resHeaders := http.Header{}
		ctx, composeErr := handler.ComposeContext(srcCtx, &incomingRequest, r)
		if ctx != nil {
			applyHeaders(resHeaders, handler.GetHeaders(ctx))
		}
		if composeErr != nil {
			panic(composeErr)
		}

		params := RequestParams{
			Params: handler.NewParams(),
		}
		dec = getDecoder(bytes.NewReader(data))
		err = dec.Decode(&params)
		if err != nil {
			ThrowError(0, 3, "Failed to prepare arguments for handler,", err)
		}
		handlerResponse, responseErr := handler.Handle(ctx, &RequestInfo{
			Headers: r.Header,
			Cookies: r.Cookies(),
			Data:    params.Params,
		})
		if responseErr != nil {
			panic(responseErr)
		}
		applyHeaders(resHeaders, handlerResponse.Headers)
		handlerResponse.Headers = resHeaders
		return handlerResponse, incomingRequest.ID
	}
}

type RawRequestParser func(ctx context.Context, r *http.Request) (context.Context, *RequestBase, *RequestParams, custom_error.CustomError)

func NewRawHandle(newContext InitialContextFactory, methodHandler HandlingInfo, parse RawRequestParser) RawRequestHandler {
	handler := HandlingInfo{
		Handle:         methodHandler.Handle,
		ComposeContext: methodHandler.ComposeContext,
		GetHeaders:     methodHandler.GetHeaders,
		NewParams:      methodHandler.NewParams,
	}
	if handler.ComposeContext == nil {
		handler.ComposeContext = dummyContextFactory
	}
	if handler.GetHeaders == nil {
		handler.GetHeaders = dummyContextExpert
	}

	return func(srcCtx context.Context, r *http.Request) (*ResponseInfo, string) {

		resHeaders := http.Header{}
		ctx, incomingRequest, params, cErr := parse(srcCtx, r)
		if cErr != nil {
			ThrowError(0, 1, "Failed to parse request.", cErr)
		}
		ctx, composeErr := handler.ComposeContext(ctx, incomingRequest, r)
		if ctx != nil {
			applyHeaders(resHeaders, handler.GetHeaders(ctx))
		}
		if composeErr != nil {
			panic(composeErr)
		}

		handlerResponse, responseErr := handler.Handle(ctx, &RequestInfo{
			Headers: r.Header,
			Cookies: r.Cookies(),
			Data:    params.Params,
		})
		if responseErr != nil {
			panic(responseErr)
		}
		applyHeaders(resHeaders, handlerResponse.Headers)
		handlerResponse.Headers = resHeaders
		return handlerResponse, incomingRequest.ID
	}
}

func CreateRawHandler(newContext InitialContextFactory, handle RawRequestHandler, defaultHeaders HeadersFromContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := newContext()
		headers := defaultHeaders(ctx)
		applyHeaders(w.Header(), headers)
		if r.Body != nil {
			defer r.Body.Close()
		}
		var response []byte
		defer func() {
			v := recover()
			if v == nil {
				w.Write(response)
				return
			}
			debug.PrintStack()
			var rpcError UntypedResponse
			rpcError.Version = JSON_RPC_VERSION
			serverError, ok := v.(ServerError)
			if !ok {
				rpcError.Err = &Error{
					Code:    0,
					Message: "Unknow error",
					Data: UnknownErrorData{
						CallStack:     string(debug.Stack()),
						OriginalError: v,
					},
				}
			} else {
				rpcError.Err = serverError.ToError()
			}

			dataToSend, err := json.Marshal(rpcError)
			if err == nil {
				w.Write(dataToSend)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}()

		handlerResponse, rid := handle(ctx, r)

		applyHeaders(w.Header(), handlerResponse.Headers)

		response, err := json.Marshal(UntypedResponse{
			ResponseBase: ResponseBase{
				Version: JSON_RPC_VERSION,
				ID:      rid,
			},
			ResponseResult: ResponseResult{
				Result: handlerResponse.Data,
			},
		})
		if err != nil {
			ThrowError(0, 4, "Failed to marshal response.", err)
		}
	}
}

func CreateJSONRpcHandlerCustomUnmarshal(handlers map[string]HandlingInfo, getDecoder func(data io.Reader) *json.Decoder, defaultHeaders HeadersFromContext) func(w http.ResponseWriter, r *http.Request) {
	initialCtxFactory := func() context.Context {
		return context.Background()
	}
	handle := NewJsonRPCHandle(handlers, getDecoder)
	return CreateRawHandler(initialCtxFactory, handle, defaultHeaders)
}

func defaultDecoder(data io.Reader) *json.Decoder {
	return json.NewDecoder(data)
}

func CreateJSONRpcHandler(handlers map[string]HandlingInfo) func(w http.ResponseWriter, r *http.Request) {
	return CreateJSONRpcHandlerCustomUnmarshal(handlers, defaultDecoder, func(ctx context.Context) http.Header {
		return http.Header{}
	})
}

func CreateJSONRpcHandlerWithDefaultHeaders(handlers map[string]HandlingInfo, defaultHeaders HeadersFromContext) func(w http.ResponseWriter, r *http.Request) {
	return CreateJSONRpcHandlerCustomUnmarshal(handlers, defaultDecoder, defaultHeaders)
}
