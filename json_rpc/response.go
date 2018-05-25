package json_rpc

type Response struct {
	Version string      `json:"jsonrpc"`
	ID      string      `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Err     *Error      `json:"error,omitempty"`
}
