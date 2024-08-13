package lsp

type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

type InitializeRequestParams struct {
	ProcessId  int         `json:"processId"`
	ClientInfo *ClientInfo `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializerResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync   int               `json:"textDocumentSync"`
	HoverProvider      bool              `json:"hoverProvider"`
	CodeActionProvider bool              `json:"codeActionProvider"`
	DiagnosticProvider DiagnosticOptions `json:"diagnosticProvider"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DiagnosticOptions struct {
	Identifier            string `json:"identifier"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

func NewInitializeResponse(id int) InitializerResponse {
	return InitializerResponse{
		Response: CreateResponse(id),
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:   1,
				HoverProvider:      true,
				CodeActionProvider: true,
				DiagnosticProvider: DiagnosticOptions{
					Identifier:            "proof",
					InterFileDependencies: false,
					WorkspaceDiagnostics:  false,
				},
			},
			ServerInfo: &ServerInfo{
				Name:    "proof",
				Version: "0.1.0",
			},
		},
	}
}
