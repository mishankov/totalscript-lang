package interpreter

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mishankov/totalscript-lang/internal/ast"
	"github.com/mishankov/totalscript-lang/internal/lexer"
	"github.com/mishankov/totalscript-lang/internal/parser"
	_ "modernc.org/sqlite" // SQLite driver
)

// ModuleCache stores loaded modules to prevent re-evaluation.
type ModuleCache struct {
	modules map[string]*Module
	mu      sync.RWMutex
}

// Global module cache instance.
//
//nolint:gochecknoglobals // Module cache must be global to work across imports
var moduleCache = &ModuleCache{
	modules: make(map[string]*Module),
}

// GetLoadedFileModules returns all loaded file module paths.
// Only returns file modules (those with absolute paths), not stdlib modules.
func GetLoadedFileModules() []string {
	moduleCache.mu.RLock()
	defer moduleCache.mu.RUnlock()

	var paths []string
	for path := range moduleCache.modules {
		// File modules have absolute paths, stdlib modules have simple names
		if filepath.IsAbs(path) {
			paths = append(paths, path)
		}
	}
	return paths
}

// ClearModuleCache clears all loaded modules from the cache.
// This should be called before reloading in watch mode.
func ClearModuleCache() {
	moduleCache.mu.Lock()
	defer moduleCache.mu.Unlock()
	moduleCache.modules = make(map[string]*Module)
}

// resolveModule resolves and loads a module by path.
// Returns the loaded module or an error object.
// currentFile is the absolute path of the file doing the import (for relative resolution).
func resolveModule(path string, currentFile string) Object {
	// Check cache first
	moduleCache.mu.RLock()
	mod, exists := moduleCache.modules[path]
	moduleCache.mu.RUnlock()
	if exists {
		return mod
	}

	// Determine if this is a stdlib or file module
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		// File module - resolve relative to current file
		return loadFileModule(path, currentFile)
	}

	// Stdlib module
	return loadStdlibModule(path)
}

// loadStdlibModule loads a standard library module by name.
func loadStdlibModule(name string) Object {
	// Check cache first
	moduleCache.mu.RLock()
	mod, exists := moduleCache.modules[name]
	moduleCache.mu.RUnlock()
	if exists {
		return mod
	}

	var module *Module
	switch name {
	case "math":
		module = createMathModule()
	case typeNameJSON:
		module = createJSONModule()
	case "fs":
		module = createFSModule()
	case "time":
		module = createTimeModule()
	case "os":
		module = createOSModule()
	case "http":
		module = createHTTPModule()
	case "db":
		module = createDBModule()
	default:
		return newError("unknown stdlib module: %s", name)
	}

	// Cache the module
	moduleCache.mu.Lock()
	moduleCache.modules[name] = module
	moduleCache.mu.Unlock()

	return module
}

// createMathModule creates the math standard library module.
// Provides mathematical constants and functions.
func createMathModule() *Module {
	env := NewEnvironment()

	// Constants
	env.Set("PI", &Float{Value: math.Pi})
	env.Set("E", &Float{Value: math.E})

	// abs(x) - absolute value
	env.Set("abs", &Builtin{
		Name: "abs",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("abs() takes exactly 1 argument, got %d", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				if arg.Value < 0 {
					return &Integer{Value: -arg.Value}
				}
				return arg
			case *Float:
				return &Float{Value: math.Abs(arg.Value)}
			default:
				return newError("abs() requires a number, got %s", args[0].Type())
			}
		},
	})

	// min(x, y, ...) - minimum value
	env.Set("min", &Builtin{
		Name: "min",
		Fn: func(args ...Object) Object {
			if len(args) == 0 {
				return newError("min() requires at least 1 argument")
			}

			var minVal float64
			for i, arg := range args {
				var val float64
				switch a := arg.(type) {
				case *Integer:
					val = float64(a.Value)
				case *Float:
					val = a.Value
				default:
					return newError("min() requires numbers, got %s at position %d", arg.Type(), i)
				}

				if i == 0 || val < minVal {
					minVal = val
				}
			}

			// Return as integer if all inputs were integers
			allIntegers := true
			for _, arg := range args {
				if _, ok := arg.(*Float); ok {
					allIntegers = false
					break
				}
			}

			if allIntegers {
				return &Integer{Value: int64(minVal)}
			}
			return &Float{Value: minVal}
		},
	})

	// max(x, y, ...) - maximum value
	env.Set("max", &Builtin{
		Name: "max",
		Fn: func(args ...Object) Object {
			if len(args) == 0 {
				return newError("max() requires at least 1 argument")
			}

			var maxVal float64
			for i, arg := range args {
				var val float64
				switch a := arg.(type) {
				case *Integer:
					val = float64(a.Value)
				case *Float:
					val = a.Value
				default:
					return newError("max() requires numbers, got %s at position %d", arg.Type(), i)
				}

				if i == 0 || val > maxVal {
					maxVal = val
				}
			}

			// Return as integer if all inputs were integers
			allIntegers := true
			for _, arg := range args {
				if _, ok := arg.(*Float); ok {
					allIntegers = false
					break
				}
			}

			if allIntegers {
				return &Integer{Value: int64(maxVal)}
			}
			return &Float{Value: maxVal}
		},
	})

	// floor(x) - floor function
	env.Set("floor", &Builtin{
		Name: "floor",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("floor() takes exactly 1 argument, got %d", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				return arg
			case *Float:
				return &Integer{Value: int64(math.Floor(arg.Value))}
			default:
				return newError("floor() requires a number, got %s", args[0].Type())
			}
		},
	})

	// ceil(x) - ceiling function
	env.Set("ceil", &Builtin{
		Name: "ceil",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("ceil() takes exactly 1 argument, got %d", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				return arg
			case *Float:
				return &Integer{Value: int64(math.Ceil(arg.Value))}
			default:
				return newError("ceil() requires a number, got %s", args[0].Type())
			}
		},
	})

	// round(x) - round to nearest integer
	env.Set("round", &Builtin{
		Name: "round",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("round() takes exactly 1 argument, got %d", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				return arg
			case *Float:
				return &Integer{Value: int64(math.Round(arg.Value))}
			default:
				return newError("round() requires a number, got %s", args[0].Type())
			}
		},
	})

	// sqrt(x) - square root
	env.Set("sqrt", &Builtin{
		Name: "sqrt",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("sqrt() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("sqrt() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Sqrt(val)}
		},
	})

	// pow(x, y) - power function
	env.Set("pow", &Builtin{
		Name: "pow",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("pow() takes exactly 2 arguments, got %d", len(args))
			}

			var base, exp float64
			switch arg := args[0].(type) {
			case *Integer:
				base = float64(arg.Value)
			case *Float:
				base = arg.Value
			default:
				return newError("pow() requires numbers, got %s for base", args[0].Type())
			}

			switch arg := args[1].(type) {
			case *Integer:
				exp = float64(arg.Value)
			case *Float:
				exp = arg.Value
			default:
				return newError("pow() requires numbers, got %s for exponent", args[1].Type())
			}

			return &Float{Value: math.Pow(base, exp)}
		},
	})

	// sin(x) - sine function
	env.Set("sin", &Builtin{
		Name: "sin",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("sin() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("sin() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Sin(val)}
		},
	})

	// cos(x) - cosine function
	env.Set("cos", &Builtin{
		Name: "cos",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("cos() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("cos() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Cos(val)}
		},
	})

	// tan(x) - tangent function
	env.Set("tan", &Builtin{
		Name: "tan",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("tan() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("tan() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Tan(val)}
		},
	})

	// log(x) - natural logarithm
	env.Set("log", &Builtin{
		Name: "log",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("log() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("log() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Log(val)}
		},
	})

	// log10(x) - base-10 logarithm
	env.Set("log10", &Builtin{
		Name: "log10",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("log10() takes exactly 1 argument, got %d", len(args))
			}

			var val float64
			switch arg := args[0].(type) {
			case *Integer:
				val = float64(arg.Value)
			case *Float:
				val = arg.Value
			default:
				return newError("log10() requires a number, got %s", args[0].Type())
			}

			return &Float{Value: math.Log10(val)}
		},
	})

	return &Module{
		Name:  "math",
		Scope: env,
	}
}

// loadFileModule loads a module from a file path.
// path is the import path (e.g., "./utils", "./lib/helpers")
// currentFile is the absolute path of the file doing the import
func loadFileModule(path string, currentFile string) Object {
	// Resolve the absolute path
	absPath, err := resolveFilePath(path, currentFile)
	if err != nil {
		return newError("module resolution error: %s", err.Error())
	}

	// Check cache with absolute path
	moduleCache.mu.RLock()
	mod, exists := moduleCache.modules[absPath]
	moduleCache.mu.RUnlock()
	if exists {
		return mod
	}

	// Read the file
	content, err := os.ReadFile(absPath) //nolint:gosec // Intentional file inclusion for module loading
	if err != nil {
		return newError("failed to read module file '%s': %s", absPath, err.Error())
	}

	// Lex and parse
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		// Collect all parse errors
		var errMsgs []string
		for _, e := range p.Errors() {
			errMsgs = append(errMsgs, e.Message)
		}
		return newError("parse errors in module '%s': %s", absPath, strings.Join(errMsgs, "; "))
	}

	// Create a new environment for the module
	moduleEnv := NewEnvironment()
	// Set the current file path so nested imports work
	moduleEnv.currentFile = absPath

	// Evaluate the module in its own environment
	result := Eval(program, moduleEnv)
	if IsError(result) {
		return newError("runtime error in module '%s': %s", absPath, result.Inspect())
	}

	// Extract module name from path
	moduleName := extractModuleName(path)

	// Create module object
	module := &Module{
		Name:  moduleName,
		Scope: moduleEnv,
	}

	// Cache the module
	moduleCache.mu.Lock()
	moduleCache.modules[absPath] = module
	// Also cache with the original path for consistency
	moduleCache.modules[path] = module
	moduleCache.mu.Unlock()

	return module
}

// resolveFilePath resolves a relative import path to an absolute file path.
// path is the import path (e.g., "./utils", "./lib/helpers")
// currentFile is the absolute path of the file doing the import
func resolveFilePath(path string, currentFile string) (string, error) {
	// Remove leading "./" or "../"
	cleanPath := path

	// Get the directory of the current file
	var baseDir string
	if currentFile != "" {
		baseDir = filepath.Dir(currentFile)
	} else {
		// If no current file (e.g., running from REPL), use current working directory
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Join with base directory
	absPath := filepath.Join(baseDir, cleanPath)

	// Add .tsl extension if not present
	if !strings.HasSuffix(absPath, ".tsl") {
		absPath += ".tsl"
	}

	// Clean the path (resolve .., etc.)
	absPath = filepath.Clean(absPath)

	// Verify file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		//nolint:err113 // Dynamic error message needed for debugging
		return "", fmt.Errorf("module file not found: %s", absPath)
	}

	return absPath, nil
}

// extractModuleName extracts the module name from an import path.
// Examples:
//
//	"math" -> "math"
//	"./utils" -> "utils"
//	"./lib/helpers" -> "helpers"
//	"./lib/geometry.tsl" -> "geometry"
func extractModuleName(path string) string {
	// Remove any .tsl extension if present
	path = strings.TrimSuffix(path, ".tsl")

	// Get the base name (last component of path)
	name := filepath.Base(path)

	return name
}

// createJSONModule creates the json standard library module.
// Provides JSON parsing and serialization.
func createJSONModule() *Module {
	env := NewEnvironment()

	// parse(jsonString) - parse JSON string to TotalScript objects
	env.Set("parse", &Builtin{
		Name: "parse",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("parse() takes exactly 1 argument, got %d", len(args))
			}

			str, ok := args[0].(*String)
			if !ok {
				return newError("parse() requires a string, got %s", args[0].Type())
			}

			var data interface{}
			if err := json.Unmarshal([]byte(str.Value), &data); err != nil {
				return newError("invalid JSON: %s", err.Error())
			}

			return convertJSONToObject(data)
		},
	})

	// stringify(value) - convert TotalScript object to JSON string
	env.Set("stringify", &Builtin{
		Name: "stringify",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("stringify() takes exactly 1 argument, got %d", len(args))
			}

			goValue := convertObjectToGo(args[0])
			jsonBytes, err := json.Marshal(goValue)
			if err != nil {
				return newError("cannot convert to JSON: %s", err.Error())
			}

			return &String{Value: string(jsonBytes)}
		},
	})

	return &Module{
		Name:  "json",
		Scope: env,
	}
}

// convertJSONToObject converts a Go interface{} from json.Unmarshal to a TotalScript Object.
func convertJSONToObject(data interface{}) Object {
	switch v := data.(type) {
	case nil:
		return NULL
	case bool:
		if v {
			return TRUE
		}
		return FALSE
	case float64:
		// JSON numbers are always float64
		// Check if it's an integer value
		if v == float64(int64(v)) {
			return &Integer{Value: int64(v)}
		}
		return &Float{Value: v}
	case string:
		return &String{Value: v}
	case []interface{}:
		elements := make([]Object, len(v))
		for i, elem := range v {
			elements[i] = convertJSONToObject(elem)
		}
		return &Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[string]Object)
		for key, val := range v {
			pairs[key] = convertJSONToObject(val)
		}
		return &Map{Pairs: pairs}
	default:
		return newError("unsupported JSON type: %T", v)
	}
}

// convertObjectToGo converts a TotalScript Object to a Go value for JSON marshaling.
func convertObjectToGo(obj Object) interface{} {
	switch o := obj.(type) {
	case *Null:
		return nil
	case *Boolean:
		return o.Value
	case *Integer:
		return o.Value
	case *Float:
		return o.Value
	case *String:
		return o.Value
	case *Array:
		arr := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			arr[i] = convertObjectToGo(elem)
		}
		return arr
	case *Map:
		m := make(map[string]interface{})
		for key, val := range o.Pairs {
			m[key] = convertObjectToGo(val)
		}
		return m
	case *ModelInstance:
		// Convert model instance to map
		m := make(map[string]interface{})
		for key, val := range o.Fields {
			m[key] = convertObjectToGo(val)
		}
		return m
	default:
		// For unsupported types, return string representation
		return obj.Inspect()
	}
}

// createFSModule creates the fs (file system) standard library module.
func createFSModule() *Module {
	env := NewEnvironment()

	// readFile(path) - read file contents as string
	env.Set("readFile", &Builtin{
		Name: "readFile",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("readFile() takes exactly 1 argument, got %d", len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("readFile() requires a string path, got %s", args[0].Type())
			}

			content, err := os.ReadFile(path.Value) //nolint:gosec // Intentional file read
			if err != nil {
				return newError("failed to read file: %s", err.Error())
			}

			return &String{Value: string(content)}
		},
	})

	// writeFile(path, content) - write string to file
	env.Set("writeFile", &Builtin{
		Name: "writeFile",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("writeFile() takes exactly 2 arguments, got %d", len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("writeFile() requires a string path, got %s", args[0].Type())
			}

			content, ok := args[1].(*String)
			if !ok {
				return newError("writeFile() requires string content, got %s", args[1].Type())
			}

			err := os.WriteFile(path.Value, []byte(content.Value), 0644) //nolint:gosec // Intentional file write
			if err != nil {
				return newError("failed to write file: %s", err.Error())
			}

			return NULL
		},
	})

	// exists(path) - check if file or directory exists
	env.Set("exists", &Builtin{
		Name: "exists",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("exists() takes exactly 1 argument, got %d", len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("exists() requires a string path, got %s", args[0].Type())
			}

			_, err := os.Stat(path.Value)
			if err == nil {
				return TRUE
			}
			if os.IsNotExist(err) {
				return FALSE
			}
			// Other error (permission denied, etc.) - return false
			return FALSE
		},
	})

	// listDir(path) - list directory contents
	env.Set("listDir", &Builtin{
		Name: "listDir",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("listDir() takes exactly 1 argument, got %d", len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("listDir() requires a string path, got %s", args[0].Type())
			}

			entries, err := os.ReadDir(path.Value)
			if err != nil {
				return newError("failed to read directory: %s", err.Error())
			}

			elements := make([]Object, len(entries))
			for i, entry := range entries {
				elements[i] = &String{Value: entry.Name()}
			}

			return &Array{Elements: elements}
		},
	})

	return &Module{
		Name:  "fs",
		Scope: env,
	}
}

// createTimeModule creates the time standard library module.
func createTimeModule() *Module {
	env := NewEnvironment()

	// now() - get current Unix timestamp in milliseconds
	env.Set("now", &Builtin{
		Name: "now",
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("now() takes no arguments, got %d", len(args))
			}

			// Return Unix timestamp in milliseconds
			return &Integer{Value: time.Now().UnixMilli()}
		},
	})

	// sleep(milliseconds) - sleep for specified milliseconds
	env.Set("sleep", &Builtin{
		Name: "sleep",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("sleep() takes exactly 1 argument, got %d", len(args))
			}

			var ms int64
			switch arg := args[0].(type) {
			case *Integer:
				ms = arg.Value
			case *Float:
				ms = int64(arg.Value)
			default:
				return newError("sleep() requires a number, got %s", args[0].Type())
			}

			if ms < 0 {
				return newError("sleep() requires a non-negative duration")
			}

			time.Sleep(time.Duration(ms) * time.Millisecond)
			return NULL
		},
	})

	return &Module{
		Name:  "time",
		Scope: env,
	}
}

// createOSModule creates the os (operating system) standard library module.
func createOSModule() *Module {
	env := NewEnvironment()

	// env(name) - get environment variable
	env.Set("env", &Builtin{
		Name: "env",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("env() takes exactly 1 argument, got %d", len(args))
			}

			name, ok := args[0].(*String)
			if !ok {
				return newError("env() requires a string variable name, got %s", args[0].Type())
			}

			value, exists := os.LookupEnv(name.Value)
			if !exists {
				return NULL
			}

			return &String{Value: value}
		},
	})

	// args() - get command line arguments (excluding program name)
	env.Set("args", &Builtin{
		Name: "args",
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("args() takes no arguments, got %d", len(args))
			}

			// os.Args[0] is the program name, we want args[1:]
			cmdArgs := os.Args
			if len(cmdArgs) > 1 {
				cmdArgs = cmdArgs[1:]
			} else {
				cmdArgs = []string{}
			}

			elements := make([]Object, len(cmdArgs))
			for i, arg := range cmdArgs {
				elements[i] = &String{Value: arg}
			}

			return &Array{Elements: elements}
		},
	})

	return &Module{
		Name:  "os",
		Scope: env,
	}
}

// createHTTPModule creates the http standard library module.
// Provides HTTP server and client capabilities.
//
//nolint:gochecknoglobals,funlen,gocognit,gocyclo
func createHTTPModule() *Module {
	env := NewEnvironment()

	// http.Request - Request model type for type annotations
	// The actual request objects are created dynamically as Maps when handling requests,
	// but this model allows type annotations like function(req: http.Request)
	requestModel := &Model{
		Name:         "Request",
		FieldNames:   []string{},
		Fields:       make(map[string]*ast.TypeExpression),
		Annotations:  make(map[string][]string),
		Methods:      make(map[string]*Function),
		Constructors: []*Function{},
	}
	env.Set("Request", requestModel)

	// http.Response(status, body?, headers?) - Response constructor
	env.Set("Response", createResponseConstructor())

	// http.ResponseType - Response model type for type annotations
	// Used for return type annotations like function(): http.Response
	responseModel := &Model{
		Name:         "ResponseType",
		FieldNames:   []string{},
		Fields:       make(map[string]*ast.TypeExpression),
		Annotations:  make(map[string][]string),
		Methods:      make(map[string]*Function),
		Constructors: []*Function{},
	}
	env.Set("ResponseType", responseModel)

	// http.client - Client object with HTTP methods
	env.Set("client", createClientObject())

	// http.Server() - Server constructor
	env.Set("Server", createServerConstructor())

	return &Module{
		Name:  "http",
		Scope: env,
	}
}

// httpServerState holds the internal state for an HTTP server instance.
type httpServerState struct {
	routes      map[string]map[string]Object // method -> path -> handler
	middleware  []Object                     // middleware functions
	staticPaths map[string]string            // route -> filesystem path
}

// createResponseConstructor creates the Response constructor function.
// Response(status)
// Response(status, body)
// Response(status, body, headers)
func createResponseConstructor() *Builtin {
	return &Builtin{
		Name: "Response",
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 3 {
				return newError("Response() takes 1-3 arguments, got %d", len(args))
			}

			// Get status (required)
			status, ok := args[0].(*Integer)
			if !ok {
				return newError("Response() status must be integer, got %s", args[0].Type())
			}

			// Create response map
			response := &Map{Pairs: make(map[string]Object)}
			response.Pairs["status"] = status
			response.Pairs["ok"] = &Boolean{Value: status.Value >= 200 && status.Value < 300}

			// Get body (optional)
			if len(args) >= 2 {
				response.Pairs["body"] = args[1]
			} else {
				response.Pairs["body"] = &String{Value: ""}
			}

			// Get headers (optional)
			if len(args) >= 3 {
				headersMap, ok := args[2].(*Map)
				if !ok {
					return newError("Response() headers must be map, got %s", args[2].Type())
				}
				response.Pairs["headers"] = headersMap
			} else {
				response.Pairs["headers"] = &Map{Pairs: make(map[string]Object)}
			}

			// Add json() method to parse body as JSON
			response.Pairs["json"] = &Builtin{
				Name: "json",
				Fn: func(methodArgs ...Object) Object {
					if len(methodArgs) != 0 {
						return newError("json() takes no arguments")
					}

					body := response.Pairs["body"]
					bodyStr, ok := body.(*String)
					if !ok {
						return newError("cannot parse non-string body as JSON")
					}

					var data interface{}
					if err := json.Unmarshal([]byte(bodyStr.Value), &data); err != nil {
						return &Error{Message: "invalid JSON: " + err.Error()}
					}

					return convertJSONToObject(data)
				},
			}

			return response
		},
	}
}

// createClientObject creates the http.client object with HTTP methods.
//
//nolint:iface
func createClientObject() Object {
	client := &Map{Pairs: make(map[string]Object)}

	client.Pairs["get"] = createClientMethod("GET")
	client.Pairs["post"] = createClientMethod("POST")
	client.Pairs["put"] = createClientMethod("PUT")
	client.Pairs["patch"] = createClientMethod("PATCH")
	client.Pairs["delete"] = createClientMethod("DELETE")

	return client
}

// createClientMethod creates an HTTP client method for the given HTTP verb.
// GET/DELETE: (url, headers?)
// POST/PUT/PATCH: (url, body, headers?)
//
//nolint:gocognit,funlen
func createClientMethod(method string) *Builtin {
	return &Builtin{
		Name: strings.ToLower(method),
		Fn: func(args ...Object) Object {
			var url string
			var body Object
			var headers *Map

			// Parse arguments based on method
			if method == "GET" || method == "DELETE" {
				// GET/DELETE: url, headers?
				if len(args) < 1 || len(args) > 2 {
					return newError("%s() takes 1-2 arguments, got %d", method, len(args))
				}

				urlObj, ok := args[0].(*String)
				if !ok {
					return newError("%s() url must be string, got %s", method, args[0].Type())
				}
				url = urlObj.Value

				if len(args) == 2 {
					var ok bool
					headers, ok = args[1].(*Map)
					if !ok {
						return newError("%s() headers must be map, got %s", method, args[1].Type())
					}
				}
			} else {
				// POST/PUT/PATCH: url, body, headers?
				if len(args) < 2 || len(args) > 3 {
					return newError("%s() takes 2-3 arguments, got %d", method, len(args))
				}

				urlObj, ok := args[0].(*String)
				if !ok {
					return newError("%s() url must be string, got %s", method, args[0].Type())
				}
				url = urlObj.Value

				body = args[1]

				if len(args) == 3 {
					var ok bool
					headers, ok = args[2].(*Map)
					if !ok {
						return newError("%s() headers must be map, got %s", method, args[2].Type())
					}
				}
			}

			// Build HTTP request
			var reqBody io.Reader
			if body != nil {
				// Convert body to JSON if it's not a string
				if str, ok := body.(*String); ok {
					reqBody = strings.NewReader(str.Value)
				} else {
					// Convert object to JSON
					goValue := convertObjectToGo(body)
					jsonBytes, err := json.Marshal(goValue)
					if err != nil {
						return &Error{Message: "failed to marshal body: " + err.Error()}
					}
					reqBody = bytes.NewReader(jsonBytes)
				}
			}

			//nolint:noctx
			req, err := http.NewRequest(method, url, reqBody)
			if err != nil {
				return &Error{Message: "failed to create request: " + err.Error()}
			}

			// Set headers
			if headers != nil {
				for key, value := range headers.Pairs {
					if arr, ok := value.(*Array); ok {
						// Headers are arrays of strings
						for _, elem := range arr.Elements {
							if str, ok := elem.(*String); ok {
								req.Header.Add(key, str.Value)
							}
						}
					} else if str, ok := value.(*String); ok {
						req.Header.Set(key, str.Value)
					}
				}
			}

			// Execute request
			httpClient := &http.Client{Timeout: 30 * time.Second}
			resp, err := httpClient.Do(req)
			if err != nil {
				return &Error{Message: err.Error()}
			}
			//nolint:errcheck
			defer resp.Body.Close()

			// Read response body
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return &Error{Message: "failed to read response: " + err.Error()}
			}

			// Convert headers to TotalScript map
			responseHeaders := &Map{Pairs: make(map[string]Object)}
			for key, values := range resp.Header {
				elements := make([]Object, len(values))
				for i, v := range values {
					elements[i] = &String{Value: v}
				}
				responseHeaders.Pairs[key] = &Array{Elements: elements}
			}

			// Create response object
			response := &Map{Pairs: make(map[string]Object)}
			response.Pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
			response.Pairs["body"] = &String{Value: string(bodyBytes)}
			response.Pairs["headers"] = responseHeaders
			response.Pairs["ok"] = &Boolean{Value: resp.StatusCode >= 200 && resp.StatusCode < 300}

			// Add json() method
			response.Pairs["json"] = &Builtin{
				Name: "json",
				Fn: func(methodArgs ...Object) Object {
					if len(methodArgs) != 0 {
						return newError("json() takes no arguments")
					}

					var data interface{}
					if err := json.Unmarshal(bodyBytes, &data); err != nil {
						return &Error{Message: "invalid JSON: " + err.Error()}
					}

					return convertJSONToObject(data)
				},
			}

			return response
		},
	}
}

// createServerConstructor creates the Server() constructor function.
//
//nolint:funlen
func createServerConstructor() *Builtin {
	return &Builtin{
		Name: "Server",
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("Server() takes no arguments")
			}

			// Create server state
			state := &httpServerState{
				routes:      make(map[string]map[string]Object),
				middleware:  []Object{},
				staticPaths: make(map[string]string),
			}

			// Create server instance (Map object)
			serverInstance := &Map{Pairs: make(map[string]Object)}

			// Add route registration methods
			serverInstance.Pairs["get"] = createRouteMethod(state, "GET")
			serverInstance.Pairs["post"] = createRouteMethod(state, "POST")
			serverInstance.Pairs["put"] = createRouteMethod(state, "PUT")
			serverInstance.Pairs["patch"] = createRouteMethod(state, "PATCH")
			serverInstance.Pairs["delete"] = createRouteMethod(state, "DELETE")

			// Add server control methods
			serverInstance.Pairs["start"] = createStartMethod(state)
			serverInstance.Pairs["static"] = createStaticMethod(state)
			serverInstance.Pairs["use"] = createUseMethod(state)

			return serverInstance
		},
	}
}

// createRouteMethod creates a route registration method (get, post, put, patch, delete).
func createRouteMethod(state *httpServerState, method string) *Builtin {
	return &Builtin{
		Name: strings.ToLower(method),
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("%s() requires 2 arguments (path, handler), got %d", method, len(args))
			}

			path, ok := args[0].(*String)
			if !ok {
				return newError("%s() path must be string, got %s", method, args[0].Type())
			}

			handler, ok := args[1].(*Function)
			if !ok {
				return newError("%s() handler must be function, got %s", method, args[1].Type())
			}

			// Store route
			if state.routes[method] == nil {
				state.routes[method] = make(map[string]Object)
			}
			state.routes[method][path.Value] = handler

			return NULL
		},
	}
}

// createStartMethod creates the start() method for the server.
//
//nolint:funlen,gocognit,gocyclo
func createStartMethod(state *httpServerState) *Builtin {
	return &Builtin{
		Name: "start",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("start() requires 1 argument (port)")
			}

			port, ok := args[0].(*Integer)
			if !ok {
				return newError("start() port must be integer, got %s", args[0].Type())
			}

			// Create HTTP server
			mux := http.NewServeMux()

			// Handle all requests
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// Find matching route
				handler, params := matchRoute(state, r.Method, r.URL.Path)
				if handler == nil {
					w.WriteHeader(http.StatusNotFound)
					//nolint:errcheck,gosec
					w.Write([]byte("404 Not Found"))
					return
				}

				// Create request object
				requestObj := createRequestObject(r, params)

				// Execute middleware chain and handler
				result := executeMiddlewareChain(state.middleware, handler, requestObj)

				// Handle errors
				if err, ok := result.(*Error); ok {
					w.WriteHeader(http.StatusInternalServerError)
					//nolint:errcheck,gosec
					w.Write([]byte(err.Message))
					return
				}

				// Convert response to HTTP response
				writeHTTPResponse(w, result)
			})

			// Serve static files
			for routePath, fsPath := range state.staticPaths {
				fs := http.FileServer(http.Dir(fsPath))
				mux.Handle(routePath, http.StripPrefix(routePath, fs))
			}

			// Start server (blocking)
			addr := fmt.Sprintf(":%d", port.Value)
			fmt.Printf("Server listening on http://localhost%s\n", addr)

			//nolint:gosec
			if err := http.ListenAndServe(addr, mux); err != nil {
				return newError("server error: %s", err.Error())
			}

			return NULL
		},
	}
}

// createStaticMethod creates the static() method for serving static files.
func createStaticMethod(state *httpServerState) *Builtin {
	return &Builtin{
		Name: "static",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("static() requires 2 arguments (routePath, filesystemPath)")
			}

			routePath, ok := args[0].(*String)
			if !ok {
				return newError("static() routePath must be string, got %s", args[0].Type())
			}

			fsPath, ok := args[1].(*String)
			if !ok {
				return newError("static() filesystemPath must be string, got %s", args[1].Type())
			}

			state.staticPaths[routePath.Value] = fsPath.Value
			return NULL
		},
	}
}

// createUseMethod creates the use() method for middleware.
func createUseMethod(state *httpServerState) *Builtin {
	return &Builtin{
		Name: "use",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("use() requires 1 argument (middleware function)")
			}

			middleware, ok := args[0].(*Function)
			if !ok {
				return newError("use() middleware must be function, got %s", args[0].Type())
			}

			state.middleware = append(state.middleware, middleware)
			return NULL
		},
	}
}

// executeMiddlewareChain executes the middleware chain and final handler.
// Each middleware receives (req, next) where next() continues to the next middleware.
// The final next() calls the actual route handler.
//
//nolint:gocognit
func executeMiddlewareChain(middleware []Object, handler Object, req Object) Object {
	// If no middleware, call handler directly
	if len(middleware) == 0 {
		return callTSFunction(handler, req)
	}

	// Build the execution chain recursively
	// The chain executes: mw[0] -> mw[1] -> ... -> handler
	var executeChain func(index int, request Object) Object
	executeChain = func(index int, request Object) Object {
		// If we've gone through all middleware, call the handler
		if index >= len(middleware) {
			return callTSFunction(handler, request)
		}

		// Create the "next" function for this middleware
		nextFunc := &Builtin{
			Name: "next",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("next() requires 1 argument (request)")
				}
				// Continue to next middleware or handler
				return executeChain(index+1, args[0])
			},
		}

		// Call current middleware with (req, next)
		return callTSFunction(middleware[index], request, nextFunc)
	}

	// Start the chain with the first middleware
	return executeChain(0, req)
}

// matchRoute finds a matching route and extracts path parameters.
// Returns (handler, params) or (nil, nil) if no match.
func matchRoute(state *httpServerState, method, path string) (Object, map[string]string) {
	methodRoutes, ok := state.routes[method]
	if !ok {
		return nil, nil
	}

	// Try exact match first
	if handler, ok := methodRoutes[path]; ok {
		return handler, make(map[string]string)
	}

	// Try pattern matching with parameters
	for pattern, handler := range methodRoutes {
		if matched, params := matchPattern(pattern, path); matched {
			return handler, params
		}
	}

	return nil, nil
}

// matchPattern matches a path against a pattern and extracts parameters.
// Pattern: /users/:id/posts/:postId
// Path: /users/123/posts/456
// Returns: (true, {"id": "123", "postId": "456"})
func matchPattern(pattern, path string) (bool, map[string]string) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	params := make(map[string]string)

	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, ":") {
			// Parameter
			paramName := patternPart[1:]
			params[paramName] = pathParts[i]
		} else if patternPart != pathParts[i] {
			// Literal mismatch
			return false, nil
		}
	}

	return true, params
}

// createRequestObject creates a TotalScript Request object from an HTTP request.
//
//nolint:funlen,iface
func createRequestObject(r *http.Request, params map[string]string) Object {
	// Read body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body

	// Convert params to map
	paramsMap := &Map{Pairs: make(map[string]Object)}
	for k, v := range params {
		paramsMap.Pairs[k] = &String{Value: v}
	}

	// Convert query parameters to map of arrays
	queryMap := &Map{Pairs: make(map[string]Object)}
	for key, values := range r.URL.Query() {
		elements := make([]Object, len(values))
		for i, v := range values {
			elements[i] = &String{Value: v}
		}
		queryMap.Pairs[key] = &Array{Elements: elements}
	}

	// Convert headers to map of arrays
	headersMap := &Map{Pairs: make(map[string]Object)}
	for key, values := range r.Header {
		elements := make([]Object, len(values))
		for i, v := range values {
			elements[i] = &String{Value: v}
		}
		headersMap.Pairs[key] = &Array{Elements: elements}
	}

	// Create request object
	request := &Map{Pairs: make(map[string]Object)}
	request.Pairs["method"] = &String{Value: r.Method}
	request.Pairs["path"] = &String{Value: r.URL.Path}
	request.Pairs["params"] = paramsMap
	request.Pairs["query"] = queryMap
	request.Pairs["headers"] = headersMap
	request.Pairs["body"] = &String{Value: string(bodyBytes)}

	// Add json() method
	request.Pairs["json"] = &Builtin{
		Name: "json",
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("json() takes no arguments")
			}

			var data interface{}
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				return &Error{Message: "invalid JSON: " + err.Error()}
			}

			return convertJSONToObject(data)
		},
	}

	return request
}

// callTSFunction calls a TotalScript function with arguments.
func callTSFunction(fn Object, args ...Object) Object {
	function, ok := fn.(*Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	// Create new environment for function execution
	extendedEnv := NewEnclosedEnvironment(function.Env)

	// Bind parameters
	if len(args) != len(function.Parameters) {
		return newError("function expects %d arguments, got %d", len(function.Parameters), len(args))
	}

	for i, param := range function.Parameters {
		extendedEnv.Set(param.Name.Value, args[i])
	}

	// Evaluate function body
	result := Eval(function.Body, extendedEnv)

	// Unwrap return value
	if returnValue, ok := result.(*ReturnValue); ok {
		return returnValue.Value
	}

	return result
}

// writeHTTPResponse writes a TotalScript response object to an HTTP ResponseWriter.
func writeHTTPResponse(w http.ResponseWriter, response Object) {
	responseMap, ok := response.(*Map)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errcheck,gosec
		w.Write([]byte("handler must return Response object"))
		return
	}

	// Get status
	status, ok := responseMap.Pairs["status"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errcheck,gosec
		w.Write([]byte("response missing status"))
		return
	}

	statusInt, ok := status.(*Integer)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errcheck,gosec
		w.Write([]byte("response status must be integer"))
		return
	}

	// Set headers
	if headers, ok := responseMap.Pairs["headers"]; ok {
		if headersMap, ok := headers.(*Map); ok {
			for key, value := range headersMap.Pairs {
				if arr, ok := value.(*Array); ok {
					for _, elem := range arr.Elements {
						if str, ok := elem.(*String); ok {
							w.Header().Add(key, str.Value)
						}
					}
				} else if str, ok := value.(*String); ok {
					w.Header().Set(key, str.Value)
				}
			}
		}
	}

	// Write status
	w.WriteHeader(int(statusInt.Value))

	// Get body
	body, ok := responseMap.Pairs["body"]
	if !ok {
		return
	}

	// Write body
	if str, ok := body.(*String); ok {
		//nolint:errcheck,gosec
		w.Write([]byte(str.Value))
	} else {
		// Convert non-string body to JSON
		goValue := convertObjectToGo(body)
		jsonBytes, err := json.Marshal(goValue)
		if err != nil {
			//nolint:errcheck,gosec
			w.Write([]byte("error marshaling response: " + err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck,gosec
		w.Write(jsonBytes)
	}
}

// Database module for SQLite persistence

// DBState holds the database connection state.
// Exported so it can be wrapped in DBStateWrapper for query execution.
type DBState struct {
	path string
	db   *sql.DB
	tx   *sql.Tx // Active transaction, if any
	mu   sync.Mutex
}

// execer returns the appropriate executor (transaction or db connection)
func (s *DBState) execer() interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
} {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func createDBModule() *Module {
	env := NewEnvironment()

	// Database state (singleton)
	state := &DBState{
		path: "data.db",
		db:   nil, // Opened lazily
	}

	// Store state reference for query execution
	// This is used internally by evalDbFindExpression
	env.Set("__db_state__", &DBStateWrapper{State: state})

	env.Set("configure", createDBConfigureFunc(state))
	env.Set("save", createDBSaveFunc(state))
	env.Set("saveAll", createDBSaveAllFunc(state))
	env.Set("delete", createDBDeleteFunc(state))
	env.Set("deleteAll", createDBDeleteAllFunc(state))
	env.Set("transaction", createDBTransactionFunc(state))

	return &Module{Name: "db", Scope: env}
}

func (s *DBState) ensureOpen() error {
	if s.db != nil {
		return nil
	}
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	s.db = db
	return s.createSchema()
}

func (s *DBState) createSchema() error {
	_, err := s.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS data (
			entity_id TEXT NOT NULL,
			model_type TEXT NOT NULL,
			field_name TEXT NOT NULL,
			field_value TEXT,
			field_type TEXT NOT NULL,
			PRIMARY KEY (entity_id, field_name)
		);
		CREATE INDEX IF NOT EXISTS idx_model ON data(model_type);
		CREATE INDEX IF NOT EXISTS idx_field ON data(model_type, field_name, field_value);
	`)
	if err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}
	return nil
}

func createDBConfigureFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "configure",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("configure() takes 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return newError("configure() argument must be string")
			}
			state.mu.Lock()
			defer state.mu.Unlock()
			if state.db != nil {
				_ = state.db.Close()
				state.db = nil
			}
			state.path = path.Value
			return NULL
		},
	}
}

func createDBSaveFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "save",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("save() takes 1 argument")
			}
			instance, ok := args[0].(*ModelInstance)
			if !ok {
				return newError("save() argument must be model instance")
			}

			state.mu.Lock()
			defer state.mu.Unlock()

			if err := state.ensureOpen(); err != nil {
				return &Error{Message: err.Error()}
			}

			// Find existing entity by @id fields, or generate new UUID
			entityID := getOrCreateEntityID(state, instance)

			// Delete existing rows for this entity (for updates)
			_, _ = state.execer().Exec("DELETE FROM data WHERE entity_id = ?", entityID)

			// Insert all fields
			for fieldName, value := range instance.Fields {
				// Skip internal _entity_id field
				if fieldName == "_entity_id" {
					continue
				}

				fieldValue, fieldType := serializeDBValue(value)

				_, err := state.execer().Exec(`
					INSERT INTO data (entity_id, model_type, field_name, field_value, field_type)
					VALUES (?, ?, ?, ?, ?)
				`, entityID, instance.Model.Name, fieldName, fieldValue, fieldType)

				if err != nil {
					return &Error{Message: err.Error()}
				}
			}

			// Store entity_id in instance for future reference
			instance.Fields["_entity_id"] = &String{Value: entityID}

			return NULL
		},
	}
}

func createDBSaveAllFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "saveAll",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("saveAll() takes 1 argument")
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("saveAll() argument must be an array")
			}

			state.mu.Lock()
			defer state.mu.Unlock()

			if err := state.ensureOpen(); err != nil {
				return &Error{Message: err.Error()}
			}

			// Save each instance
			for _, element := range arr.Elements {
				instance, ok := element.(*ModelInstance)
				if !ok {
					return newError("saveAll() array must contain only model instances")
				}

				// Find existing entity by @id fields, or generate new UUID
				entityID := getOrCreateEntityID(state, instance)

				// Delete existing rows for this entity (for updates)
				_, _ = state.execer().Exec("DELETE FROM data WHERE entity_id = ?", entityID)

				// Insert all fields
				for fieldName, value := range instance.Fields {
					// Skip internal _entity_id field
					if fieldName == "_entity_id" {
						continue
					}

					fieldValue, fieldType := serializeDBValue(value)

					_, err := state.execer().Exec(`
						INSERT INTO data (entity_id, model_type, field_name, field_value, field_type)
						VALUES (?, ?, ?, ?, ?)
					`, entityID, instance.Model.Name, fieldName, fieldValue, fieldType)

					if err != nil {
						return &Error{Message: err.Error()}
					}
				}

				// Store entity_id in instance for future reference
				instance.Fields["_entity_id"] = &String{Value: entityID}
			}

			return NULL
		},
	}
}

func getOrCreateEntityID(state *DBState, instance *ModelInstance) string {
	// Check for @id annotated fields
	idFields := getIDFields(instance.Model)

	if len(idFields) > 0 {
		// Try to find existing entity by @id field values
		existingID := findEntityByIDFields(state, instance, idFields)
		if existingID != "" {
			return existingID // Update existing
		}
	}

	// No @id or no existing match - generate new UUID
	return uuid.New().String()
}

func getIDFields(model *Model) []string {
	idFields := []string{}
	for _, fieldName := range model.FieldNames {
		if annotations, ok := model.Annotations[fieldName]; ok {
			for _, ann := range annotations {
				if ann == "id" {
					idFields = append(idFields, fieldName)
					break
				}
			}
		}
	}
	return idFields
}

func findEntityByIDFields(state *DBState, instance *ModelInstance, idFields []string) string {
	if len(idFields) == 0 {
		return ""
	}

	// Build query to find entity where all @id fields match
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT DISTINCT entity_id FROM data WHERE model_type = ?")
	args := []interface{}{instance.Model.Name}

	for _, fieldName := range idFields {
		value := instance.Fields[fieldName]
		fieldValue, _ := serializeDBValue(value)
		queryBuilder.WriteString(" AND entity_id IN (SELECT entity_id FROM data WHERE field_name = ? AND field_value = ?)")
		args = append(args, fieldName, fieldValue)
	}
	queryBuilder.WriteString(" LIMIT 1")
	query := queryBuilder.String()

	var entityID string
	row := state.execer().QueryRow(query, args...)
	if err := row.Scan(&entityID); err != nil {
		return "" // Not found
	}
	return entityID
}

func serializeDBValue(obj Object) (string, string) {
	switch v := obj.(type) {
	case *Integer:
		return strconv.FormatInt(v.Value, 10), typeNameInteger
	case *Float:
		return strconv.FormatFloat(v.Value, 'f', -1, 64), typeNameFloat
	case *String:
		return v.Value, typeNameString
	case *Boolean:
		if v.Value {
			return stringTrue, typeNameBoolean
		}
		return stringFalse, typeNameBoolean
	case *Null:
		return "", typeNameNull
	case *ModelInstance, *Array, *Map:
		// Serialize as JSON
		jsonBytes, _ := json.Marshal(convertObjectToGo(v))
		return string(jsonBytes), typeNameJSON
	default:
		return obj.Inspect(), typeNameString
	}
}

func deserializeDBValue(value, ftype string) Object {
	switch ftype {
	case "integer":
		i, _ := strconv.ParseInt(value, 10, 64)
		return &Integer{Value: i}
	case "float":
		f, _ := strconv.ParseFloat(value, 64)
		return &Float{Value: f}
	case "string":
		return &String{Value: value}
	case "boolean":
		return nativeBoolToBooleanObject(value == "true")
	case "null":
		return NULL
	case "json":
		// Parse JSON back to TotalScript object
		var goValue interface{}
		if err := json.Unmarshal([]byte(value), &goValue); err != nil {
			return NULL
		}
		return convertGoToObject(goValue)
	default:
		return &String{Value: value}
	}
}

func createDBDeleteFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "delete",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("delete() takes 1 argument")
			}
			instance, ok := args[0].(*ModelInstance)
			if !ok {
				return newError("delete() argument must be model instance")
			}

			state.mu.Lock()
			defer state.mu.Unlock()

			if err := state.ensureOpen(); err != nil {
				return &Error{Message: err.Error()}
			}

			// Get entity_id from saved instance
			entityIDObj, ok := instance.Fields["_entity_id"]
			if !ok {
				return newError("instance has not been saved to database")
			}
			entityIDStr, ok := entityIDObj.(*String)
			if !ok {
				return newError("internal error: _entity_id is not a string")
			}
			entityID := entityIDStr.Value

			_, err := state.execer().Exec("DELETE FROM data WHERE entity_id = ?", entityID)
			if err != nil {
				return &Error{Message: err.Error()}
			}
			return NULL
		},
	}
}

func createDBDeleteAllFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "deleteAll",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("deleteAll() takes 1 argument")
			}
			model, ok := args[0].(*Model)
			if !ok {
				return newError("deleteAll() argument must be a model type")
			}

			state.mu.Lock()
			defer state.mu.Unlock()

			if err := state.ensureOpen(); err != nil {
				return &Error{Message: err.Error()}
			}

			_, err := state.execer().Exec("DELETE FROM data WHERE model_type = ?", model.Name)
			if err != nil {
				return &Error{Message: err.Error()}
			}
			return NULL
		},
	}
}

func createDBTransactionFunc(state *DBState) *Builtin {
	return &Builtin{
		Name: "transaction",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("transaction() takes 1 argument")
			}
			fn, ok := args[0].(*Function)
			if !ok {
				return newError("transaction() argument must be a function")
			}

			// Begin transaction (with lock)
			state.mu.Lock()
			if err := state.ensureOpen(); err != nil {
				state.mu.Unlock()
				return &Error{Message: err.Error()}
			}

			tx, err := state.db.BeginTx(context.Background(), nil)
			if err != nil {
				state.mu.Unlock()
				return &Error{Message: err.Error()}
			}

			// Set active transaction
			state.tx = tx
			state.mu.Unlock()

			// Cleanup function
			defer func() {
				state.mu.Lock()
				state.tx = nil
				state.mu.Unlock()
			}()

			// Execute the function (without holding the lock)
			// Individual db operations will lock/unlock as needed
			result := applyFunction(fn, []Object{}, NewEnvironment())

			// Commit or rollback (with lock)
			state.mu.Lock()
			defer state.mu.Unlock()

			// Check if function returned an error
			if IsError(result) {
				_ = tx.Rollback()
				return result
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return &Error{Message: err.Error()}
			}

			return result
		},
	}
}

func convertGoToObject(val interface{}) Object {
	switch v := val.(type) {
	case nil:
		return NULL
	case bool:
		return nativeBoolToBooleanObject(v)
	case int:
		return &Integer{Value: int64(v)}
	case int64:
		return &Integer{Value: v}
	case float64:
		return &Float{Value: v}
	case string:
		return &String{Value: v}
	case []interface{}:
		elements := make([]Object, len(v))
		for i, elem := range v {
			elements[i] = convertGoToObject(elem)
		}
		return &Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[string]Object)
		for key, val := range v {
			pairs[key] = convertGoToObject(val)
		}
		return &Map{Pairs: pairs}
	default:
		return &String{Value: fmt.Sprintf("%v", v)}
	}
}
