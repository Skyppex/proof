{
  description = "";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages.default = pkgs.buildGoModule {
        pname = "proof";
        version = "0.1.0";

        src = self;

        vendorHash = "sha256-eXa3+wiLrShg8kBv3ZxYxqMfnj2wrgg7qRW3pRd2apY=";
      };

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          golangci-lint
          alejandra
          nixd
        ];
      };
    });
}
