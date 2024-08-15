package analysis

import (
	"fmt"
	"log"
	"proof/lsp"
	"strings"
	"unicode"

	"github.com/f1monkey/spellchecker"
)

const NumberOfSuggestions = 5
const suggestions = 10

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
	// return diagnostics // This is just to make the noise go away

	severity := lsp.Hint

	for row, line := range strings.Split(text, "\n") {
		if strings.Trim(line, "\t \r\n") == "" {
			continue
		}

		line_diagnostics := checkSplitWordsWithStruct(row, line, s, logger, severity)

		diagnostics = append(diagnostics, line_diagnostics...)

	}

	return diagnostics
}

func checkSplitWordsWithStruct(row int, line string, s *State, logger *log.Logger, severity lsp.DiagnosticSeverity) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}

	words := splitIntoWords(row, 0, line)

	for _, word := range words {
		word_lower := strings.ToLower(word.Text)

		if s.Spellchecker.IsCorrect(word_lower) {
			continue
		}

		if strings.HasSuffix(word_lower, "s") {
			word_lower = word_lower[:len(word_lower)-1]
			if s.Spellchecker.IsCorrect(word_lower) {
				continue
			}
		}

		diagnostics = append(diagnostics, lsp.Diagnostic{
			Range:    LineRange(word.Row, word.Start, word.End),
			Severity: &severity,
			Source:   "proof",
			Message:  fmt.Sprintf("Typo in word: %s", word.Text),
		})
	}

	return diagnostics
}

type Word struct {
	Text  string
	Row   int
	Start int
	End   int
}

func splitIntoWords(row int, offset_from_start int, line string) []Word {
	words := []Word{}
	runes := []rune(line)

	start := 0
	current_word := []rune{}

	for i, r := range runes {
		switch {
		case unicode.IsUpper(r):
			if len(current_word) > 0 {
				words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: i + offset_from_start})
			}

			current_word = []rune{r}
			start = i

		case unicode.IsLower(r):
			current_word = append(current_word, r)

		default:
			if len(current_word) > 0 {
				words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: i + offset_from_start})
			}

			current_word = []rune{}
			start = i + 1
		}
	}

	if len(current_word) > 0 {
		words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: len(runes) + offset_from_start})
	}

	return words
}

func (s *State) OpenDocument(uri string, text string, logger *log.Logger) []lsp.Diagnostic {
	s.Documents[uri] = text
	return getDiagnostics(text, s, logger)
}

func (s *State) UpdateDocument(uri string, text string, logger *log.Logger) []lsp.Diagnostic {
	s.Documents[uri] = text
	return getDiagnostics(text, s, logger)
}

func (s *State) CodeAction(request lsp.CodeActionRequest, uri string, logger *log.Logger) lsp.CodeActionResponse {
	params := request.Params
	rng := params.Range

	if rng.Start.Line != rng.End.Line {
		return lsp.CodeActionResponse{
			Response: lsp.CreateResponse(request.ID),
			Result:   []lsp.CodeAction{},
		}
	}

	actions := []lsp.CodeAction{}
	text := s.Documents[uri]
	lines := strings.Split(text, "\n")
	line := lines[rng.Start.Line]
	full_range := growRange(line, rng)

	relevant_text := line[full_range.Start.Character : full_range.End.Character+1]

	words := splitIntoWords(rng.Start.Line, full_range.Start.Character, relevant_text)

	for _, word := range words {
		if s.Spellchecker.IsCorrect(strings.ToLower(word.Text)) {
			continue
		}

		if strings.HasSuffix(word.Text, "s") {
			if s.Spellchecker.IsCorrect(strings.ToLower(word.Text[:len(word.Text)-1])) {
				continue
			}
		}

		suggestions, err := s.Spellchecker.Suggest(strings.ToLower(word.Text), NumberOfSuggestions)

		if err != nil {
			logger.Print("Failed to get suggestions:")
			continue
		}

		for _, suggestion := range suggestions {
			action := lsp.CodeAction{
				Title: fmt.Sprintf("Replace with '%s'", suggestion),
				Edit: &lsp.WorkspaceEdit{
					Changes: map[string][]lsp.TextEdit{
						uri: {
							{
								Range:   LineRange(word.Row, word.Start, word.End),
								NewText: suggestion,
							},
						},
					},
				},
			}

			actions = append(actions, action)
		}

	}

	response := lsp.CodeActionResponse{
		Response: lsp.CreateResponse(request.ID),
		Result:   actions,
	}

	return response
}

func growRange(text string, rng lsp.Range) lsp.Range {
	start := rng.Start
	end := rng.End

	for start.Character > 0 {
		if !unicode.IsLetter(rune(text[start.Character-1])) {
			break
		}

		start.Character--
	}

	for end.Character < len(text) {
		if !unicode.IsLetter(rune(text[end.Character+1])) {
			break
		}

		end.Character++
	}

	return lsp.Range{Start: start, End: end}
}

func LineRange(row, start, end int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: row, Character: start},
		End:   lsp.Position{Line: row, Character: end},
	}
}
