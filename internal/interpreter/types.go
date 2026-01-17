package interpreter

import (
	"strings"

	"github.com/mishankov/totalscript-lang/internal/ast"
)

// Type name constants for validation.
const (
	typeNameInteger  = "integer"
	typeNameFloat    = "float"
	typeNameString   = "string"
	typeNameBoolean  = "boolean"
	typeNameNull     = "null"
	typeNameArray    = "array"
	typeNameMap      = "map"
	typeNameFunction = "function"
	typeNameJSON     = "json"
)

// String literal constants.
const (
	stringTrue  = "true"
	stringFalse = "false"
)

// validateType checks if an object matches a type expression.
// Returns nil if valid, error object if invalid.
func validateType(obj Object, typeExpr *ast.TypeExpression, env *Environment) Object {
	if typeExpr == nil {
		// No type annotation, any value is valid
		return nil
	}

	// Handle optional types
	if typeExpr.Optional {
		if obj.Type() == NullObj {
			// null is valid for optional types
			return nil
		}
		// For non-null values, validate against the base type
		// Create a non-optional version of the type expression
		baseType := &ast.TypeExpression{
			Token:   typeExpr.Token,
			Name:    typeExpr.Name,
			Union:   typeExpr.Union,
			Generic: typeExpr.Generic,
		}
		return validateType(obj, baseType, env)
	}

	// Handle union types
	if len(typeExpr.Union) > 0 {
		// Check if object matches any of the union types
		for _, typeName := range typeExpr.Union {
			singleType := &ast.TypeExpression{
				Token:   typeExpr.Token,
				Name:    typeName,
				Generic: typeExpr.Generic,
			}
			if validateType(obj, singleType, env) == nil {
				return nil
			}
		}
		// No match found
		return newError("type mismatch: expected %s, got %s",
			typeExpr.String(), getTypeName(obj))
	}

	// Handle generic types
	if len(typeExpr.Generic) > 0 {
		return validateGenericType(obj, typeExpr, env)
	}

	// Handle simple types
	return validateSimpleType(obj, typeExpr.Name, env)
}

func validateSimpleType(obj Object, typeName string, env *Environment) Object {
	// Check built-in types
	switch typeName {
	case typeNameInteger:
		if obj.Type() != IntegerObj {
			return newError("type mismatch: expected integer, got %s", getTypeName(obj))
		}
		return nil
	case typeNameFloat:
		// Allow automatic integer-to-float conversion
		if obj.Type() == IntegerObj {
			// Integer is acceptable for float (will be converted)
			return nil
		}
		if obj.Type() != FloatObj {
			return newError("type mismatch: expected float, got %s", getTypeName(obj))
		}
		return nil
	case typeNameString:
		if obj.Type() != StringObj {
			return newError("type mismatch: expected string, got %s", getTypeName(obj))
		}
		return nil
	case typeNameBoolean:
		if obj.Type() != BooleanObj {
			return newError("type mismatch: expected boolean, got %s", getTypeName(obj))
		}
		return nil
	case typeNameNull:
		if obj.Type() != NullObj {
			return newError("type mismatch: expected null, got %s", getTypeName(obj))
		}
		return nil
	case typeNameArray:
		if obj.Type() != ArrayObj {
			return newError("type mismatch: expected array, got %s", getTypeName(obj))
		}
		return nil
	case typeNameMap:
		if obj.Type() != MapObj {
			return newError("type mismatch: expected map, got %s", getTypeName(obj))
		}
		return nil
	case typeNameFunction:
		if obj.Type() != FunctionObj {
			return newError("type mismatch: expected function, got %s", getTypeName(obj))
		}
		return nil
	}

	// Check for user-defined types (models, enums)
	typeObj, exists := env.Get(typeName)
	if !exists {
		return newError("unknown type: %s", typeName)
	}

	// Check if it's a model type
	if model, ok := typeObj.(*Model); ok {
		instance, ok := obj.(*ModelInstance)
		if !ok {
			return newError("type mismatch: expected %s, got %s", typeName, getTypeName(obj))
		}
		if instance.Model != model {
			return newError("type mismatch: expected %s, got %s",
				typeName, instance.Model.Name)
		}
		return nil
	}

	// Check if it's an enum type
	if enum, ok := typeObj.(*Enum); ok {
		enumValue, ok := obj.(*EnumValue)
		if !ok {
			return newError("type mismatch: expected %s, got %s", typeName, getTypeName(obj))
		}
		if enumValue.EnumName != enum.Name {
			return newError("type mismatch: expected %s, got %s",
				typeName, enumValue.EnumName)
		}
		return nil
	}

	return newError("'%s' is not a valid type", typeName)
}

func validateGenericType(obj Object, typeExpr *ast.TypeExpression, env *Environment) Object {
	// Currently only support array<T> and map<K, V>
	switch typeExpr.Name {
	case typeNameArray:
		if obj.Type() != ArrayObj {
			return newError("type mismatch: expected array, got %s", getTypeName(obj))
		}

		if len(typeExpr.Generic) == 0 {
			// No element type specified, any array is valid
			return nil
		}

		// Validate array element types
		arr, ok := obj.(*Array)
		if !ok {
			return newError("type assertion failed: expected *Array")
		}
		elementTypeName := typeExpr.Generic[0]

		// Check if element type is a union (contains " | ")
		var elementType *ast.TypeExpression
		if strings.Contains(elementTypeName, " | ") {
			// Parse union type
			parts := strings.Split(elementTypeName, " | ")
			elementType = &ast.TypeExpression{
				Token: typeExpr.Token,
				Union: parts,
			}
		} else {
			elementType = &ast.TypeExpression{
				Token: typeExpr.Token,
				Name:  elementTypeName,
			}
		}

		for i, elem := range arr.Elements {
			if err := validateType(elem, elementType, env); err != nil {
				errObj, ok := err.(*Error)
				if !ok {
					return newError("unexpected error type in validation")
				}
				return newError("array element %d: %s", i, errObj.Message)
			}
			// Coerce element if needed (e.g., integer to float)
			arr.Elements[i] = coerceValue(elem, elementType)
		}
		return nil

	case typeNameMap:
		if obj.Type() != MapObj {
			return newError("type mismatch: expected map, got %s", getTypeName(obj))
		}

		if len(typeExpr.Generic) < 2 {
			// No key/value types specified, any map is valid
			return nil
		}

		// Validate map key and value types
		mapObj, ok := obj.(*Map)
		if !ok {
			return newError("type assertion failed: expected *Map")
		}
		keyTypeName := typeExpr.Generic[0]
		valueTypeName := typeExpr.Generic[1]

		keyType := &ast.TypeExpression{
			Token: typeExpr.Token,
			Name:  keyTypeName,
		}
		valueType := &ast.TypeExpression{
			Token: typeExpr.Token,
			Name:  valueTypeName,
		}

		for key, value := range mapObj.Pairs {
			// Keys are always strings in current implementation
			keyObj := &String{Value: key}
			if err := validateType(keyObj, keyType, env); err != nil {
				errObj, ok := err.(*Error)
				if !ok {
					return newError("unexpected error type in validation")
				}
				return newError("map key '%s': %s", key, errObj.Message)
			}
			if err := validateType(value, valueType, env); err != nil {
				errObj, ok := err.(*Error)
				if !ok {
					return newError("unexpected error type in validation")
				}
				return newError("map value for key '%s': %s", key, errObj.Message)
			}
		}
		return nil

	default:
		return newError("generic type '%s' is not supported", typeExpr.Name)
	}
}

// coerceValue coerces a value to match a type if possible (e.g., integer to float).
// Returns the coerced value or the original value if no coercion is needed.
func coerceValue(obj Object, typeExpr *ast.TypeExpression) Object {
	if typeExpr == nil {
		return obj
	}

	// Handle optional types - coerce the base type
	if typeExpr.Optional {
		if obj.Type() == NullObj {
			return obj
		}
		baseType := &ast.TypeExpression{
			Token:   typeExpr.Token,
			Name:    typeExpr.Name,
			Union:   typeExpr.Union,
			Generic: typeExpr.Generic,
		}
		return coerceValue(obj, baseType)
	}

	// Handle union types - no coercion needed, unions accept multiple types
	if len(typeExpr.Union) > 0 {
		return obj
	}

	// Handle generic types - no coercion at container level
	if len(typeExpr.Generic) > 0 {
		return obj
	}

	// Handle integer-to-float coercion
	if typeExpr.Name == typeNameFloat && obj.Type() == IntegerObj {
		intObj, ok := obj.(*Integer)
		if ok {
			return &Float{Value: float64(intObj.Value)}
		}
	}

	return obj
}

func getTypeName(obj Object) string {
	switch obj := obj.(type) {
	case *Integer:
		return typeNameInteger
	case *Float:
		return typeNameFloat
	case *String:
		return typeNameString
	case *Boolean:
		return typeNameBoolean
	case *Null:
		return typeNameNull
	case *Array:
		return typeNameArray
	case *Map:
		return typeNameMap
	case *Function:
		return typeNameFunction
	case *ModelInstance:
		return obj.Model.Name
	case *EnumValue:
		return obj.EnumName
	case *Model:
		return "model " + obj.Name
	case *Enum:
		return "enum " + obj.Name
	default:
		return strings.ToLower(string(obj.Type()))
	}
}
