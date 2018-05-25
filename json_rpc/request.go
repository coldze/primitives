package json_rpc

type Request struct {
	Version string      `json:"jsonrpc"`
	ID      string      `json:"id,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Method  string      `json:"method"`
}
