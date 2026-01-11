// Package interpreter implements the TotalScript interpreter.
package interpreter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mishankov/totalscript-lang/internal/ast"
)

// ObjectType represents the type of an object.
type ObjectType string

// ObjectType constants define all runtime value types.
const (
	// INTEGER_OBJ represents an integer value.
	INTEGER_OBJ          ObjectType = "INTEGER"
	FLOAT_OBJ            ObjectType = "FLOAT"
	STRING_OBJ           ObjectType = "STRING"
	BOOLEAN_OBJ          ObjectType = "BOOLEAN"
	NULL_OBJ             ObjectType = "NULL"
	RETURN_VALUE_OBJ     ObjectType = "RETURN_VALUE"
	ERROR_OBJ            ObjectType = "ERROR"
	FUNCTION_OBJ         ObjectType = "FUNCTION"
	ARRAY_OBJ            ObjectType = "ARRAY"
	MAP_OBJ              ObjectType = "MAP"
	BREAK_OBJ            ObjectType = "BREAK"
	CONTINUE_OBJ         ObjectType = "CONTINUE"
	BUILTIN_OBJ          ObjectType = "BUILTIN"
	BOUND_METHOD_OBJ     ObjectType = "BOUND_METHOD"
	MODEL_OBJ            ObjectType = "MODEL"
	MODEL_INSTANCE_OBJ   ObjectType = "MODEL_INSTANCE"
	ENUM_OBJ             ObjectType = "ENUM"
	ENUM_VALUE_OBJ       ObjectType = "ENUM_VALUE"
	MODULE_OBJ           ObjectType = "MODULE"
	DB_STATE_WRAPPER_OBJ ObjectType = "DB_STATE_WRAPPER"
)

// Object is the interface for all runtime values.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer represents an integer value.
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// Float represents a float value.
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

// String represents a string value.
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// Boolean represents a boolean value.
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// Null represents a null value.
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue wraps a value being returned.
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error represents an error.
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Function represents a function.
type Function struct {
	Parameters []*ast.Parameter
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("function")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// BuiltinFunction is the type for built-in function implementations.
type BuiltinFunction func(args ...Object) Object

// Builtin represents a built-in function.
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function: " + b.Name }

// BoundMethod represents a method bound to a receiver object.
type BoundMethod struct {
	Receiver Object
	Method   BuiltinFunction
}

func (bm *BoundMethod) Type() ObjectType { return BOUND_METHOD_OBJ }
func (bm *BoundMethod) Inspect() string  { return "bound method" }

// Array represents an array.
type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// Map represents a map.
type Map struct {
	Pairs map[string]Object
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range m.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key, value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// Break represents a break statement.
type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

// Continue represents a continue statement.
type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

// Singleton instances
var (
	NULL     = &Null{}
	TRUE     = &Boolean{Value: true}
	FALSE    = &Boolean{Value: false}
	BREAK    = &Break{}
	CONTINUE = &Continue{}
)

// IsTruthy returns whether an object is considered truthy.
func IsTruthy(obj Object) bool {
	switch obj := obj.(type) {
	case *Null:
		return false
	case *Boolean:
		return obj.Value
	case *Integer:
		return obj.Value != 0
	case *Float:
		return obj.Value != 0.0
	case *String:
		return obj.Value != ""
	default:
		// Arrays, Maps, Functions, and other types are always truthy
		return true
	}
}

// IsError returns whether an object is an error.
func IsError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

// Model represents a model definition (the type itself).
type Model struct {
	Name         string
	FieldNames   []string                       // Maintains field order
	Fields       map[string]*ast.TypeExpression // Quick field lookup
	Annotations  map[string][]string            // Field annotations (e.g., ["id"] for @id)
	Methods      map[string]*Function
	Constructors []*Function // Custom constructors
}

func (m *Model) Type() ObjectType { return MODEL_OBJ }
func (m *Model) Inspect() string  { return "model " + m.Name }

// ModelInstance represents an instance of a model.
type ModelInstance struct {
	Model  *Model
	Fields map[string]Object
}

func (mi *ModelInstance) Type() ObjectType { return MODEL_INSTANCE_OBJ }
func (mi *ModelInstance) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	// Iterate over fields in order
	for _, fieldName := range mi.Model.FieldNames {
		if v, ok := mi.Fields[fieldName]; ok {
			pairs = append(pairs, fmt.Sprintf("%s: %s", fieldName, v.Inspect()))
		}
	}
	out.WriteString(mi.Model.Name)
	out.WriteString("(")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString(")")
	return out.String()
}

// Enum represents an enum definition (the type itself).
type Enum struct {
	Name   string
	Values map[string]Object // name -> underlying value
}

func (e *Enum) Type() ObjectType { return ENUM_OBJ }
func (e *Enum) Inspect() string  { return "enum " + e.Name }

// EnumValue represents a specific enum value.
type EnumValue struct {
	EnumName string
	Name     string
	Value    Object // underlying value (integer, string, boolean)
}

func (ev *EnumValue) Type() ObjectType { return ENUM_VALUE_OBJ }
func (ev *EnumValue) Inspect() string {
	return ev.EnumName + "." + ev.Name
}

// Module represents a loaded module with its exported scope.
type Module struct {
	Name  string       // Module name (e.g., "math", "utils")
	Scope *Environment // Module's exported scope
}

func (m *Module) Type() ObjectType { return MODULE_OBJ }
func (m *Module) Inspect() string  { return "module " + m.Name }

// DBStateWrapper wraps a database state for use in query execution.
// This is an internal type used to pass database state to db.find() evaluation.
type DBStateWrapper struct {
	State *DBState
}

func (w *DBStateWrapper) Type() ObjectType { return DB_STATE_WRAPPER_OBJ }
func (w *DBStateWrapper) Inspect() string  { return "<db state>" }
