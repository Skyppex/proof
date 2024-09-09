# Proof

Proof is a simple and lightweight spell checking lsp primarily made for neovim.
It uses the
[f1monkey/spellchecker](https://github.com/f1monkey/spellchecker)
    library to provide spell checking capabilities and uses the words from
[makifdb/spellcheck/main/words.txt](https://raw.githubusercontent.com/makifdb/spellcheck/main/words.txt)
with some additions for developer specific words, abbreviations and tools.

## Features

- **Uses Diagnostic LSP**: Proof uses the diagnostic LSP to provide spell
checking diagnostics.
- **Fast**: Proof diagnostics across the entire file you're
working on instantly (unless you have a horrendously large file of say 300'000
lines :eyes:).
- **Customizable**: You can add your own words to a dictionary file which is
used by all instances of proof (after they have started).
- **Lightweight**: I have seen proof using at most 60 MBof memory after running for a while
  opening several different files.

## Installation

To use proof, you need to have a working LSP client. Here is an example using
lspconfig with the built-in LSP client:

```lua
local lspconfig = require("lspconfig")
local configs = require("lspconfig.configs")

if not configs.proof then
    configs.proof = {
        default_config = {
            -- Optional log file path. If not set, proof will not log anything.
            -- Use a log file for debugging issues or when contributing.
            cmd = { proof_exe, log_file },
            -- Use "*" for all filetypes and excludes in the settings or specify
            -- only some filetypes here.
            filetypes = { "*" },
            single_file_support = true,
            root_dir = lspconfig.util.find_git_ancestor,
            settings = {},
            -- Make sure to let proof know about the LSP clients capabilities.
            capabilities = capabilities,
        },
    }
end

lspconfig.proof.setup({
    settings = {
        proof = {
            -- Full path to a dictionary file on your system
            dictionaryPath = string.gsub(vim.fn.stdpath("config") .. "/proof/dictionary.txt", "\\", "/"),

            -- max diff in bits between the "search word" and a "dictionary word".
            -- i.e. one simple symbol replacement (problam => problem) is a two-bit difference.
            -- Making this value too high will result in a hit to performance.
            maxErrors = 2,

            -- Max number of suggestions to show when doing a code action.
            -- The number of possible suggestions grows based on the maxErrors
            -- value.
            maxSuggestions = 5,

            -- If true, words which end with 's' will be valid even if the
            -- dictionary only contains the word without the 's' at the end.
            -- The same is true for 'es' words.
            allowImplicitPlurals = true,

            -- You can also choose to feed some words to the spell checker here.
            ignoredWords = {},

            -- File names which should be excluded from spell checking.
            excludedFileNames = {},

            -- File types which should be excluded from spell checking.
            -- This uses neovim's `&filetype` variable. Or more specifically the
            -- languageId sent to proof by the LSP client.
            excludedFileTypes = {},

            -- File extensions which should be excluded from spell checking.
            excludedFileExtensions = {},
        },
    },
})
```
