package stdlib

import (
	"strings"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

// StringMethods returns all string methods.
func StringMethods() map[string]interpreter.BuiltinFunction {
	return map[string]interpreter.BuiltinFunction{
		"length": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 1 {
				return &interpreter.Error{Message: "length() takes no arguments"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "length() can only be called on strings"}
			}
			return &interpreter.Integer{Value: int64(len(str.Value))}
		},
		"upper": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 1 {
				return &interpreter.Error{Message: "upper() takes no arguments"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "upper() can only be called on strings"}
			}
			return &interpreter.String{Value: strings.ToUpper(str.Value)}
		},
		"lower": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 1 {
				return &interpreter.Error{Message: "lower() takes no arguments"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "lower() can only be called on strings"}
			}
			return &interpreter.String{Value: strings.ToLower(str.Value)}
		},
		"trim": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 1 {
				return &interpreter.Error{Message: "trim() takes no arguments"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "trim() can only be called on strings"}
			}
			return &interpreter.String{Value: strings.TrimSpace(str.Value)}
		},
		"split": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 2 {
				return &interpreter.Error{Message: "split() requires 1 argument: separator"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "split() can only be called on strings"}
			}
			sep, ok := args[1].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "split() separator must be a string"}
			}
			parts := strings.Split(str.Value, sep.Value)
			elements := make([]interpreter.Object, len(parts))
			for i, part := range parts {
				elements[i] = &interpreter.String{Value: part}
			}
			return &interpreter.Array{Elements: elements}
		},
		"contains": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 2 {
				return &interpreter.Error{Message: "contains() requires 1 argument: substring"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "contains() can only be called on strings"}
			}
			substr, ok := args[1].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "contains() argument must be a string"}
			}
			if strings.Contains(str.Value, substr.Value) {
				return interpreter.TRUE
			}
			return interpreter.FALSE
		},
		"startsWith": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 2 {
				return &interpreter.Error{Message: "startsWith() requires 1 argument: prefix"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "startsWith() can only be called on strings"}
			}
			prefix, ok := args[1].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "startsWith() argument must be a string"}
			}
			if strings.HasPrefix(str.Value, prefix.Value) {
				return interpreter.TRUE
			}
			return interpreter.FALSE
		},
		"endsWith": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 2 {
				return &interpreter.Error{Message: "endsWith() requires 1 argument: suffix"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "endsWith() can only be called on strings"}
			}
			suffix, ok := args[1].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "endsWith() argument must be a string"}
			}
			if strings.HasSuffix(str.Value, suffix.Value) {
				return interpreter.TRUE
			}
			return interpreter.FALSE
		},
		"replace": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 3 {
				return &interpreter.Error{Message: "replace() requires 2 arguments: old, new"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "replace() can only be called on strings"}
			}
			old, ok := args[1].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "replace() first argument must be a string"}
			}
			newStr, ok := args[2].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "replace() second argument must be a string"}
			}
			return &interpreter.String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
		},
		"substring": func(args ...interpreter.Object) interpreter.Object {
			if len(args) != 3 {
				return &interpreter.Error{Message: "substring() requires 2 arguments: start, end"}
			}
			str, ok := args[0].(*interpreter.String)
			if !ok {
				return &interpreter.Error{Message: "substring() can only be called on strings"}
			}
			start, ok := args[1].(*interpreter.Integer)
			if !ok {
				return &interpreter.Error{Message: "substring() start must be an integer"}
			}
			end, ok := args[2].(*interpreter.Integer)
			if !ok {
				return &interpreter.Error{Message: "substring() end must be an integer"}
			}

			// Bounds checking
			strLen := int64(len(str.Value))
			if start.Value < 0 || start.Value > strLen {
				return &interpreter.Error{Message: "substring() start index out of bounds"}
			}
			if end.Value < start.Value || end.Value > strLen {
				return &interpreter.Error{Message: "substring() end index out of bounds"}
			}

			return &interpreter.String{Value: str.Value[start.Value:end.Value]}
		},
	}
}
