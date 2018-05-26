package json_rpc

type RequestBase struct {
	Version string `json:"jsonrpc"`
	ID      string `json:"id,omitempty"`
	Method  string `json:"method"`
}

type RequestParams struct {
	Params interface{} `json:"params,omitempty"`
}
