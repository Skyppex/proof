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
	DictionaryPath       string   `json:"dictionaryPath"`
	AllowImplicitPlurals bool     `json:"allowImplicitPlurals"`
	MaxErrors            int      `json:"maxErrors"`
	MaxSuggestions       int      `json:"maxSuggestions"`
	IgnoredWords         []string `json:"ignoredWords"`
	ExcludedFilePatterns []string `json:"excludedFilePatterns"`
	ExcludedFileTypes    []string `json:"excludedFileTypes"`
}
