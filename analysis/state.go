package analysis

import (
	"fmt"
	"log"
	"proof/lsp"
	"regexp"
	"strings"

	"github.com/f1monkey/spellchecker"
)

// state

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
	return diagnostics // This is just to make the noise go away

	severity := lsp.Hint

	regex := regexp.MustCompile("[_a-zA-Z]+")

	for row, line := range strings.Split(text, "\n") {
		if strings.Trim(line, "\t \r\n") == "" {
			continue
		}

		line_diagnostics := checkRegexMatches(row, line, s, logger, severity, regex)
		// line_diagnostics := checkSplitWords(row, line, s, logger, severity)

		diagnostics = append(diagnostics, line_diagnostics...)

	}

	return diagnostics
}

func checkRegexMatches(row int, line string, s *State, logger *log.Logger, severity lsp.DiagnosticSeverity, regex *regexp.Regexp) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}

	matches := regex.FindAllStringIndex(line, -1)

	for _, match := range matches {
		start, end := match[0], match[1]

		word := strings.Trim(line[start:end], "\t \r\n")

		if word == "" {
			continue
		}

		word_lower := strings.ToLower(word)

		logger.Printf("Word: %s is at index %d", word, start)

		if s.Spellchecker.IsCorrect(word_lower) {
			continue
		}

		logger.Printf("Incorrect: %s", word)

		// if word != "VSCode" {
		// 	continue
		// }
		//
		// logger.Printf("Is VSCode: %s", word)

		diagnostics = append(diagnostics, lsp.Diagnostic{
			Range:    LineRange(row, start, end),
			Severity: &severity,
			Source:   "proof",
			Message:  fmt.Sprintf("Typo in word: %s", word),
		})
	}

	return diagnostics
}

func checkSplitWords(row int, line string, s *State, logger *log.Logger, severity lsp.DiagnosticSeverity) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	column := 0

	for _, word := range strings.Split(line, " ") {
		index := strings.Index(line[column:], word) + column

		logger.Printf("Row: %d, Col: %d, Line: %s", row, column, line)
		word := strings.Trim(word, "\t \r\n")

		if word == "" {
			column = index + len(word) + 1
			continue
		}

		logger.Printf("Word: %s is at index %d", word, column)

		if s.Spellchecker.IsCorrect(word) {
			column = index + len(word) + 1
			continue
		}

		logger.Printf("Incorrect: %s", word)

		if word != "VSCode" {
			column = index + len(word) + 1
			continue
		}

		logger.Printf("Is VSCode: %s", word)

		column = index + len(word) + 1

		diagnostics = append(diagnostics, lsp.Diagnostic{
			Range:    LineRange(row, column, column+len(word)),
			Severity: &severity,
			Source:   "proof",
			Message:  fmt.Sprintf("Typo in word: %s", word),
		})
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
