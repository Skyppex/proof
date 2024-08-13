package lsp

type Request struct {
	RPC    string `json:"jsonrpc"`
	ID     int    `json:"id"`
	Method string `json:"method"`

	// We will just specify the type of the params in all the request types
	// later
	// Params ...
}

type Response struct {
	RPC string `json:"jsonrpc"`
	ID  *int   `json:"id,omitempty"`

	// Result
	// Error
}

func CreateResponse(id int) Response {
	return Response{
		RPC: "2.0",
		ID:  &id,
	}
}

type Notification struct {
	RPC    string `json:"jsonrpc"`
	Method string `json:"method"`
}

func CreateNotification(method string) Notification {
	return Notification{
		RPC:    "2.0",
		Method: method,
	}
}
