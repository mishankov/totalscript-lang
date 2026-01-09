package interpreter

import "github.com/mishankov/totalscript-lang/internal/ast"

// Environment represents a scope for variables.
type Environment struct {
	store       map[string]Object
	types       map[string]*ast.TypeExpression
	outer       *Environment
	currentFile string // Absolute path of the current file (for module imports)
}

// NewEnvironment creates a new environment.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	t := make(map[string]*ast.TypeExpression)
	return &Environment{store: s, types: t, outer: nil}
}

// NewEnclosedEnvironment creates a new environment with an outer environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a value from the environment.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set sets a value in the environment.
// If the variable exists in an outer scope, it updates that variable.
// Otherwise, it creates a new variable in the current scope.
func (e *Environment) Set(name string, val Object) Object {
	// Check if variable exists in current scope
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val
	}

	// Check if variable exists in outer scopes
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			return e.outer.Set(name, val)
		}
	}

	// Variable doesn't exist anywhere, create it in current scope
	e.store[name] = val
	return val
}

// GetType retrieves the type annotation for a variable.
func (e *Environment) GetType(name string) (*ast.TypeExpression, bool) {
	typeExpr, ok := e.types[name]
	if !ok && e.outer != nil {
		typeExpr, ok = e.outer.GetType(name)
	}
	return typeExpr, ok
}

// SetType sets the type annotation for a variable in the current scope.
func (e *Environment) SetType(name string, typeExpr *ast.TypeExpression) {
	e.types[name] = typeExpr
}

// SetCurrentFile sets the absolute path of the current file being executed.
// This is used for resolving relative module imports.
func (e *Environment) SetCurrentFile(path string) {
	e.currentFile = path
}
