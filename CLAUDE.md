# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TotalScript is a scripting language implementation in Go with batteries included (built-in database and HTTP server). The project implements a complete interpreter following the classic compiler architecture pattern.

**Current Status**: Phases 1-6 complete! Core language with models, enums, and standard library (~70% of specification implemented).

**Implementation Progress**:
- ✅ **Phase 1**: Lexer - All tokens, comments, string escapes (100%)
- ✅ **Phase 2**: Parser - AST nodes for core features (100%)
- ✅ **Phase 3**: Interpreter - Core evaluation, control flow, closures (100%)
- ✅ **Phase 4**: CLI - File execution with tsl binary (100%)
- ✅ **Phase 5**: Built-ins - stdlib functions and methods (100%)
- ✅ **Phase 6**: Models & Enums - User-defined types (100% spec compliant)
- ❌ **Phase 7**: Advanced Types - Union types, optional types (0%)
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

### Type System
- **Type annotations**: Parsed for documentation (not enforced at runtime)
- **Type checking with `is`**: Works for models, enums, and built-in types
  - Built-in types: `value is integer`, `value is string`, `value is boolean`, `value is float`, `value is null`, `value is array`, `value is map`
  - User types: `instance is Point`, `status is HttpStatus`
- **Union types**: Syntax parsed (e.g., `integer | string`) but not enforced
- **Optional types**: Syntax parsed (e.g., `string?`) but not enforced

## What's Partially Working ⚠️

Features that are parsed but not fully functional:

1. **Type annotations**
   - Status: Fully parsed (simple types, unions, optionals, generics)
   - Issue: Not enforced or validated at runtime
   - Impact: No type safety, documentation only
   - Location: Type information in AST but unused in evaluation

## What's Missing ❌

Features defined in `specification.md` but not yet implemented:

### Collection Limitations
- **Index assignment**: Cannot do `arr[0] = val` or `map["key"] = val` (collections effectively immutable via index)
- **Array slicing**: Cannot do `arr[1..3]`, `arr[2..]`, `arr[..3]`

### Advanced Features
- **Union types enforcement**: Syntax parsed but not validated at runtime
- **Optional types enforcement**: Syntax parsed but not validated at runtime
- **Function overloading**: Multiple signatures with same name
- **Modules**: `import` statement, module system, qualified access
- **Standard library modules**: No `math`, `json`, `fs`, `time`, `os`, `crypto` modules
- **Database**: No `db` object, no SQLite integration, no persistence
- **HTTP**: No `server` or `client` objects, no Request/Response types

## Known Limitations

Current implementation limitations to be aware of:

### Assignment Limitations
1. **Only simple identifiers**: Can only assign to variables like `x = 5`
2. **No index assignment**: Cannot do `arr[0] = val` or `map["key"] = val` (must use `.push()`, `.insert()`, etc.)
3. **No member assignment**: Cannot do `obj.field = val` (model fields are set at construction)
4. **Workaround**: Use array/map methods (`.push()`, `.insert()`, `.remove()`) or recreate objects

### Expression Limitations
5. **No array slicing**: Only single index access works (`arr[0]`), not ranges (`arr[1..3]`)

### Type System Limitations
6. **Type annotations not enforced**: Parsed for documentation but no runtime validation
7. **No type narrowing**: `is` operator checks type but doesn't affect subsequent code flow
8. **No generics enforcement**: Generic syntax `array<integer>` parsed but not enforced
9. **No union type enforcement**: Syntax parsed but not validated
10. **No optional type enforcement**: Syntax parsed but not validated

### Missing Features
11. **No standard library modules**: No `math`, `json`, `fs`, `time`, `os`, `crypto` modules
12. **No modules/imports**: Cannot split code across multiple files
13. **No database integration**: No built-in SQLite support
14. **No HTTP support**: No built-in HTTP server or client

## Specification Compliance

**Compliance Status**: ✅ **100% compliant for implemented features**

All implemented features correctly follow `specification.md`. There are no deviations or specification violations. The implementation uses a phased approach:

- **Phases 1-6** (Core Language, Built-ins, Models & Enums): ✅ Complete and spec-compliant
- **Phases 7-10** (Advanced features): ❌ Not yet started

### Feature Coverage Matrix

| Category | Spec Coverage | Status |
|----------|---------------|--------|
| **Primitive Types** | 100% | ✅ Complete |
| **Operators** | 100% | ✅ Complete |
| **Control Flow** | 100% | ✅ Complete |
| **Functions** | 95% | ✅ Complete (no overloading) |
| **Collections** | 90% | ✅ Complete (no slicing, no index assignment) |
| **Type System** | 70% | ⚠️ `is` works, annotations not enforced |
| **Built-in Functions** | 100% | ✅ Complete |
| **String Methods** | 100% | ✅ Complete |
| **Array Methods** | 100% | ✅ Complete |
| **Map Methods** | 100% | ✅ Complete |
| **Models** | 100% | ✅ Complete (spec compliant) |
| **Enums** | 100% | ✅ Complete (spec compliant) |
| **Modules** | 0% | ❌ Not implemented |
| **Database** | 0% | ❌ Not implemented |
| **HTTP** | 0% | ❌ Not implemented |
| **Overall** | ~70% | ✅ Core Complete |

### Testing Coverage

All implemented features have comprehensive test suites:
- `internal/token/token_test.go` - Token recognition
- `internal/lexer/lexer_test.go` - Lexical analysis
- `internal/parser/parser_test.go` - Parsing all constructs
- `internal/interpreter/interpreter_test.go` - Runtime evaluation

Test execution: `go test ./...` - All tests pass ✅

## Next Steps (Phase 7+)

### Phase 7: Advanced Type System
**Priority**: MEDIUM - Type safety improvements

Features to implement:
1. **Union type enforcement**: Runtime validation of `integer | string` types
2. **Optional type enforcement**: Runtime validation of `string?` types
3. **Type narrowing**: `is` operator affects subsequent code flow
4. **Generic type enforcement**: Validate `array<integer>` at runtime

Implementation notes:
- Extend `evalVarStatement` and `evalConstStatement` to validate types
- Add type checking in assignment operations
- Implement type guards for narrowing

### Phase 8: Collection Assignment & Slicing
**Priority**: HIGH - Makes collections more practical

Features:
1. **Index assignment**: `arr[0] = newValue`, `map["key"] = newValue`
2. **Member assignment**: `obj.field = newValue` for model instances
3. **Array slicing**: `arr[1..3]`, `arr[2..]`, `arr[..3]`
4. Modify `evalAssignmentExpression()` to handle `IndexExpression` and `MemberExpression` as left-hand side

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
4. **Phase 7**: Advanced type system (union types, optional types, type narrowing)
5. **Phase 8**: Collection assignment and slicing
6. **Phase 9**: Modules and imports
7. **Phase 10**: Database integration
8. **Phase 11**: HTTP server and client

**Current Status**: Core language complete (Phases 1-6). Ready for advanced features.
