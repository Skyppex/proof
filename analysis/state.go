package analysis

import (
	"fmt"
	"proof/lsp"
	"strings"
)

type State struct {
	Documents map[string]string
}

func NewState() State {
	return State{Documents: map[string]string{}}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) Hover(id int, uri string, position lsp.Position) lsp.HoverResponse {
	document := s.Documents[uri]

	return lsp.HoverResponse{
		Response: lsp.CreateResponse(id),
		Result: lsp.HoverResult{
			Contents: fmt.Sprintf("File: %s, Characters: %d", uri, len(document)),
		},
	}
}

func (s *State) CodeAction(id int, uri string) lsp.CodeActionResponse {
	text := s.Documents[uri]

	actions := []lsp.CodeAction{}

	for row, line := range strings.Split(text, "\n") {
		idx := strings.Index(line, "// ")

		if idx >= 0 {
			todoChange := map[string][]lsp.TextEdit{}
			todoChange[uri] = []lsp.TextEdit{
				{
					Range:   LineRange(row, idx, idx+2),
					NewText: "// TODO: ",
				},
			}

			actions = append(actions, lsp.CodeAction{
				Title: "Comment TODO",
				Edit:  &lsp.WorkspaceEdit{Changes: todoChange},
			})

			noteChange := map[string][]lsp.TextEdit{}
			noteChange[uri] = []lsp.TextEdit{
				{
					Range:   LineRange(row, idx, idx+2),
					NewText: "// NOTE: ",
				},
			}

			actions = append(actions, lsp.CodeAction{
				Title: "Comment NOTE",
				Edit:  &lsp.WorkspaceEdit{Changes: noteChange},
			})
		}
	}

	response := lsp.CodeActionResponse{
		Response: lsp.CreateResponse(id),
		Result:   actions,
	}

	return response
}

func LineRange(row, start, end int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: row, Character: start},
		End:   lsp.Position{Line: row, Character: end},
	}
}
