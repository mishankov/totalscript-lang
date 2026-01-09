# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TotalScript is a scripting language implementation in Go with batteries included (built-in database and HTTP server). The project implements a complete interpreter following the classic compiler architecture pattern.

**Current Status**: Phases 1-3 complete (Lexer, Parser, Interpreter). Phase 4 (CLI) in progress.

## Commands

### Development Workflow
```bash
# Run all checks (lint + test)
task check

# Run linter only
task lint

# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests for a specific package
go test ./internal/interpreter

# Run a single test
go test ./internal/parser -run TestVarStatements

# Run tests with coverage report
task test:coverage
# Opens coverage.html in browser

# Build CLI
task build
# Output: bin/tsl

# Run a TotalScript file (after building)
./bin/tsl examples/hello.tsl
```

### Code Quality
```bash
# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Install development tools
task install:tools
```

## Architecture

### Core Pipeline: Source → Tokens → AST → Execution

The interpreter follows a three-stage pipeline:

1. **Lexer** (`internal/lexer/`) - Converts source code string into tokens
2. **Parser** (`internal/parser/`) - Converts tokens into Abstract Syntax Tree (AST)
3. **Interpreter** (`internal/interpreter/`) - Tree-walking evaluation of AST nodes

### Package Structure

```
internal/
├── token/          # Token type definitions and keyword lookup
├── lexer/          # Lexical analysis (string → tokens)
├── ast/            # AST node types and interfaces
├── parser/         # Recursive descent parser (tokens → AST)
├── interpreter/    # Tree-walking interpreter (AST → execution)
│   ├── object.go       # Runtime value types (Integer, String, Function, etc.)
│   ├── environment.go  # Variable scoping (lexical environments)
│   └── interpreter.go  # Eval() function and expression evaluation
├── stdlib/         # Built-in functions (future)
├── database/       # SQLite integration (future)
└── http/           # HTTP server/client (future)

cmd/
├── tsl/            # CLI interpreter binary
└── tsl-lsp/        # LSP server (future)
```

### Key Architectural Patterns

#### 1. Token-based Lexing
The lexer (`internal/lexer/lexer.go`) maintains:
- `position` and `readPosition` for two-character lookahead
- `line` and `column` for error reporting
- Token location tracking on every token

**Important**: String reading (`readString()`) must consume the closing quote character - this was a bug that was fixed.

#### 2. Pratt Parser with Precedence Climbing
Parser (`internal/parser/parser.go`) uses:
- **Prefix parse functions**: Handle tokens that start expressions (literals, identifiers, `-`, `!`, `(`, etc.)
- **Infix parse functions**: Handle binary operators (`+`, `-`, `*`, etc.)
- **Precedence table**: Maps token types to precedence levels (LOWEST → CALL → INDEX → MEMBER)

Register parse functions in `New()`:
```go
p.registerPrefix(token.INTEGER, p.parseIntegerLiteral)
p.registerInfix(token.PLUS, p.parseInfixExpression)
```

#### 3. Object System
Runtime values (`internal/interpreter/object.go`) implement the `Object` interface:
```go
type Object interface {
    Type() ObjectType
    Inspect() string
}
```

**Singleton pattern** for common values:
- `NULL`, `TRUE`, `FALSE`, `BREAK`, `CONTINUE` are package-level singletons
- Use `nativeBoolToBooleanObject()` to get correct singleton

#### 4. Environment-based Scoping
Variables are stored in linked `Environment` structures:
- Each scope creates a new environment with `NewEnclosedEnvironment(outer)`
- Lookup walks the chain: current → outer → outer.outer
- Functions capture their creation environment (closures)

#### 5. Special Control Flow Objects
Loop control uses runtime objects:
- `BREAK` and `CONTINUE` are propagated up through evaluation
- `evalBlockStatement()` checks for these and returns them immediately
- Loop evaluators (`evalWhileStatement`, etc.) handle them appropriately

### Parser Extension Pattern

To add new syntax:

1. **Add tokens** to `internal/token/token.go`
2. **Update lexer** in `internal/lexer/lexer.go` to recognize new tokens
3. **Create AST node** in `internal/ast/ast.go` implementing `Expression` or `Statement`
4. **Register parse function** in parser's `New()`
5. **Add case** to `Eval()` in `internal/interpreter/interpreter.go`

Example: The range operator (`..` and `..=`) required:
- Tokens: `DOTDOT`, `DOTDOTEQ`
- AST: `RangeExpression` with `Start`, `End`, `Inclusive`
- Parser: `parseRangeExpression()` as infix with RANGE precedence
- Interpreter: `evalRangeExpression()` creates `Array` of integers

### Error Handling Pattern

Follow these conventions:

1. **Parser errors**: Use `ParseError` with line/column from token
   ```go
   p.addError("expected next token to be X")
   ```

2. **Interpreter errors**: Return `*Error` object
   ```go
   return newError("division by zero")
   ```

3. **Check errors immediately**: After every `Eval()` call:
   ```go
   val := Eval(node, env)
   if IsError(val) {
       return val  // Propagate up
   }
   ```

### Testing Patterns

All packages use **table-driven tests**:

```go
tests := []struct {
    name     string
    input    string
    expected interface{}
}{
    {"case 1", "input 1", result1},
    {"case 2", "input 2", result2},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

Interpreter tests use helper functions:
- `testEval(input)` - Lex → Parse → Eval pipeline
- `testIntegerObject(t, obj, expected)` - Type-safe assertion
- `testBooleanObject()`, `testNullObject()`, etc.

## Language Features Reference

See `specification.md` for complete language definition. Key implementation notes:

- **Integer division**: `/` returns float, `//` returns integer
- **Ranges**: `0..10` (exclusive), `0..=10` (inclusive) - both create Array objects
- **Negative indexing**: `arr[-1]` gets last element
- **Map keys**: Currently only strings (maps use `map[string]Object` internally)
- **Assignment operators**: `=`, `+=`, `-=`, `*=`, `/=`, `%=` all work on identifiers

## Development Standards

See `DEVELOPMENT.md` for detailed coding standards. Key points:

- **Error types**: Each package defines its own error type with location info
- **Documentation**: All exported types, functions, and methods must have doc comments
- **Testing**: Table-driven tests preferred, use `t.Helper()` in test utilities
- **Commit messages**: Follow conventional commits (feat, fix, refactor, test, docs)

## Pre-Commit Checklist

**CRITICAL**: Always run tests and linter before committing code. Never commit code that doesn't pass these checks.

```bash
# Run both tests and linter (recommended)
task check

# Or run individually:
go test ./...           # All tests must pass
golangci-lint run ./... # No errors allowed
```

If tests fail or linter reports errors, fix them before committing. This ensures:
- No broken functionality is committed
- Code quality standards are maintained
- Other developers don't encounter issues when pulling changes

## Known Limitations

Current implementation limitations to be aware of:

1. **Assignment**: Only simple assignment works (not `arr[0] = val` or `obj.field = val`)
2. **Member access**: `.` operator parsed but not implemented in interpreter
3. **Type system**: No type checking - type annotations are parsed but ignored
4. **Models/Enums**: Not yet implemented
5. **Built-in functions**: Only language constructs work, no `println()`, `typeof()`, etc.

## Next Steps (Phase 4+)

When continuing development:

1. **Phase 4 - CLI**: Create `cmd/tsl/main.go` to run `.tsl` files
2. **Phase 5 - Built-ins**: Implement `internal/stdlib/` with `println()`, type conversions, string/array methods
3. **Phase 6 - Models**: Add model definition and instantiation, `this` keyword
4. **Phase 7 - Advanced**: Union types, optional types, imports
5. **Phase 8 - Database**: SQLite integration with `db` global
6. **Phase 9 - HTTP**: Server and client with `server` and `client` globals
