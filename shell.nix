{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    git
    curl
    golangci-lint
  ];

  shellHook = ''
    echo "Gmail Exporter development environment"
    echo "Go version: $(go version)"
    echo "golangci-lint version: $(golangci-lint version)"
  '';
} 