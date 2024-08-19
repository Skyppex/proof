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
	AllowImplicitPlurals bool                `json:"allowImplicitPlurals"`
	SpellCheckNodes      map[string][]string `json:"spellCheckNodes"`
}
