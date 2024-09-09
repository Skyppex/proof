package lsp

type DiagnosticRequest struct {
	Request
	Params DiagnosticRequestParams `json:"params"`
}

type DiagnosticRequestParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentDiagnosticReportKind string

const (
	Full      DocumentDiagnosticReportKind = "full"
	Unchanged DocumentDiagnosticReportKind = "unchanged"
)

type DiagnosticResponse struct {
	Response
	Result DocumentDiagnosticReport `json:"result"`
}

type DocumentDiagnosticReport struct {
	Kind             DocumentDiagnosticReportKind        `json:"kind"`
	RelatedDocuments map[string]DocumentDiagnosticReport `json:"relatedDocuments"`
	Items            *[]Diagnostic                       `json:"items"`
	ResultId         string                              `json:"resultId"`
}

func NewDiagnosticResponse(id int, kind DocumentDiagnosticReportKind, items []Diagnostic, uri string) DiagnosticResponse {
	maybeItems := items

	if kind == Unchanged {
		maybeItems = nil
	}

	return DiagnosticResponse{
		Response: CreateResponse(id),
		Result: DocumentDiagnosticReport{
			Kind:             kind,
			RelatedDocuments: make(map[string]DocumentDiagnosticReport),
			Items:            &maybeItems,
			ResultId:         uri,
		},
	}
}
