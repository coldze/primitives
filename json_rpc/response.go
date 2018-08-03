package json_rpc

type ResponseBase struct {
	Version string `json:"jsonrpc"`
	ID      string `json:"id,omitempty"`
	Err     *Error `json:"error,omitempty"`
}

type ResponseResult struct {
	Result interface{} `json:"result,omitempty"`
}

type UntypedResponse struct {
	ResponseBase
	ResponseResult
}
