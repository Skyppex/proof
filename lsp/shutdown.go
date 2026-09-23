package lsp

type Shutdown struct {
	Request
}

type ShutdownResponse struct {
	Response
	Result *any `json:"result"`
}

func NewShutdownResponse(id int) ShutdownResponse {
	return ShutdownResponse{
		Response: CreateResponse(id),
		Result:   nil,
	}
}
