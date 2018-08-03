package json_rpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"
	"github.com/coldze/primitives/custom_error"
	"runtime/debug"
	"github.com/google/uuid"
)

type ResponseResultFactory func() interface{}
type IDFactory func() string

type RPCArguments struct {
	Headers http.Header
	Cookies []*http.Cookie
	Data    interface{}
}

type Client interface {
	Call(url string, method string, args RPCArguments, expectedReult ResponseResultFactory) (*UntypedResponse, custom_error.CustomError)
}

type client struct {
	httpClient *http.Client
	rpcVersion string
	getID IDFactory
}

func (c *client) Call(url string, method string, args RPCArguments, expectedResult ResponseResultFactory) (response *UntypedResponse, resError custom_error.CustomError) {
	requestID := c.getID()
	errorBuilder := custom_error.NewPrefixedErrorBuilder(fmt.Sprintf("json-rpc request ID: '%v'. Method: '%v'. URL: '%v'. ", requestID, method, url))
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		stacktrace := "\nPanic stack trace:\n" + string(debug.Stack())
		response = nil
		customErr, ok := r.(custom_error.CustomError)
		if ok {
			resError = errorBuilder.NewErrorf(customErr, "Failed to make a call. Panic occurred.%v", stacktrace)
			return
		}
		err, ok := r.(error)
		if ok {
			resError = errorBuilder.NewErrorf(customErr, "Failed to make a call. Panic occurred with error: %v.%v", err, stacktrace)
			return
		}
		resError = errorBuilder.NewErrorf(customErr, "Failed to make a call. Panic occurred with unknown error: %+v. Type: %T.%v", err, err, stacktrace)
	}()
	request := UntypedRequest{
		RequestBase{
			Method:  method,
			ID:      requestID,
			Version: c.rpcVersion,
		},
		RequestParams{
			Params: args.Data,
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errorBuilder.MakeErrorf("Failed to marshal request. Error: %v", err)
	}
	r := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, errorBuilder.MakeErrorf("Failed to create http-request. Error: %v", err)
	}
	//if args.Headers != nil {
		req.Header = args.Headers
	//}
	req.Header.Set("Content-Type", "application/json")
	for i := range args.Cookies {
		req.AddCookie(args.Cookies[i])
	}

	resp, err := c.httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to send request. Error: %v", err)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to read response. Error: %v", err)
	}
	responseBase := UntypedResponse{
		ResponseResult: ResponseResult{
			Result: expectedResult(),
		},
	}
	err = json.Unmarshal(respData, &responseBase)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to unmarshal response. Error: %v.", err)
	}

	return &responseBase, nil
}

func guidID() string {
	return uuid.New().String()
}

func NewClient(httpClient *http.Client) Client {
	return &client{
		httpClient: httpClient,
		rpcVersion: json_rpc_version,
		getID: guidID,
	}
}
