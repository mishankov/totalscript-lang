# TotalScript Development Guide

## Requirements

- Go 1.25+
- [Task](https://taskfile.dev/) - task runner
- [golangci-lint](https://golangci-lint.run/) - linter

## Project Structure

```
totalscript/
├── cmd/
│   ├── tsl/              # CLI interpreter
│   └── tsl-lsp/          # LSP server
├── internal/
│   ├── token/            # Token types
│   ├── lexer/            # Lexer
│   ├── ast/              # AST nodes
│   ├── parser/           # Parser
│   ├── types/            # Type system
│   ├── analyzer/         # Semantic analysis
│   ├── interpreter/      # Runtime execution
│   ├── stdlib/           # Built-in functions
│   ├── database/         # SQLite integration
│   └── lsp/              # LSP protocol handling
├── Taskfile.yaml
├── .golangci.yml
└── go.mod
```

## Common Tasks

```bash
# Run all checks (lint + test)
task check

# Run linter
task lint

# Run tests
task test

# Run tests with coverage
task test:coverage

# Build CLI
task build

# Build LSP
task build:lsp

# Build all
task build:all

# Clean build artifacts
task clean
```

## Coding Standards

### Error Handling

Use custom error types for each package:

```go
// internal/parser/errors.go
package parser

type ParseError struct {
    Line    int
    Column  int
    Message string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Message)
}
```

### Testing

Use table-driven tests:

```go
func TestLexer_Numbers(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []token.Token
    }{
        {
            name:  "integer",
            input: "42",
            expected: []token.Token{
                {Type: token.INTEGER, Literal: "42"},
            },
        },
        {
            name:  "float",
            input: "3.14",
            expected: []token.Token{
                {Type: token.FLOAT, Literal: "3.14"},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lexer := New(tt.input)
            // ... assertions
        })
    }
}
```

### Documentation

Document all exported functions, types, and methods:

```go
// Token represents a lexical token in TotalScript source code.
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
    // ...
}
```

### Package Organization

- Keep packages focused and small
- Avoid circular dependencies
- Use interfaces for abstraction between packages

## Git Workflow

### Commit Messages

Follow conventional commits:

```
feat: add lexer for string literals
fix: handle escaped characters in strings
refactor: simplify parser state machine
test: add table-driven tests for lexer
docs: update development guide
```

### Branches

- `main` - stable, release-ready code
- `feat/*` - new features
- `fix/*` - bug fixes
