package lsp

type ExecuteCommandRequest struct {
	Request
	Params ExecuteCommandParams `json:"params"`
}

type ExecuteCommandParams struct {
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}
