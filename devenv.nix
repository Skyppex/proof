{pkgs, ...}: {
  # https://devenv.sh/packages/
  packages = with pkgs; [
    golangci-lint
    alejandra
  ];

  # https://devenv.sh/languages/
  languages.go.enable = true;
  languages.nix.enable = true;
}
