# Roadmap

- [x] Spell-checking single words in English

  - [x] Single words separated by spaces
  - [x] Words should be separated by casing
    - [x] PascalCase
    - [x] camelCase
    - [x] snake_case or SCREAMING_SNAKE_CASE
    - [x] kebab-case, KEBAB-CASE or Train-Case

- [ ] Add code actions

  - [x] Add code action to add a word to the dictionary
  - [x] Add code action to replace a word with a suggestion
  - [ ] Add code action to replace all the same words in buffer with a suggestion

- [ ] Add treesitter support
  - [ ] Allow user to configure which nodes should be spell-checked for
        each language
  - [ ] Allow user to use custom captures by making queries in the
        proof.scm file
