# Roadmap

- [ ] Spell-checking single words in English
    - [x] Single words separated by spaces
    - [x] Words should be separated by casing
        - [x] PascalCase
        - [x] camelCase
        - [x] snake_case or SCREAMING_SNAKE_CASE
        - [x] kebab-case, KEBAB-CASE or Train-Case

- [ ] Add treesitter support
    - [ ] Allow user to configure which nodes should be spell-checked for
        each language
    - [ ] Allow user to use custom captures by making queries in the
        proof.scm file

- [x] Add ignored words
    - [x] Add ignored words to a dictionary file with a user-defined path
    - [x] Add ignored words to the configuration file
