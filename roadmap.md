# Roadmap

- [ ] Spell-checking single words in English
    - [ ] Single words separated by spaces
    - [ ] Words should be separated by casing
        - [ ] PascalCase
        - [ ] camelCase
        - [ ] snake_case or SCREAMING_SNAKE_CASE
        - [ ] kebab-case, KEBAB-CASE or Train-Case
    - [ ] Add treesitter support
        - [ ] Allow user to configure which nodes should be spell-checked for
            each language
        - [ ] Allow user to use custom captures by making queries in the
            proof.scm file

- [ ] Add ignored words
    - [ ] Add ignored words to the dictionary file found in neovim's root path
        Same as where the `spell` file is, might even use that file
    - [ ] Add ignored words to the configuration file
