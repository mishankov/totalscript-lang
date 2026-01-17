package stdlib

import (
	"strings"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

// ArrayMethods returns all built-in methods for arrays.
func ArrayMethods() map[string]interpreter.BuiltinFunction {
	return map[string]interpreter.BuiltinFunction{
		"length":   arrayLength,
		"push":     arrayPush,
		"pop":      arrayPop,
		"insert":   arrayInsert,
		"remove":   arrayRemove,
		"contains": arrayContains,
		"indexOf":  arrayIndexOf,
		"join":     arrayJoin,
		"map":      arrayMap,
		"filter":   arrayFilter,
		"reduce":   arrayReduce,
		"each":     arrayEach,
	}
}

// length() returns the number of elements in the array.
func arrayLength(args ...interpreter.Object) interpreter.Object {
	if len(args) != 1 {
		return &interpreter.Error{Message: "length() takes no arguments"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "length() can only be called on arrays"}
	}

	return &interpreter.Integer{Value: int64(len(arr.Elements))}
}

// push(value) adds an element to the end of the array (mutates).
func arrayPush(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "push() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "push() can only be called on arrays"}
	}

	// Mutate the array by appending the new element
	arr.Elements = append(arr.Elements, args[1])

	return interpreter.NULL
}

// pop() removes and returns the last element of the array (mutates).
func arrayPop(args ...interpreter.Object) interpreter.Object {
	if len(args) != 1 {
		return &interpreter.Error{Message: "pop() takes no arguments"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "pop() can only be called on arrays"}
	}

	if len(arr.Elements) == 0 {
		return &interpreter.Error{Message: "pop() called on empty array"}
	}

	// Get the last element
	last := arr.Elements[len(arr.Elements)-1]

	// Mutate the array by removing the last element
	arr.Elements = arr.Elements[:len(arr.Elements)-1]

	return last
}

// insert(index, value) inserts an element at the specified index (mutates).
func arrayInsert(args ...interpreter.Object) interpreter.Object {
	if len(args) != 3 {
		return &interpreter.Error{Message: "insert() takes exactly 2 arguments"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "insert() can only be called on arrays"}
	}

	indexObj, ok := args[1].(*interpreter.Integer)
	if !ok {
		return &interpreter.Error{Message: "insert() index must be an integer"}
	}

	index := int(indexObj.Value)
	if index < 0 || index > len(arr.Elements) {
		return &interpreter.Error{Message: "insert() index out of bounds"}
	}

	// Insert the element at the specified index
	arr.Elements = append(arr.Elements[:index], append([]interpreter.Object{args[2]}, arr.Elements[index:]...)...)

	return interpreter.NULL
}

// remove(index) removes the element at the specified index (mutates).
func arrayRemove(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "remove() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "remove() can only be called on arrays"}
	}

	indexObj, ok := args[1].(*interpreter.Integer)
	if !ok {
		return &interpreter.Error{Message: "remove() index must be an integer"}
	}

	index := int(indexObj.Value)
	if index < 0 || index >= len(arr.Elements) {
		return &interpreter.Error{Message: "remove() index out of bounds"}
	}

	// Remove the element at the specified index
	arr.Elements = append(arr.Elements[:index], arr.Elements[index+1:]...)

	return interpreter.NULL
}

// contains(value) returns true if the array contains the value.
func arrayContains(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "contains() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "contains() can only be called on arrays"}
	}

	searchValue := args[1]

	// Check if the array contains the value
	for _, element := range arr.Elements {
		if objectsEqual(element, searchValue) {
			return interpreter.TRUE
		}
	}

	return interpreter.FALSE
}

// indexOf(value) returns the index of the first occurrence or Error if not found.
func arrayIndexOf(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "indexOf() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "indexOf() can only be called on arrays"}
	}

	searchValue := args[1]

	// Find the index of the value
	for i, element := range arr.Elements {
		if objectsEqual(element, searchValue) {
			return &interpreter.Integer{Value: int64(i)}
		}
	}

	return &interpreter.Error{Message: "value not found in array"}
}

// join(separator) concatenates all array elements into a string with the given separator.
func arrayJoin(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "join() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "join() can only be called on arrays"}
	}

	separator, ok := args[1].(*interpreter.String)
	if !ok {
		return &interpreter.Error{Message: "join() argument must be a string"}
	}

	// Convert all elements to strings and join
	var builder strings.Builder
	for i, element := range arr.Elements {
		if i > 0 {
			builder.WriteString(separator.Value)
		}
		// Convert element to string
		builder.WriteString(elementToString(element))
	}

	return &interpreter.String{Value: builder.String()}
}

// elementToString converts an object to its string representation
func elementToString(obj interpreter.Object) string {
	switch v := obj.(type) {
	case *interpreter.String:
		return v.Value
	case *interpreter.Integer:
		return v.Inspect()
	case *interpreter.Float:
		return v.Inspect()
	case *interpreter.Boolean:
		return v.Inspect()
	case *interpreter.Null:
		return "null"
	default:
		return obj.Inspect()
	}
}

// map(fn) returns a new array with transformed elements.
func arrayMap(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "map() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "map() can only be called on arrays"}
	}

	fn, ok := args[1].(*interpreter.Function)
	if !ok {
		return &interpreter.Error{Message: "map() argument must be a function"}
	}

	// Create a new array with transformed elements
	newElements := make([]interpreter.Object, len(arr.Elements))
	for i, element := range arr.Elements {
		// Call the function with the element
		result := applyFunction(fn, []interpreter.Object{element})
		if interpreter.IsError(result) {
			return result
		}
		newElements[i] = result
	}

	return &interpreter.Array{Elements: newElements}
}

// filter(fn) returns a new array with filtered elements.
func arrayFilter(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "filter() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "filter() can only be called on arrays"}
	}

	fn, ok := args[1].(*interpreter.Function)
	if !ok {
		return &interpreter.Error{Message: "filter() argument must be a function"}
	}

	// Create a new array with filtered elements
	newElements := make([]interpreter.Object, 0)
	for _, element := range arr.Elements {
		// Call the function with the element
		result := applyFunction(fn, []interpreter.Object{element})
		if interpreter.IsError(result) {
			return result
		}

		// Add to new array if result is truthy
		if interpreter.IsTruthy(result) {
			newElements = append(newElements, element)
		}
	}

	return &interpreter.Array{Elements: newElements}
}

// reduce(initial, fn) reduces the array to a single value.
func arrayReduce(args ...interpreter.Object) interpreter.Object {
	if len(args) != 3 {
		return &interpreter.Error{Message: "reduce() takes exactly 2 arguments"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "reduce() can only be called on arrays"}
	}

	accumulator := args[1]

	fn, ok := args[2].(*interpreter.Function)
	if !ok {
		return &interpreter.Error{Message: "reduce() second argument must be a function"}
	}

	// Reduce the array
	for _, element := range arr.Elements {
		// Call the function with accumulator and element
		result := applyFunction(fn, []interpreter.Object{accumulator, element})
		if interpreter.IsError(result) {
			return result
		}
		accumulator = result
	}

	return accumulator
}

// each(fn) iterates over the array and calls the function for each element.
func arrayEach(args ...interpreter.Object) interpreter.Object {
	if len(args) != 2 {
		return &interpreter.Error{Message: "each() takes exactly 1 argument"}
	}

	arr, ok := args[0].(*interpreter.Array)
	if !ok {
		return &interpreter.Error{Message: "each() can only be called on arrays"}
	}

	fn, ok := args[1].(*interpreter.Function)
	if !ok {
		return &interpreter.Error{Message: "each() argument must be a function"}
	}

	// Iterate over the array
	for _, element := range arr.Elements {
		// Call the function with the element
		result := applyFunction(fn, []interpreter.Object{element})
		if interpreter.IsError(result) {
			return result
		}
	}

	return interpreter.NULL
}

// Helper function to check if two objects are equal.
func objectsEqual(a, b interpreter.Object) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch a := a.(type) {
	case *interpreter.Integer:
		bInt, ok := b.(*interpreter.Integer)
		if !ok {
			return false
		}
		return a.Value == bInt.Value
	case *interpreter.Float:
		bFloat, ok := b.(*interpreter.Float)
		if !ok {
			return false
		}
		return a.Value == bFloat.Value
	case *interpreter.String:
		bStr, ok := b.(*interpreter.String)
		if !ok {
			return false
		}
		return a.Value == bStr.Value
	case *interpreter.Boolean:
		bBool, ok := b.(*interpreter.Boolean)
		if !ok {
			return false
		}
		return a.Value == bBool.Value
	case *interpreter.Null:
		return true
	default:
		return false
	}
}

// Helper function to apply a function (duplicated from interpreter package to avoid circular dependency).
func applyFunction(fn *interpreter.Function, args []interpreter.Object) interpreter.Object {
	// Create new environment for function execution
	env := interpreter.NewEnclosedEnvironment(fn.Env)

	// Bind parameters to arguments
	for i, param := range fn.Parameters {
		if i < len(args) {
			env.Set(param.Name.Value, args[i])
		}
	}

	// Execute function body
	evaluated := interpreter.Eval(fn.Body, env)

	// Unwrap return value
	if returnValue, ok := evaluated.(*interpreter.ReturnValue); ok {
		return returnValue.Value
	}

	return evaluated
}
