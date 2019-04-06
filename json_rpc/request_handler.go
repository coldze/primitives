package json_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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

type RequestHandler func(ctx context.Context, request *RequestInfo) (*ResponseInfo, ServerError)
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
		log.Printf(" IN 1")
		incomingRequest := RequestBase{}
		err = json.Unmarshal(data, &incomingRequest)
		log.Printf(" IN 2. Request: %+v. URI: %+v. Body: %+v", incomingRequest, r.RequestURI, string(data))
		if err != nil {
			log.Printf(" IN 3")
			ThrowError(0, 1, "Failed to parse request body.", err)
		}
		log.Printf(" IN 4")

		if incomingRequest.Version != json_rpc_version {
			log.Printf(" IN 5")
			ThrowError(0, 2, "Unsupported JSON RPC version", errors.New("Expected version = 2.0"))
		}

		log.Printf(" IN 6")
		handler, ok := methodHandlers[incomingRequest.Method]
		if !ok {
			log.Printf(" IN 7")
			ThrowError(0, 3, "Unsupported method: "+incomingRequest.Method, nil)
		}

		log.Printf(" IN 8")
		ctx, composeErr := handler.ComposeContext(&incomingRequest, r)
		if ctx != nil {
			log.Printf(" IN 9")
			applyHeaders(w, handler.GetHeaders(ctx))
		}
		log.Printf(" IN 10")
		if composeErr != nil {
			log.Printf(" IN 11")
			panic(composeErr)
		}

		log.Printf(" IN 12")
		params := RequestParams{
			Params: handler.NewParams(),
		}
		log.Printf(" IN 13")
		err = json.Unmarshal(data, &params)
		if err != nil {
			log.Printf(" IN 14")
			ThrowError(0, 3, "Failed to prepare arguments for handler,", err)
		}
		handlerResponse, responseErr := handler.Handle(ctx, &RequestInfo{
			Headers: r.Header,
			Cookies: r.Cookies(),
			Data:    params.Params,
		})
		log.Printf(" IN 15")
		if responseErr != nil {
			log.Printf(" IN 16")
			panic(responseErr)
		}
		log.Printf(" IN 17")
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
		log.Printf(" IN 18. Response: %+v", string(response))
		if err != nil {
			log.Printf(" IN 19")
			ThrowError(0, 4, "Failed to marshal response.", err)
		}
		log.Printf(" IN 20")

	}
}
