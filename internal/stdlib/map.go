package stdlib

import (
	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

// MapMethods returns all built-in methods for maps.
func MapMethods() map[string]interpreter.BuiltinFunction {
	return map[string]interpreter.BuiltinFunction{
		"length":   mapLength,
		"keys":     mapKeys,
		"values":   mapValues,
		"contains": mapContains,
		"remove":   mapRemove,
	}
}

// length() returns the number of keys in the map.
func mapLength(args ...interpreter.Object) interpreter.Object {
	if len(args) != 1 {
		return &interpreter.Error{Message: "length() takes no arguments"}
	}

	m, ok := args[0].(*interpreter.Map)
	if !ok {
		return &interpreter.Error{Message: "length() can only be called on maps"}
	}

	return &interpreter.Integer{Value: int64(len(m.Pairs))}
}

// keys() returns an array of all keys in the map.
func mapKeys(args ...interpreter.Object) interpreter.Object {
	if len(args) != 1 {
		return &interpreter.Error{Message: "keys() takes no arguments"}
	}

	m, ok := args[0].(*interpreter.Map)
	if !ok {
		return &interpreter.Error{Message: "keys() can only be called on maps"}
	}

	// Collect all keys
	keys := make([]interpreter.Object, 0, len(m.Pairs))
	for keyStr := range m.Pairs {
		keys = append(keys, &interpreter.String{Value: keyStr})
	}

	return &interpreter.Array{Elements: keys}
}

// values() returns an array of all values in the map.
func mapValues(args ...interpreter.Object) interpreter.Object {
	if len(args) != 1 {
		return &interpreter.Error{Message: "values() takes no arguments"}
	}

	m, ok := args[0].(*interpreter.Map)
	if !ok {
		return &interpreter.Error{Message: "values() can only be called on maps"}
	}

	// Collect all values
	values := make([]interpreter.Object, 0, len(m.Pairs))
	for _, value := range m.Pairs {
		values = append(values, value)
	}

	return &interpreter.Array{Elements: values}
}

// contains(key) returns true if the map contains the key.
func mapContains(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "contains() takes exactly 1 argument"}
	}

	m, ok := args[0].(*interpreter.Map)
	if !ok {
		return &interpreter.Error{Message: "contains() can only be called on maps"}
	}

	key := args[1]

	// Map keys must be strings
	keyStr, ok := key.(*interpreter.String)
	if !ok {
		return &interpreter.Error{Message: "map key must be string, got " + string(key.Type())}
	}

	// Check if the map contains the key
	_, exists := m.Pairs[keyStr.Value]

	if exists {
		return interpreter.TRUE
	}

	return interpreter.FALSE
}

// remove(key) removes the key from the map (mutates).
func mapRemove(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "remove() takes exactly 1 argument"}
	}

	m, ok := args[0].(*interpreter.Map)
	if !ok {
		return &interpreter.Error{Message: "remove() can only be called on maps"}
	}

	key := args[1]

	// Map keys must be strings
	keyStr, ok := key.(*interpreter.String)
	if !ok {
		return &interpreter.Error{Message: "map key must be string, got " + string(key.Type())}
	}

	// Remove the key from the map
	delete(m.Pairs, keyStr.Value)

	return interpreter.NULL
}
