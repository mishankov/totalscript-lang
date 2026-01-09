# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TotalScript is a scripting language implementation in Go with batteries included (built-in database and HTTP server). The project implements a complete interpreter following the classic compiler architecture pattern.

**Current Status**: Phases 1-4 complete (Lexer, Parser, Interpreter, CLI). Core language working (~45% of specification implemented).

**Implementation Progress**:
- ✅ **Phase 1**: Lexer - All tokens, comments, string escapes (100%)
- ✅ **Phase 2**: Parser - AST nodes for core features (95%)
- ✅ **Phase 3**: Interpreter - Core evaluation, control flow, closures (95%)
- ✅ **Phase 4**: CLI - File execution with tsl binary (100%)
- ⚠️ **Phase 5**: Built-ins - stdlib functions and methods (0%)
- ❌ **Phase 6**: Models - User-defined types (0%)
- ❌ **Phase 7**: Enums - Enumeration types (0%)
- ❌ **Phase 8**: Modules - Import system (0%)
- ❌ **Phase 9**: Database - SQLite integration (0%)
- ❌ **Phase 10**: HTTP - Server and client (0%)

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

## What's Fully Working ✅

The following features from `specification.md` are fully implemented and tested:

### Core Language
- **Primitive types**: `integer`, `float`, `string`, `boolean`, `null`
- **Operators**: All arithmetic (`+`, `-`, `*`, `/`, `//`, `%`, `**`), comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical (`&&`, `||`, `!`), and assignment (`=`, `+=`, `-=`, `*=`, `/=`, `%=`)
- **Variables**: `var` declarations with optional type annotations (annotations parsed but not enforced)
- **Constants**: `const` declarations

### Control Flow
- **Conditionals**: `if`/`else if`/`else` (both statement and expression forms)
- **Switch**: Multiple case values, default clause, expression evaluation
- **Loops**: `while`, `for-in` (with value or index+value), C-style `for`
- **Loop control**: `break` and `continue` work correctly
- **Range operators**: `0..10` (exclusive), `0..=10` (inclusive) create arrays

### Functions & Closures
- **Function literals**: First-class functions with `function(params) { body }` syntax
- **Closures**: Functions capture lexical environment correctly
- **Higher-order functions**: Functions can accept and return functions
- **Nested functions**: Inner functions can access outer scope variables
- **Anonymous functions**: Immediate function invocation works

### Collections
- **Array literals**: `[1, 2, 3]` including empty arrays `[]`
- **Map literals**: `{"key": value}` including empty maps `{}`
- **Array indexing**: `arr[0]`, `arr[-1]` (negative indices from end)
- **Map indexing**: `map["key"]` returns value or NULL if missing
- **Heterogeneous collections**: Arrays and maps can hold mixed types

### Type System (Partial)
- **Type annotations**: Parsed for documentation (not enforced)
- **Type checking**: `is` operator works for basic type checking
- **Union types**: Syntax parsed (e.g., `integer | string`)
- **Optional types**: Syntax parsed (e.g., `string?`)

## What's Partially Working ⚠️

Features that are parsed but not fully functional:

1. **Member expressions** (`.` operator)
   - Status: Parsed correctly by parser
   - Issue: `evalMemberExpression()` returns "not yet supported" error
   - Impact: Cannot access object properties, no method calls possible
   - Location: `internal/interpreter/interpreter.go:706-708`

2. **Type annotations**
   - Status: Fully parsed (simple types, unions, optionals, generics)
   - Issue: Not enforced or validated at runtime
   - Impact: No type safety, documentation only
   - Location: Type information in AST but unused in evaluation

3. **Error handling pattern**
   - Status: Error objects exist internally
   - Issue: No `Error()` constructor for user code
   - Impact: Cannot create `Error("message")` in TotalScript code
   - Workaround: Errors automatically created for runtime errors

## What's Missing ❌

Features defined in `specification.md` but not yet implemented:

### Critical for Usability
- **Built-in functions**: `println()`, `string()`, `integer()`, `float()`, `boolean()`, `typeof()`
- **String methods**: `.length()`, `.upper()`, `.lower()`, `.split()`, `.trim()`, `.contains()`, `.replace()`, `.substring()`
- **Array methods**: `.length()`, `.push()`, `.pop()`, `.insert()`, `.remove()`, `.contains()`, `.map()`, `.filter()`, `.reduce()`
- **Map methods**: `.length()`, `.keys()`, `.values()`, `.contains()`, `.remove()`
- **Index assignment**: Cannot do `arr[0] = val` or `map["key"] = val` (collections effectively immutable)
- **Array slicing**: Cannot do `arr[1..3]`, `arr[2..]`, `arr[..3]`

### Advanced Features
- **Models**: User-defined types, `this` keyword, methods, constructors, annotations
- **Enums**: Enumeration types with named values
- **Modules**: `import` statement, module system, qualified access
- **Standard library**: No `math`, `json`, `fs`, `time`, `os`, `crypto` modules
- **Database**: No `db` object, no SQLite integration, no persistence
- **HTTP**: No `server` or `client` objects, no Request/Response types
- **Function overloading**: Multiple signatures with same name
- **Error Model**: Cannot instantiate `Error()` in user code

## Known Limitations

Current implementation limitations to be aware of:

### Assignment Limitations
1. **Only simple identifiers**: Can only assign to variables like `x = 5`
2. **No index assignment**: Cannot do `arr[0] = val` or `map["key"] = val`
3. **No member assignment**: Cannot do `obj.field = val`
4. **Workaround**: Reassign entire collection or use future array/map methods

### Expression Limitations
5. **Member access not working**: `.` operator parsed but evaluation returns error
6. **No array slicing**: Only single index access works (`arr[0]`), not ranges (`arr[1..3]`)
7. **No method calls**: No `.method()` syntax since member access doesn't work

### Type System Limitations
8. **Type annotations ignored**: Parsed for documentation but no runtime validation
9. **No type narrowing**: `is` operator checks type but doesn't affect subsequent code
10. **No generics**: Generic syntax `array<integer>` parsed but not enforced

### Missing Core Features
11. **No built-in functions**: Must write all I/O and utilities from scratch
12. **No standard library**: No math, string manipulation, file I/O, etc.
13. **Collections are limited**: Can read but not easily modify arrays/maps
14. **No models or enums**: Only primitive types and collections available
15. **No modules**: Cannot split code across multiple files

## Specification Compliance

**Compliance Status**: ✅ **100% compliant for implemented features**

All implemented features correctly follow `specification.md`. There are no deviations or specification violations. The implementation uses a phased approach:

- **Phases 1-4** (Lexer, Parser, Interpreter, CLI): ✅ Complete and spec-compliant
- **Phases 5-11** (Advanced features): ❌ Not yet started

### Feature Coverage Matrix

| Category | Spec Coverage | Status |
|----------|---------------|--------|
| **Primitive Types** | 100% | ✅ Complete |
| **Operators** | 100% | ✅ Complete |
| **Control Flow** | 100% | ✅ Complete |
| **Functions** | 95% | ✅ Complete (no overloading) |
| **Collections** | 70% | ⚠️ Partial (no methods, no slicing) |
| **Type System** | 40% | ⚠️ Parsed only |
| **Built-in Functions** | 0% | ❌ Not implemented |
| **Models** | 0% | ❌ Not implemented |
| **Enums** | 0% | ❌ Not implemented |
| **Modules** | 0% | ❌ Not implemented |
| **Database** | 0% | ❌ Not implemented |
| **HTTP** | 0% | ❌ Not implemented |
| **Overall** | ~45% | ⚠️ In Progress |

### Testing Coverage

All implemented features have comprehensive test suites:
- `internal/token/token_test.go` - Token recognition
- `internal/lexer/lexer_test.go` - Lexical analysis
- `internal/parser/parser_test.go` - Parsing all constructs
- `internal/interpreter/interpreter_test.go` - Runtime evaluation

Test execution: `go test ./...` - All tests pass ✅

## Next Steps (Phase 5+)

### Phase 5: Built-in Functions & Standard Library
**Priority**: HIGH - Critical for usability

Files to create:
- `internal/stdlib/builtins.go` - Core functions: `println()`, `typeof()`, type conversions
- `internal/stdlib/string.go` - String methods
- `internal/stdlib/array.go` - Array methods
- `internal/stdlib/map.go` - Map methods

Key features:
1. **Output**: `println()` for console output
2. **Type conversions**: `integer()`, `float()`, `string()`, `boolean()`
3. **Type inspection**: `typeof()` returns type name
4. **String methods**: `length()`, `upper()`, `lower()`, `split()`, `trim()`, `contains()`, `replace()`, `substring()`
5. **Array methods**: `length()`, `push()`, `pop()`, `insert()`, `remove()`, `contains()`, `indexOf()`, `map()`, `filter()`, `reduce()`
6. **Map methods**: `length()`, `keys()`, `values()`, `contains()`, `remove()`

Implementation approach:
- Register built-in functions in global environment
- Member expressions need completion: `evalMemberExpression()` must dispatch to object methods
- Methods return new objects (functional approach) or modify in place (decide per spec)

### Phase 6: Assignment to Collections
**Priority**: HIGH - Makes collections practical

Features:
1. **Index assignment**: `arr[0] = newValue`, `map["key"] = newValue`
2. **Member assignment**: `obj.field = newValue` (needs models first)
3. Modify `evalAssignmentExpression()` to handle `IndexExpression` and `MemberExpression` as left-hand side

### Phase 7: Models
**Priority**: MEDIUM - Required for advanced features

Files:
- Update `internal/ast/ast.go` with `ModelLiteral` node
- Update `internal/parser/parser.go` with `parseModelLiteral()`
- Update `internal/interpreter/object.go` with `Model` and `ModelInstance` types
- Implement `this` keyword binding

Features:
1. Model definition: `model { fields... }`
2. Model instantiation: `Point(x, y)`
3. Model methods with `this` keyword
4. Multiple constructors with overloading
5. Field access via member expressions
6. Annotations: `@id` for primary keys

### Phase 8: Enums
**Priority**: LOW - Nice to have

Features:
1. Enum definition with typed values
2. Enum member access
3. `Enum.values()`, `Enum.fromValue()` methods
4. Switch on enum values

### Phase 9: Modules and Imports
**Priority**: MEDIUM - Code organization

Features:
1. `import` statement parsing
2. Module loading from file system
3. Module namespace and qualified access
4. Import aliases with `as`
5. Standard library modules: `math`, `json`, `fs`, `time`, `os`, `crypto`

### Phase 10: Database Integration
**Priority**: LOW - Advanced feature

Features:
1. SQLite wrapper in `internal/database/`
2. Global `db` object
3. Model persistence: `db.save()`, `db.find()`, `db.delete()`
4. Query builder with pattern matching
5. Transactions

### Phase 11: HTTP Server & Client
**Priority**: LOW - Advanced feature

Features:
1. HTTP server in `internal/http/server.go`
2. HTTP client in `internal/http/client.go`
3. Request/Response models
4. Route handlers: `server.get()`, `server.post()`, etc.
5. Middleware support
6. Static file serving

### Recommended Implementation Order

Based on dependencies and impact:

1. ✅ **Phases 1-4**: COMPLETE
2. 🔄 **Phase 5a**: Built-in functions (`println`, type conversions, `typeof`)
3. 🔄 **Phase 5b**: Complete member expression evaluation (fix `.` operator)
4. 🔄 **Phase 5c**: String/Array/Map methods
5. 🔄 **Phase 6**: Index assignment for collections
6. **Phase 7**: Models and `this` keyword
7. **Phase 8**: Enums
8. **Phase 9**: Modules and standard library
9. **Phase 10**: Database integration
10. **Phase 11**: HTTP server and client
