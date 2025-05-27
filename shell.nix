{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    git
    curl
  ];

  shellHook = ''
    echo "Gmail Exporter development environment"
    echo "Go version: $(go version)"
  '';
} 