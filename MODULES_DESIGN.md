# Module System Design

This document outlines the design and implementation plan for TotalScript's module system (Phase 9).

## Requirements Summary

From `specification.md`:

1. **Each `.tsl` file is a module** - All top-level declarations are automatically exported
2. **Import syntax**:
   - `import math` - Standard library module
   - `import ./utils` - Local file (relative path with `./`)
   - `import ./lib/helpers` - Nested relative paths
   - `import math as m` - Import with alias
3. **Qualified access** - Imported items accessed via module name: `math.sin()`, `geo.PI`
4. **Standard library modules**: math, json, time, fs, os, crypto

## Architecture Components

### 1. Tokens & Lexer

**Status**: ✅ Already complete

- `IMPORT` token defined (line 94, token.go)
- `AS` token defined (line 95, token.go)
- Both keywords registered in keywords map
- Lexer already handles string literals for file paths

**No changes needed**.

### 2. AST Nodes

**New node**: `ImportStatement`

```go
// ImportStatement represents an import declaration.
// Examples:
//   import math
//   import ./utils
//   import ./geometry as geo
type ImportStatement struct {
    Token      token.Token // The 'import' token
    Path       string      // Module path: "math", "./utils", "./lib/helpers"
    Alias      string      // Optional alias from 'as' clause, "" if none
    ModuleName string      // Computed module name: "math", "utils", "helpers"
}
```

**Field details**:
- `Path`: Raw import path from source code
- `Alias`: User-specified alias via `as` clause
- `ModuleName`: Computed name for qualified access
  - For stdlib: use path as-is (`"math"` → `"math"`)
  - For files: use filename without extension (`"./utils"` → `"utils"`, `"./lib/helpers"` → `"helpers"`)
  - Overridden by `Alias` if provided

### 3. Parser

**Function**: `parseImportStatement()`

```go
func (p *Parser) parseImportStatement() *ast.ImportStatement {
    // Consume 'import' token
    // Parse path (IDENT for stdlib, STRING for file paths)
    // Check for optional 'as' clause
    // Compute ModuleName from path and alias
}
```

**Parsing logic**:

```tsl
import math         # Path: "math", Alias: "", ModuleName: "math"
import ./utils      # Path: "./utils", Alias: "", ModuleName: "utils"
import math as m    # Path: "math", Alias: "m", ModuleName: "m"
```

**Path parsing**:
- Stdlib modules: next token must be IDENT
- File modules: next token must be STRING or DOT (for relative paths)
- For relative paths starting with `./`, parse as special identifier sequence or use string literal

**Note**: To simplify, we'll require file paths to be string literals:
```tsl
import "math"           # stdlib (quoted)
import "./utils"        # file (quoted with ./)
import "./geometry" as geo
```

### 4. Interpreter - Module Object

**New object type**: `Module`

```go
const MODULE_OBJ = "MODULE"

type Module struct {
    Name   string               // Module name
    Scope  *Environment         // Module's exported scope
}

func (m *Module) Type() ObjectType { return MODULE_OBJ }
func (m *Module) Inspect() string  { return "module " + m.Name }
```

**Module exports**:
- Each module has its own `Environment` containing exported values
- All top-level `var`, `const`, `function`, `model`, `enum` declarations are automatically exported
- Accessing module members: `module.Get(name)` from module's Scope

### 5. Interpreter - Module Loading

**Module cache** (prevent re-evaluation):

```go
type ModuleCache struct {
    modules map[string]*Module  // path -> Module
}

var moduleCache = &ModuleCache{modules: make(map[string]*Module)}
```

**Module resolution**:

```go
func resolveModule(path string, currentFile string) (*Module, Object)
```

1. Check cache first - return if already loaded
2. Determine module type (stdlib vs file)
3. For stdlib: call `loadStdlibModule(name)`
4. For files: call `loadFileModule(path, currentFile)`
5. Cache result and return

**Stdlib loading**:

```go
func loadStdlibModule(name string) (*Module, Object) {
    switch name {
    case "math":
        return createMathModule()
    case "json":
        return createJsonModule()
    // ... other stdlib modules
    default:
        return nil, newError("unknown stdlib module: %s", name)
    }
}
```

**File loading**:

```go
func loadFileModule(path string, currentFile string) (*Module, Object) {
    // 1. Resolve relative path to absolute
    // 2. Read file contents
    // 3. Lex and parse
    // 4. Create new environment for module
    // 5. Evaluate in module environment
    // 6. Create Module object with environment
    // 7. Return Module
}
```

### 6. Interpreter - Import Statement Evaluation

**New case in Eval()**:

```go
case *ast.ImportStatement:
    return evalImportStatement(node, env)
```

**evalImportStatement()**:

```go
func evalImportStatement(node *ast.ImportStatement, env *Environment) Object {
    // 1. Resolve and load module
    module, err := resolveModule(node.Path, env.currentFile)
    if err != nil {
        return err
    }

    // 2. Store module in environment under ModuleName
    moduleName := node.ModuleName
    env.Set(moduleName, module)

    return NULL  // Import statements don't produce values
}
```

### 7. Interpreter - Qualified Access

**Member access on modules**:

Modify `evalMemberExpression()` to handle modules:

```go
func evalMemberExpression(node *ast.MemberExpression, env *Environment) Object {
    left := Eval(node.Object, env)
    if IsError(left) {
        return left
    }

    // Existing: Model instances
    if instance, ok := left.(*ModelInstance); ok {
        // ... existing logic
    }

    // NEW: Module access
    if module, ok := left.(*Module); ok {
        memberName := node.Member.Value
        value, exists := module.Scope.Get(memberName)
        if !exists {
            return newError("module '%s' has no member '%s'", module.Name, memberName)
        }
        return value
    }

    return newError("member access not supported on %s", left.Type())
}
```

### 8. Standard Library Modules

**Create `internal/stdlib/` directory** with module implementations:

```
internal/stdlib/
├── math.go      # Math module
├── json.go      # JSON module
├── fs.go        # File system module
├── time.go      # Time module
├── os.go        # OS module
└── crypto.go    # Crypto module
```

**Example: math module** (`internal/stdlib/math.go`):

```go
package stdlib

import "math"

func CreateMathModule() *Module {
    env := NewEnvironment()

    // Constants
    env.Set("PI", &Float{Value: math.Pi})
    env.Set("E", &Float{Value: math.E})

    // Functions
    env.Set("abs", &Builtin{
        Fn: func(args ...Object) Object {
            if len(args) != 1 {
                return newError("abs() takes exactly 1 argument")
            }
            // Implement abs logic
        },
    })

    env.Set("sqrt", &Builtin{...})
    // ... more functions

    return &Module{Name: "math", Scope: env}
}
```

## Implementation Plan

### Step 1: AST Changes
- Add `ImportStatement` to `internal/ast/ast.go`
- Implement `TokenLiteral()` and `String()` methods

### Step 2: Parser Changes
- Add `parseImportStatement()` to `internal/parser/parser.go`
- Update `parseStatement()` to handle IMPORT token
- Add parser tests

### Step 3: Interpreter - Module Object
- Add `Module` type to `internal/interpreter/object.go`
- Add `MODULE_OBJ` constant

### Step 4: Interpreter - Module Loading
- Create `internal/interpreter/module.go` with:
  - `ModuleCache` type
  - `resolveModule()` function
  - `loadStdlibModule()` function
  - `loadFileModule()` function
- Update `Environment` to track current file path (for relative imports)

### Step 5: Interpreter - Import Evaluation
- Add `evalImportStatement()` to `internal/interpreter/interpreter.go`
- Add case for `*ast.ImportStatement` in `Eval()`
- Update `evalMemberExpression()` to handle module access

### Step 6: Standard Library - Math Module
- Create `internal/stdlib/math.go`
- Implement math constants (PI, E)
- Implement math functions (abs, min, max, floor, ceil, round, sqrt, pow, sin, cos, tan, log, log10)

### Step 7: Standard Library - Other Modules
- `json.go` - parse(), stringify()
- `fs.go` - readFile(), writeFile(), exists(), listDir()
- `time.go` - now(), sleep()
- `os.go` - env(), args()
- `crypto.go` - hash functions

### Step 8: Testing
- Parser tests for import statements
- Interpreter tests for module loading
- Interpreter tests for qualified access
- Stdlib module tests

### Step 9: Examples
- Create `examples/modules/` directory
- Create example module files
- Create example using stdlib modules

## Path Resolution Algorithm

For file imports:

```go
func resolveFilePath(importPath string, currentFile string) (string, error) {
    // importPath examples: "./utils", "./lib/helpers"
    // currentFile example: "/Users/project/src/main.tsl"

    if !strings.HasPrefix(importPath, "./") {
        return "", errors.New("file imports must start with './'")
    }

    // Remove "./" prefix
    relativePath := strings.TrimPrefix(importPath, "./")

    // Get directory of current file
    currentDir := filepath.Dir(currentFile)

    // Resolve relative path
    modulePath := filepath.Join(currentDir, relativePath)

    // Add .tsl extension if not present
    if !strings.HasSuffix(modulePath, ".tsl") {
        modulePath += ".tsl"
    }

    // Verify file exists
    if _, err := os.Stat(modulePath); os.IsNotExist(err) {
        return "", fmt.Errorf("module file not found: %s", modulePath)
    }

    return modulePath, nil
}
```

## Error Handling

Import errors to handle:

1. **Module not found**: Stdlib module doesn't exist or file doesn't exist
2. **Circular imports**: Module A imports B, B imports A (detect with loading stack)
3. **Parse errors**: Imported file has syntax errors
4. **Runtime errors**: Imported file has runtime errors during evaluation
5. **Member not found**: Accessing non-existent module member

## Future Enhancements (Not in Phase 9)

These are ideas for future improvement:

1. **Selective imports**: `from math import sin, cos`
2. **Re-exports**: Allow modules to re-export imported items
3. **Module path configuration**: Allow custom module search paths
4. **Precompiled modules**: Cache parsed AST or bytecode
5. **Private members**: Convention for private members (e.g., `_private`)

## Testing Strategy

### Parser Tests

```go
func TestImportStatements(t *testing.T) {
    tests := []struct {
        input           string
        expectedPath    string
        expectedAlias   string
        expectedModule  string
    }{
        {`import "math"`, "math", "", "math"},
        {`import "./utils"`, "./utils", "", "utils"},
        {`import "math" as m`, "math", "m", "m"},
        {`import "./geometry" as geo`, "./geometry", "geo", "geo"},
    }
    // ...
}
```

### Interpreter Tests

```go
func TestModuleLoading(t *testing.T) {
    // Test loading stdlib module
    // Test loading file module
    // Test module caching
    // Test error cases
}

func TestQualifiedAccess(t *testing.T) {
    // Test accessing module constants
    // Test accessing module functions
    // Test accessing module models
    // Test error for non-existent members
}
```

### Integration Tests

Create test files:
- `testdata/modules/utils.tsl` - Simple utility module
- `testdata/modules/geometry.tsl` - Module with models and functions
- `testdata/modules/main.tsl` - Main file that imports the above

## CLI Changes

Update `cmd/tsl/main.go` to:

1. Pass current file path to interpreter
2. Handle module loading errors gracefully
3. Display helpful error messages for import issues

## Compatibility

This implementation maintains backward compatibility:
- Existing scripts without imports continue to work
- No changes to existing language features
- Module system is purely additive

## Summary

The module system adds:
- `ImportStatement` AST node
- `Module` object type
- Module loading and caching
- Qualified member access on modules
- Standard library modules in `internal/stdlib/`

All while maintaining the simplicity and clarity of TotalScript's design.
