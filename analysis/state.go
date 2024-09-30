package analysis

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"proof/lsp"
	"regexp"
	"strings"
	"unicode"

	"github.com/f1monkey/spellchecker"
)

type State struct {
	Spellchecker         *spellchecker.Spellchecker
	DictionaryPath       string
	AllowImplicitPlurals bool
	Documents            map[string]documentData
	MaxSuggestions       int
	ExcludedFilePatterns []string
	ExcludedFileTypes    []string
}

type documentData struct {
	URI         string
	LanguageID  string
	Extension   string
	Text        string
	Diagnostics []lsp.Diagnostic
}

func NewState(sc *spellchecker.Spellchecker) State {
	const DefaultMaxSuggestions = 5

	return State{
		Spellchecker:   sc,
		MaxSuggestions: DefaultMaxSuggestions,
		Documents:      make(map[string]documentData),
	}
}

// Workspace

func (s *State) UpdateSettings(settings lsp.Settings, logger *log.Logger) {
	s.AllowImplicitPlurals = settings.Proof.AllowImplicitPlurals
	s.MaxSuggestions = settings.Proof.MaxSuggestions
	s.DictionaryPath = settings.Proof.DictionaryPath
	s.ExcludedFilePatterns = settings.Proof.ExcludedFilePatterns
	s.ExcludedFileTypes = settings.Proof.ExcludedFileTypes

	s.Spellchecker.WithOpts(spellchecker.WithMaxErrors(settings.Proof.MaxErrors))

	reader := strings.NewReader(strings.Join(settings.Proof.IgnoredWords, "\n"))
	s.Spellchecker.AddFrom(reader)

	if s.DictionaryPath != "" {
		if err := ensureDir(s.DictionaryPath); err != nil {
			logger.Printf("Failed to create dictionary directory: %s", err)
		}

		file, err := os.Open(s.DictionaryPath)

		if err != nil {
			logger.Printf("Failed to open dictionary file: %s", err)
			return
		}

		defer file.Close()

		reader := bufio.NewReader(file)
		s.Spellchecker.AddFrom(reader)
	}

	logger.Printf(
		"Updated Settings "+
			"| AllowImplicitPlurals: %v "+
			"| DictionaryPath: %s "+
			"| MaxSuggestions: %d "+
			"| MaxErrors: %v "+
			"| IgnoredWords: %v "+
			"| ExcludedFileNames: %v "+
			"| ExcludedFileTypes: %v ",
		s.AllowImplicitPlurals,
		s.DictionaryPath,
		s.MaxSuggestions,
		settings.Proof.MaxErrors,
		settings.Proof.IgnoredWords,
		s.ExcludedFilePatterns,
		s.ExcludedFileTypes)
}

func (s *State) ExecuteCommand(command string, arguments []string, logger *log.Logger) (string, []lsp.Diagnostic) {
	switch command {
	case "proof.add_to_dictionary":
		if len(arguments) < 2 {
			logger.Print("No arguments provided for 'proof.add_to_dictionary'")
			return "", []lsp.Diagnostic{}
		}

		uri := arguments[0]
		word := strings.ToLower(arguments[1])

		s.Spellchecker.Add(word)

		if s.DictionaryPath != "" {
			if err := ensureDir(s.DictionaryPath); err != nil {
				logger.Printf("Failed to create dictionary directory: %s", err)
				return "", []lsp.Diagnostic{}
			}

			file, err := os.OpenFile(s.DictionaryPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

			if err != nil {
				logger.Printf("Failed to open dictionary file: %s", err)
				return "", []lsp.Diagnostic{}
			}

			defer file.Close()

			written_bytes, err := file.WriteString(word + "\n")

			if err != nil {
				logger.Printf("Failed to write to dictionary file: %s", err)
				return "", []lsp.Diagnostic{}
			}

			bytes := []byte(word + "\n")

			if written_bytes != len(bytes) {
				logger.Printf("Failed to write all bytes to dictionary file. written: %d | expected %d", written_bytes, len(bytes))
				return "", []lsp.Diagnostic{}
			}
		}

		logger.Printf("Added '%s' to dictionary", word)
		return uri, getDiagnosticsForFile(uri, s, logger)

	default:
		logger.Printf("Unknown command: %s", command)
		return "", []lsp.Diagnostic{}
	}
}

// Documents

func (s *State) OpenDocument(document lsp.TextDocumentItem, logger *log.Logger) ([]lsp.Diagnostic, bool) {
	uri := document.URI
	data := createDocumentData(document)
	diagnostics := getDiagnostics(data, s, logger)
	data.Diagnostics = diagnostics

	if data.isExcluded(s, logger) {
		return []lsp.Diagnostic{}, false
	}

	s.Documents[uri] = data

	return diagnostics, true
}

func (s *State) UpdateDocument(identifier lsp.VersionedTextDocumentIdentifier, change string, logger *log.Logger) ([]lsp.Diagnostic, bool) {
	uri := identifier.URI
	document := s.Documents[uri]
	currentDiagnostics := document.Diagnostics
	data := updateDocumentData(document, change)
	diagnostics := getDiagnostics(data, s, logger)
	data.Diagnostics = diagnostics

	if data.isExcluded(s, logger) {
		return []lsp.Diagnostic{}, false
	}

	s.Documents[uri] = data

	return diagnostics, !sliceEqual(currentDiagnostics, diagnostics)
}

func (s *State) Diagnostic(uri string, logger *log.Logger) ([]lsp.Diagnostic, bool) {
	data := s.Documents[uri]
	diagnostics := getDiagnosticsForFile(uri, s, logger)

	if sliceEqual(data.Diagnostics, diagnostics) {
		return []lsp.Diagnostic{}, false
	}

	data.Diagnostics = diagnostics
	s.Documents[uri] = data

	return diagnostics, true
}

func (s *State) CodeAction(request lsp.CodeActionRequest, uri string, logger *log.Logger) lsp.CodeActionResponse {
	params := request.Params
	rng := params.Range

	if rng.Start.Line != rng.End.Line {
		logger.Print("3")
		return lsp.CodeActionResponse{
			Response: lsp.CreateResponse(request.ID),
			Result:   []lsp.CodeAction{},
		}
	}

	actions := []lsp.CodeAction{}
	document, ok := s.Documents[uri]

	if !ok {
		return lsp.CodeActionResponse{
			Response: lsp.CreateResponse(request.ID),
			Result:   []lsp.CodeAction{},
		}
	}

	text := document.Text
	lines := strings.Split(text, "\n")
	line := lines[rng.Start.Line]
	full_range := growRange(line, rng)

	relevant_text := line[full_range.Start.Character : full_range.End.Character+1]

	words := splitIntoWords(rng.Start.Line, full_range.Start.Character, relevant_text)

	for _, word := range words {
		if s.Spellchecker.IsCorrect(strings.ToLower(word.Text)) {
			continue
		}

		has_trailing_s := strings.HasSuffix(word.Text, "s")
		has_trailing_es := strings.HasSuffix(word.Text, "es")

		if has_trailing_s {
			if s.Spellchecker.IsCorrect(strings.ToLower(word.Text[:len(word.Text)-1])) {
				continue
			}
		}

		if has_trailing_es {
			if s.Spellchecker.IsCorrect(strings.ToLower(word.Text[:len(word.Text)-1])) {
				continue
			}
		}

		suggestions, suggestions_err := s.Spellchecker.Suggest(word.Text, s.MaxSuggestions)

		if s.DictionaryPath != "" {
			if has_trailing_s && s.AllowImplicitPlurals {
				actions = append(actions, lsp.CodeAction{
					Title: fmt.Sprintf("Add '%s' to dictionary", word.Text),
					Command: &lsp.Command{
						Title:   "Add to dictionary",
						Command: "proof.add_to_dictionary",
						Arguments: []string{
							uri,
							word.Text[:len(word.Text)-1],
						},
					},
				})
			} else if has_trailing_es && s.AllowImplicitPlurals {
				actions = append(actions, lsp.CodeAction{
					Title: fmt.Sprintf("Add '%s' to dictionary", word.Text),
					Command: &lsp.Command{
						Title:   "Add to dictionary",
						Command: "proof.add_to_dictionary",
						Arguments: []string{
							uri,
							word.Text[:len(word.Text)-2],
						},
					},
				})
			} else {
				actions = append(actions, lsp.CodeAction{
					Title: fmt.Sprintf("Add '%s' to dictionary", word.Text),
					Command: &lsp.Command{
						Title:   "Add to dictionary",
						Command: "proof.add_to_dictionary",
						Arguments: []string{
							uri,
							word.Text,
						},
					},
				})
			}
		}

		if suggestions_err != nil {
			break
		}

		for _, suggestion := range suggestions {
			action := lsp.CodeAction{
				Title: fmt.Sprintf("Replace with '%s'", suggestion),
				Edit: &lsp.WorkspaceEdit{
					Changes: map[string][]lsp.TextEdit{
						uri: {
							{
								Range:   lineRange(word.Row, word.Start, word.End),
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

func lineRange(row, start, end int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: row, Character: start},
		End:   lsp.Position{Line: row, Character: end},
	}
}

func getDiagnosticsForFile(uri string, s *State, logger *log.Logger) []lsp.Diagnostic {
	document := s.Documents[uri]
	return getDiagnostics(document, s, logger)
}

func getDiagnostics(document documentData, s *State, logger *log.Logger) []lsp.Diagnostic {
	text := document.Text

	diagnostics := []lsp.Diagnostic{}
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

func checkSplitWordsWithStruct(row int, line string, s *State, _ *log.Logger, severity lsp.DiagnosticSeverity) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}

	words := splitIntoWords(row, 0, line)

	for _, word := range words {
		word_lower := strings.ToLower(word.Text)

		if s.Spellchecker.IsCorrect(word_lower) {
			continue
		}

		if s.AllowImplicitPlurals && strings.HasSuffix(word_lower, "s") {
			word_lower = word_lower[:len(word_lower)-1]

			if s.Spellchecker.IsCorrect(word_lower) {
				continue
			}
		}

		if s.AllowImplicitPlurals && strings.HasSuffix(word_lower, "es") {
			word_lower = word_lower[:len(word_lower)-1]

			if s.Spellchecker.IsCorrect(word_lower) {
				continue
			}
		}

		diagnostics = append(diagnostics, lsp.Diagnostic{
			Range:    lineRange(word.Row, word.Start, word.End),
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
	prev_rune := rune(0)

	for i, r := range runes {
		switch {
		case unicode.IsUpper(r):
			if unicode.IsLower(prev_rune) {
				if len(current_word) > 0 {
					words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: i + offset_from_start})
				}

				current_word = []rune{r}
				start = i
				prev_rune = r
				continue
			}

			if i < len(runes)-1 {
				next := runes[i+1]

				if !unicode.IsLower(next) {
					current_word = append(current_word, r)
					prev_rune = r
					continue
				}

				if len(current_word) > 0 {
					words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: i + offset_from_start})
				}

				current_word = []rune{r}
				start = i
				prev_rune = r
				continue
			}

			current_word = append(current_word, r)
			prev_rune = r

		case unicode.IsLower(r):
			current_word = append(current_word, r)
			prev_rune = r

		default:
			if len(current_word) > 0 {
				words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: i + offset_from_start})
			}

			current_word = []rune{}
			start = i + 1
			prev_rune = r
		}
	}

	if len(current_word) > 0 {
		words = append(words, Word{Text: string(current_word), Row: row, Start: start + offset_from_start, End: len(runes) + offset_from_start})
	}

	return words
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

func createDocumentData(document lsp.TextDocumentItem) documentData {
	uri := document.URI
	text := document.Text
	languageID := document.LanguageID

	last_index := strings.LastIndex(uri, ".")

	if last_index == -1 {
		return documentData{
			URI:        uri,
			Text:       text,
			LanguageID: languageID,
			Extension:  "",
		}
	}

	extension := uri[last_index:]

	return documentData{
		URI:        uri,
		Text:       text,
		LanguageID: languageID,
		Extension:  extension,
	}
}

func updateDocumentData(document documentData, change string) documentData {
	return documentData{
		URI:        document.URI,
		Text:       change,
		LanguageID: document.LanguageID,
		Extension:  document.Extension,
	}
}

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}

func contains[T comparable](list []T, item T) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}

func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (data documentData) isExcluded(s *State, logger *log.Logger) bool {
	for _, pattern := range s.ExcludedFilePatterns {
		re, err := regexp.Compile(pattern)

		if err != nil {
			logger.Panicf("Failed to compile regex pattern: %s, %s", pattern, err)
		}

		if re.MatchString(data.URI) {
			return true
		}
	}

	return contains(s.ExcludedFileTypes, data.LanguageID)
}
