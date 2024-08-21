package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func ParseFile(s *State, document documentData) (tree *sitter.Tree, err error) {
	// Get the language from the document
	// lang_id := document.LanguageID

	// Then use that with the ts_map to find the parser for that language based
	// on the ts_path in the state object

	// language := sitter.NewLanguage(ptr)
	// s.Parser.SetLanguage(javascript.GetLanguage())
	// s.Parser.ParseCtx()

	panic("not implemented")
}
