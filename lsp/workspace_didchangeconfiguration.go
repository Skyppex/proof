package lsp

type DidChangeConfigurationRequest struct {
	Request
	Params DidChangeConfigurationParams `json:"params"`
}

type DidChangeConfigurationParams struct {
	Settings Settings `json:"settings"`
}

type Settings struct {
	Proof ProofSettings `json:"proof"`
}

type ProofSettings struct {
	DictionaryPath       string              `json:"dictionaryPath"`
	AllowImplicitPlurals bool                `json:"allowImplicitPlurals"`
	MaxSuggestions       int                 `json:"maxSuggestions"`
	SpellCheckNodes      map[string][]string `json:"spellCheckNodes"`
}
