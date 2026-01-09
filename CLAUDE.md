# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TotalScript is a scripting language implementation in Go with batteries included (database and HTTP server modules). The project implements a complete interpreter following the classic compiler architecture pattern.

**Current Status**: Phases 1-9 complete! Core language with models, enums, type enforcement, collections, and modules with full stdlib (~90% of specification implemented).

**Implementation Progress**:
- ✅ **Phase 1**: Lexer - All tokens, comments, string escapes (100%)
- ✅ **Phase 2**: Parser - AST nodes for core features (100%)
- ✅ **Phase 3**: Interpreter - Core evaluation, control flow, closures (100%)
- ✅ **Phase 4**: CLI - File execution with tsl binary (100%)
- ✅ **Phase 5**: Built-ins - stdlib functions and methods (100%)
- ✅ **Phase 6**: Models & Enums - User-defined types (100% spec compliant)
- ✅ **Phase 7**: Type Enforcement - Union types, optional types, generics, mixed-type arithmetic (100% complete)
- ✅ **Phase 8**: Collection Assignment & Slicing - Index/member assignment, array slicing (100%)
- ✅ **Phase 9**: Modules & Standard Library - Import system with math, json, fs, time, os modules (100%, 5/6 stdlib modules)
- ❌ **Phase 10**: Database - SQLite integration (0%)
- ❌ **Phase 11**: HTTP - Server and client (0%)

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
│   ├── object.go       # Runtime value types (Integer, String, Function, Module, etc.)
│   ├── environment.go  # Variable scoping (lexical environments)
│   ├── interpreter.go  # Eval() function and expression evaluation
│   └── module.go       # Module loading and stdlib (math, json, fs, time, os, db, http)
└── stdlib/         # Built-in functions (println, typeof, conversions)

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
- **Variables**: `var` declarations with optional type annotations (enforced at declaration)
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
- **Array index assignment**: `arr[0] = value`, `arr[-1] = value` with compound operators (`+=`, `-=`, etc.)
- **Map index assignment**: `map["key"] = value` including new key creation
- **Model field assignment**: `obj.field = value` with compound operators
- **Array slicing**: `arr[1..3]` (exclusive), `arr[1..=3]` (inclusive), `arr[2..]`, `arr[..3]`, `arr[..]`
- **Negative index slicing**: `arr[-3..-1]`, `arr[-3..]`, `arr[..-3]`

### Built-in Functions & Standard Library
- **Output**: `println()` for console output
- **Type conversions**: `integer()`, `float()`, `string()`, `boolean()` convert between types
- **Type inspection**: `typeof()` returns type name as string
- **String methods**: `.length()`, `.upper()`, `.lower()`, `.split()`, `.trim()`, `.contains()`, `.replace()`, `.substring()`
- **Array methods**: `.length()`, `.push()`, `.pop()`, `.insert()`, `.remove()`, `.contains()`, `.indexOf()`, `.join()`, `.map()`, `.filter()`, `.reduce()`
- **Map methods**: `.length()`, `.keys()`, `.values()`, `.contains()`, `.remove()`

### Models (User-Defined Types)
- **Model definition**: `const Point = model { x: float, y: float }`
- **Model instantiation**: `var p = Point(3, 4)` using default constructor
- **Multiple constructors**: Custom constructors with different parameter counts
- **Model methods**: Functions inside models with `this` keyword support
- **Nested models**: Models can contain other models as fields
- **Field access**: `p.x`, `p.y` access model fields
- **Method calls**: `p.distance()` calls model methods
- **Built-in Error model**: `Error("message")` for error handling

### Enums (Enumeration Types)
- **Enum definition**: `const Status = enum { OK = 200, NotFound = 404 }`
- **Enum values**: Integer, string, and boolean underlying types
- **Enum member access**: `Status.OK` returns enum value
- **Enum value property**: `status.value` returns underlying value (200)
- **Enum methods**: `.values()` returns all values, `.fromValue(n)` finds by value
- **Enum comparison**: `status == Status.OK` works correctly

### Type System & Type Enforcement
- **Type annotations**: Fully enforced at runtime for `var` and `const` declarations
- **Type checking with `is`**: Works for models, enums, and built-in types
  - Built-in types: `value is integer`, `value is string`, `value is boolean`, `value is float`, `value is null`, `value is array`, `value is map`
  - User types: `instance is Point`, `status is HttpStatus`
- **Union types**: `integer | string` - Enforced at runtime, validates value matches one of the types
- **Optional types**: `string?` - Enforced at runtime, allows null or specified type
- **Generic types**: `array<integer>` - Enforced at runtime, validates all elements match type
- **Union types in generics**: `array<integer | string>` - Enforced at runtime, allows mixed-type arrays
- **Mixed-type arithmetic**: Integer and float can be mixed in arithmetic operations (`2 ** 0.5`)
- **Type validation**: Variables and constants are validated against their declared types on assignment
- **Automatic type coercion**: Integers automatically convert to floats when float type is expected (variables, parameters, model fields, array elements)

### Modules & Imports
- **Import statement**: `import "math"` for stdlib, `import "./utils"` for local files
- **Import aliases**: `import "./geometry" as geo` for custom namespace
- **Qualified access**: `math.PI`, `math.sqrt(16)`, `geo.circleArea(5)`
- **Module caching**: Modules are loaded once and cached across imports
- **Relative imports**: `./file`, `./lib/helpers` resolved relative to importing file
- **File modules**: Each .tsl file is a module, all top-level declarations automatically exported

### Standard Library Modules
- **math**: Mathematical functions (abs, min, max, floor, ceil, round, sqrt, pow, sin, cos, tan, log, log10) and constants (PI, E)
- **json**: JSON parsing and serialization (parse, stringify) with full type conversion
- **fs**: File system operations (readFile, writeFile, exists, listDir)
- **time**: Time operations (now, sleep) - timestamps in milliseconds
- **os**: Operating system utilities (env, args) - environment variables and command-line arguments

## What's Partially Working ⚠️

No partially working features currently - all implemented features are fully functional!

## What's Missing ❌

Features defined in `specification.md` but not yet implemented:

### Advanced Features
- **Type narrowing**: `is` operator checks type but doesn't affect subsequent code flow
- **Crypto module**: Hash functions and encryption utilities not yet implemented
- **Database module**: No `db` module implementation, no SQLite integration, no persistence
- **HTTP module**: No `http` module implementation (http.server and http.client), no Request/Response types

## Known Limitations

Current implementation limitations to be aware of:

### Missing Features
1. **Crypto module**: Hash functions and encryption not yet implemented
2. **Database module**: No `db` module implementation, no SQLite support
3. **HTTP module**: No `http` module implementation with server and client functionality

## Specification Compliance

**Compliance Status**: ✅ **100% compliant for implemented features**

All implemented features correctly follow `specification.md`. There are no deviations or specification violations. The implementation uses a phased approach:

- **Phases 1-9** (Core Language, Built-ins, Models & Enums, Type Enforcement, Assignment & Slicing, Modules): ✅ Complete and spec-compliant
- **Phases 10-11** (Database, HTTP): ❌ Not yet started

### Feature Coverage Matrix

| Category | Spec Coverage | Status |
|----------|---------------|--------|
| **Primitive Types** | 100% | ✅ Complete |
| **Operators** | 100% | ✅ Complete |
| **Control Flow** | 100% | ✅ Complete |
| **Functions** | 100% | ✅ Complete |
| **Collections** | 100% | ✅ Complete (full assignment & slicing) |
| **Type System** | 90% | ✅ Enforced for var/const (not for reassignments) |
| **Built-in Functions** | 100% | ✅ Complete |
| **String Methods** | 100% | ✅ Complete |
| **Array Methods** | 100% | ✅ Complete |
| **Map Methods** | 100% | ✅ Complete |
| **Models** | 100% | ✅ Complete (spec compliant) |
| **Enums** | 100% | ✅ Complete (spec compliant) |
| **Modules** | 100% | ✅ Complete (import system) |
| **Standard Library** | 83% | ✅ 5/6 modules (missing crypto) |
| **Database** | 0% | ❌ Not implemented |
| **HTTP** | 0% | ❌ Not implemented |
| **Overall** | ~90% | ✅ Core Complete |

### Testing Coverage

All implemented features have comprehensive test suites:
- `internal/token/token_test.go` - Token recognition
- `internal/lexer/lexer_test.go` - Lexical analysis
- `internal/parser/parser_test.go` - Parsing all constructs
- `internal/interpreter/interpreter_test.go` - Runtime evaluation

Test execution: `go test ./...` - All tests pass ✅

## Next Steps (Phase 10+)

### Phase 9: Modules and Standard Library (✅ 100% Complete)

All planned features implemented:
- ✅ Import statement parsing (`import "math"`, `import "./file" as alias`)
- ✅ Module loading and caching
- ✅ Qualified access (`math.PI`, `math.sqrt()`)
- ✅ Relative path resolution for file modules
- ✅ **math** module: Mathematical functions and constants
- ✅ **json** module: parse() and stringify() with full type conversion
- ✅ **fs** module: readFile(), writeFile(), exists(), listDir()
- ✅ **time** module: now(), sleep()
- ✅ **os** module: env(), args()

Optional remaining work:
- **crypto** module: Hash functions and encryption (lower priority)

### Phase 7: Advanced Type System (✅ 100% Complete - COMPLETED EARLIER)

All planned features implemented:
- ✅ Union type enforcement (`integer | string`)
- ✅ Optional type enforcement (`string?`)
- ✅ Generic type enforcement (`array<integer>`)
- ✅ Union types in generics (`array<integer | string>`)
- ✅ Mixed-type arithmetic (`2 ** 0.5`)
- ✅ Type validation for `var` and `const` declarations
- ✅ Type validation on reassignments
- ✅ Function parameter type validation
- ✅ Automatic integer-to-float coercion

Optional remaining feature (not critical):
- Type narrowing: `is` operator affects subsequent code flow (requires control flow analysis)

### Phase 9b: Crypto Module (Optional)
**Priority**: LOW - Security utilities

Module to implement:
- `crypto` - hash functions (sha256, md5, etc.)

### Phase 10: Database Module
**Priority**: MEDIUM - Advanced feature

The `db` module provides SQLite integration through `import "db"`.

Features:
1. SQLite wrapper in `internal/interpreter/module.go` (createDBModule)
2. Module exports: `db.save()`, `db.find()`, `db.delete()`, `db.saveAll()`, `db.deleteAll()`, `db.configure()`, `db.transaction()`
3. Model persistence with automatic table creation
4. Query builder with pattern matching syntax
5. Transaction support
6. CLI argument: `--db=myapp.db` to specify database file

### Phase 11: HTTP Module
**Priority**: MEDIUM - Advanced feature

The `http` module provides HTTP server and client through `import "http"`.

Features:
1. HTTP server and client in `internal/interpreter/module.go` (createHTTPModule)
2. Module exports:
   - `http.server` - Server object with `.get()`, `.post()`, `.put()`, `.delete()`, `.start()`, `.static()`, `.use()`
   - `http.client` - Client object with `.get()`, `.post()`, `.put()`, `.patch()`, `.delete()`
   - `http.Request` - Request model type
   - `http.Response` - Response constructor function
3. Route handlers with path parameters (`:id`)
4. Middleware support
5. Static file serving

### Recommended Implementation Order

Based on dependencies and impact:

1. ✅ **Phases 1-4**: Core language (Lexer, Parser, Interpreter, CLI) - COMPLETE
2. ✅ **Phase 5**: Built-in functions and standard library - COMPLETE
   - Built-in functions: `println()`, `typeof()`, type conversions
   - String methods: `.length()`, `.upper()`, `.lower()`, `.split()`, etc.
   - Array methods: `.length()`, `.push()`, `.pop()`, `.map()`, `.filter()`, etc.
   - Map methods: `.length()`, `.keys()`, `.values()`, `.remove()`, etc.
3. ✅ **Phase 6**: Models & Enums - COMPLETE (100% spec compliant)
   - Models with multiple constructors
   - Model methods with `this` keyword
   - Enums with `.values()` and `.fromValue()` methods
   - Extended `is` operator for built-in types
   - Built-in Error model
4. ✅ **Phase 7**: Type Enforcement - COMPLETE (90%)
   - Union type enforcement (`integer | string`)
   - Optional type enforcement (`string?`)
   - Generic type enforcement (`array<integer>`)
   - Union types in generics (`array<integer | string>`)
   - Mixed-type arithmetic (`2 ** 0.5`)
   - Type validation for `var` and `const` declarations
5. ✅ **Phase 8**: Collection Assignment & Slicing - COMPLETE
   - Array index assignment: `arr[0] = value`, compound operators
   - Map index assignment: `map["key"] = value`, new key creation
   - Model field assignment: `obj.field = value`, compound operators
   - Array slicing: `arr[1..3]`, `arr[2..]`, `arr[..3]`, `arr[..]`
   - Negative index slicing: `arr[-3..-1]`, `arr[-3..]`, `arr[..-3]`
6. ✅ **Phase 9**: Modules and Standard Library - COMPLETE
   - Import statement parsing (`import "math"`, `import "./file" as alias`)
   - Module loading and caching
   - Standard library modules: math, json, fs, time, os
7. **Phase 10**: Database module (`import "db"`)
8. **Phase 11**: HTTP module (`import "http"`)

**Current Status**: Core language with modules complete (Phases 1-9). Ready for database and HTTP modules.
