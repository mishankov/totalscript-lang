package interpreter

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mishankov/totalscript-lang/internal/lexer"
	"github.com/mishankov/totalscript-lang/internal/parser"
)

// ModuleCache stores loaded modules to prevent re-evaluation.
type ModuleCache struct {
	modules map[string]*Module
}

// Global module cache instance.
//
//nolint:gochecknoglobals // Module cache must be global to work across imports
var moduleCache = &ModuleCache{
	modules: make(map[string]*Module),
}

// resolveModule resolves and loads a module by path.
// Returns the loaded module or an error object.
// currentFile is the absolute path of the file doing the import (for relative resolution).
func resolveModule(path string, currentFile string) Object {
	// Check cache first
	if mod, exists := moduleCache.modules[path]; exists {
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
	if mod, exists := moduleCache.modules[name]; exists {
		return mod
	}

	var module *Module
	switch name {
	case "math":
		module = createMathModule()
	case "json":
		module = createJSONModule()
	case "fs":
		module = createFSModule()
	case "time":
		module = createTimeModule()
	case "os":
		module = createOSModule()
	default:
		return newError("unknown stdlib module: %s", name)
	}

	// Cache the module
	moduleCache.modules[name] = module

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
	if mod, exists := moduleCache.modules[absPath]; exists {
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
	moduleCache.modules[absPath] = module
	// Also cache with the original path for consistency
	moduleCache.modules[path] = module

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
