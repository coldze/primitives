package json_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
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

type HandlingInfo struct {
	Handle    RequestHandler
	NewParams RequestParamsFactory
}

func CreateJSONRpcHandler(methodHandlers map[string]HandlingInfo) func(w http.ResponseWriter, r *http.Request) {
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
			var rpcError Response
			rpcError.Version = json_rpc_version
			serverError, ok := v.(ServerError)
			if !ok {
				rpcError.Err = &Error{
					Code:    0,
					Message: "Unknow error",
					Data:    v,
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
		params := RequestParams{
			Params: handler.NewParams(),
		}
		err = json.Unmarshal(data, &params)
		if err != nil {
			ThrowError(0, 3, "Failed to prepare arguments for handler,", err)
		}
		ctx := context.Background()
		handlerResponse, responseErr := handler.Handle(ctx, &RequestInfo{
			Headers: r.Header,
			Cookies: r.Cookies(),
			Data:    params.Params,
		})
		if responseErr != nil {
			panic(responseErr)
		}
		for k, hs := range handlerResponse.Headers {
			for i := range hs {
				w.Header().Add(k, hs[i])
			}
		}

		response, err = json.Marshal(Response{
			Version: json_rpc_version,
			ID:      incomingRequest.ID,
			Result:  handlerResponse.Data,
		})
		if err != nil {
			ThrowError(0, 4, "Failed to marshal response.", err)
		}

	}
}
