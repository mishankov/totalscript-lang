// Package stdlib provides built-in functions for TotalScript.
package stdlib

import (
	"fmt"
	"strconv"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

// Println prints arguments to stdout with a newline.
var Println = &interpreter.Builtin{
	Name: "println",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		for i, arg := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(arg.Inspect())
		}
		fmt.Println()
		return interpreter.NULL
	},
}

// Typeof returns the type name of the given value.
var Typeof = &interpreter.Builtin{
	Name: "typeof",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "typeof expects exactly 1 argument"}
		}
		return &interpreter.String{Value: string(args[0].Type())}
	},
}

// Integer converts a value to an integer.
var Integer = &interpreter.Builtin{
	Name: "integer",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "integer expects exactly 1 argument"}
		}

		switch arg := args[0].(type) {
		case *interpreter.Integer:
			return arg
		case *interpreter.Float:
			return &interpreter.Integer{Value: int64(arg.Value)}
		case *interpreter.String:
			val, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return &interpreter.Error{Message: fmt.Sprintf("cannot convert '%s' to integer", arg.Value)}
			}
			return &interpreter.Integer{Value: val}
		case *interpreter.Boolean:
			if arg.Value {
				return &interpreter.Integer{Value: 1}
			}
			return &interpreter.Integer{Value: 0}
		default:
			return &interpreter.Error{Message: fmt.Sprintf("cannot convert %s to integer", arg.Type())}
		}
	},
}

// Float converts a value to a float.
var Float = &interpreter.Builtin{
	Name: "float",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "float expects exactly 1 argument"}
		}

		switch arg := args[0].(type) {
		case *interpreter.Float:
			return arg
		case *interpreter.Integer:
			return &interpreter.Float{Value: float64(arg.Value)}
		case *interpreter.String:
			val, err := strconv.ParseFloat(arg.Value, 64)
			if err != nil {
				return &interpreter.Error{Message: fmt.Sprintf("cannot convert '%s' to float", arg.Value)}
			}
			return &interpreter.Float{Value: val}
		case *interpreter.Boolean:
			if arg.Value {
				return &interpreter.Float{Value: 1.0}
			}
			return &interpreter.Float{Value: 0.0}
		default:
			return &interpreter.Error{Message: fmt.Sprintf("cannot convert %s to float", arg.Type())}
		}
	},
}

// String converts a value to a string.
var String = &interpreter.Builtin{
	Name: "string",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "string expects exactly 1 argument"}
		}
		return &interpreter.String{Value: args[0].Inspect()}
	},
}

// Boolean converts a value to a boolean based on truthiness.
var Boolean = &interpreter.Builtin{
	Name: "boolean",
	Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "boolean expects exactly 1 argument"}
		}
		if interpreter.IsTruthy(args[0]) {
			return interpreter.TRUE
		}
		return interpreter.FALSE
	},
}

// Builtins returns all built-in functions as a map.
func Builtins() map[string]*interpreter.Builtin {
	return map[string]*interpreter.Builtin{
		"println": Println,
		"typeof":  Typeof,
		"integer": Integer,
		"float":   Float,
		"string":  String,
		"boolean": Boolean,
	}
}

// RegisterBuiltins registers all built-in functions in the given environment.
func RegisterBuiltins(env *interpreter.Environment) {
	for name, builtin := range Builtins() {
		env.Set(name, builtin)
	}
}

// RegisterMethods registers all built-in methods for object types.
func RegisterMethods() {
	// Register string methods
	for name, method := range StringMethods() {
		interpreter.RegisterMethod(interpreter.STRING_OBJ, name, method)
	}
}
