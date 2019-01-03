package json_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"runtime/debug"
)

const (
	json_rpc_version = "2.0"
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

type RequestHandler func(ctx context.Context, request *RequestInfo) (ResponseInfo, ServerError)
type RequestParamsFactory func() interface{}
type HeadersFromContext func(ctx context.Context) http.Header

type HandlingInfo struct {
	Handle         RequestHandler
	NewParams      RequestParamsFactory
	ComposeContext ContextFactory
	GetHeaders     HeadersFromContext
}

type UnknownErrorData struct {
	CallStack     string      `json:"call_stack"`
	OriginalError interface{} `json:"original_error"`
}

func applyHeaders(w http.ResponseWriter, headers http.Header) {
	for k, hs := range headers {
		for i := range hs {
			w.Header().Add(k, hs[i])
		}
	}
}

func dummyContextFactory(request *RequestBase, r *http.Request) (context.Context, ServerError) {
	return context.Background(), nil
}

func dummyContextExpert(ctx context.Context) http.Header {
	return nil
}

func CreateJSONRpcHandler(handlers map[string]HandlingInfo) func(w http.ResponseWriter, r *http.Request) {
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
	return func(w http.ResponseWriter, r *http.Request) {
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
			var rpcError UntypedResponse
			rpcError.Version = json_rpc_version
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

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ThrowError(0, 0, "Failed to read request body.", err)
		}
		incomingRequest := RequestBase{}
		err = json.Unmarshal(data, &incomingRequest)
		if err != nil {
			ThrowError(0, 1, "Failed to parse request body.", err)
		}

		if incomingRequest.Version != json_rpc_version {
			ThrowError(0, 2, "Unsupported JSON RPC version", errors.New("Expected version = 2.0"))
		}

		handler, ok := methodHandlers[incomingRequest.Method]
		if !ok {
			ThrowError(0, 3, "Unsupported method: "+incomingRequest.Method, nil)
		}

		ctx, composeErr := handler.ComposeContext(&incomingRequest, r)
		if ctx != nil {
			applyHeaders(w, handler.GetHeaders(ctx))
		}
		if composeErr != nil {
			panic(composeErr)
		}

		params := RequestParams{
			Params: handler.NewParams(),
		}
		err = json.Unmarshal(data, &params)
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
		applyHeaders(w, handlerResponse.Headers)

		response, err = json.Marshal(UntypedResponse{
			ResponseBase: ResponseBase{
				Version: json_rpc_version,
				ID:      incomingRequest.ID,
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
