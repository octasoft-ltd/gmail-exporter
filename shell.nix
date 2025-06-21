{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go toolchain
    go

    # Development tools
    git
    curl

    # Additional useful tools for Go development
    gopls          # Go language server
    delve          # Go debugger

    # Build tools
    gnumake
  ];

  # Environment variables for private modules (matching CI)
  GOPRIVATE = "github.com/octasoft-ltd/*";
  GONOPROXY = "github.com/octasoft-ltd/*";
  GONOSUMDB = "github.com/octasoft-ltd/*";

  shellHook = ''
    echo "Gmail Exporter development environment"
    echo "Go version: $(go version)"

    # Ensure Go bin is in PATH for installed tools
    export PATH="$HOME/go/bin:$PATH"

    # Install golangci-lint if not present
    if ! command -v golangci-lint &> /dev/null; then
      echo "Installing golangci-lint latest version..."
      go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi

    echo "golangci-lint version: $(golangci-lint version 2>/dev/null || echo 'not found')"
    echo ""
    echo "Environment configured for private modules:"
    echo "  GOPRIVATE=$GOPRIVATE"
    echo "  GONOPROXY=$GONOPROXY"
    echo "  GONOSUMDB=$GONOSUMDB"
    echo ""
    echo "Available commands:"
    echo "  go test ./...              - Run tests"
    echo "  golangci-lint run          - Run linter"
    echo "  make                       - Build project"
    echo "  dlv debug                  - Debug with delve"
    echo ""
    echo "To update golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  '';
}
