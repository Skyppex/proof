package analysis

import (
	"fmt"
	"github.com/f1monkey/spellchecker"
	"log"
	"proof/lsp"
	"strings"
)

type State struct {
	Spellchecker *spellchecker.Spellchecker
	Documents    map[string]string
}

func NewState(sc *spellchecker.Spellchecker) State {
	return State{
		Spellchecker: sc,
		Documents:    map[string]string{},
	}
}

func getDiagnostics(text string, s *State, logger *log.Logger) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	severity := lsp.Error

	for row, line := range strings.Split(text, "\n") {
		for column, word := range strings.Split(line, " ") {
			logger.Printf("Word: %s", word)

			if s.Spellchecker.IsCorrect(word) {
				continue
			}

			logger.Printf("Incorrect: %s", word)

			if word != "VSCode" {
				continue
			}

			logger.Printf("Is VSCode: %s", word)

			diagnostics = append(diagnostics, lsp.Diagnostic{
				Range:    LineRange(row, column, column+len(word)),
				Severity: &severity,
				Source:   "proof",
				Message:  fmt.Sprintf("Unknown word: %s", word),
			})
		}
	}

	return diagnostics
}

func (s *State) OpenDocument(uri string, text string, logger *log.Logger) []lsp.Diagnostic {
	s.Documents[uri] = text
	return getDiagnostics(text, s, logger)
}

func (s *State) UpdateDocument(uri string, text string, logger *log.Logger) []lsp.Diagnostic {
	s.Documents[uri] = text
	return getDiagnostics(text, s, logger)
}

func (s *State) CodeAction(id int, uri string) lsp.CodeActionResponse {
	actions := []lsp.CodeAction{}

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
