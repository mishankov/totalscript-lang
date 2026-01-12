package interpreter

import (
	"fmt"
	"math"

	"github.com/mishankov/totalscript-lang/internal/ast"
)

// Eval evaluates an AST node and returns the result.
func Eval(node ast.Node, env *Environment) Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.VarStatement:
		var val Object = NULL
		if node.Value != nil {
			val = Eval(node.Value, env)
			if IsError(val) {
				return val
			}
		}
		// Validate type if type annotation is present
		if node.Type != nil {
			if err := validateType(val, node.Type, env); err != nil {
				return err
			}
			// Coerce value if needed (e.g., integer to float)
			val = coerceValue(val, node.Type)
			// Store type annotation for future reassignments
			env.SetType(node.Name.Value, node.Type)
		}
		// If assigning a model or enum, set its name
		if model, ok := val.(*Model); ok {
			model.Name = node.Name.Value
		} else if enum, ok := val.(*Enum); ok {
			enum.Name = node.Name.Value
		}
		env.Set(node.Name.Value, val)
		return val

	case *ast.ConstStatement:
		val := Eval(node.Value, env)
		if IsError(val) {
			return val
		}
		// Validate type if type annotation is present
		if node.Type != nil {
			if err := validateType(val, node.Type, env); err != nil {
				return err
			}
			// Coerce value if needed (e.g., integer to float)
			val = coerceValue(val, node.Type)
			// Store type annotation for constants too
			env.SetType(node.Name.Value, node.Type)
		}
		// If assigning a model or enum, set its name
		if model, ok := val.(*Model); ok {
			model.Name = node.Name.Value
		} else if enum, ok := val.(*Enum); ok {
			enum.Name = node.Name.Value
		}
		env.Set(node.Name.Value, val)
		return val

	case *ast.ImportStatement:
		return evalImportStatement(node, env)

	case *ast.ReturnStatement:
		var val Object = NULL
		if node.ReturnValue != nil {
			val = Eval(node.ReturnValue, env)
			if IsError(val) {
				return val
			}
		}
		return &ReturnValue{Value: val}

	case *ast.BreakStatement:
		return BREAK

	case *ast.ContinueStatement:
		return CONTINUE

	case *ast.WhileStatement:
		return evalWhileStatement(node, env)

	case *ast.ForStatement:
		return evalForStatement(node, env)

	case *ast.SwitchStatement:
		return evalSwitchStatement(node, env)

	// Expressions
	case *ast.IntegerLiteral:
		return &Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &Float{Value: node.Value}

	case *ast.StringLiteral:
		return &String{Value: node.Value}

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.NullLiteral:
		return NULL

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		// Handle assignment operators specially - don't evaluate left side
		if node.Operator == "=" || node.Operator == "+=" || node.Operator == "-=" ||
			node.Operator == "*=" || node.Operator == "/=" || node.Operator == "%=" {
			return evalAssignmentExpression(node, env)
		}

		left := Eval(node.Left, env)
		if IsError(left) {
			return left
		}

		// Special handling for 'is' operator with type names
		if node.Operator == "is" {
			// Check if right side is an identifier (potential type name)
			if ident, ok := node.Right.(*ast.Identifier); ok {
				return evalIsOperatorWithTypeName(left, ident.Value, env)
			}
			// Check if right side is a literal that represents a type
			if _, ok := node.Right.(*ast.NullLiteral); ok {
				return evalIsOperatorWithTypeName(left, "null", env)
			}
			if boolLit, ok := node.Right.(*ast.BooleanLiteral); ok {
				if boolLit.Value {
					return evalIsOperatorWithTypeName(left, "boolean", env)
				}
				return evalIsOperatorWithTypeName(left, "boolean", env)
			}
		}

		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if IsError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && IsError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, env)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && IsError(elements[0]) {
			return elements[0]
		}
		return &Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if IsError(left) {
			return left
		}

		// Check if this is a slice operation (index is a RangeExpression)
		if rangeExpr, ok := node.Index.(*ast.RangeExpression); ok {
			return evalSliceExpression(left, rangeExpr, env)
		}

		index := Eval(node.Index, env)
		if IsError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.MapLiteral:
		return evalMapLiteral(node, env)

	case *ast.MemberExpression:
		return evalMemberExpression(node, env)

	case *ast.RangeExpression:
		return evalRangeExpression(node, env)

	case *ast.ModelLiteral:
		return evalModelLiteral(node, env)

	case *ast.EnumLiteral:
		return evalEnumLiteral(node, env)

	case *ast.ThisExpression:
		return evalThisExpression(env)

	case *ast.DbFindExpression:
		return evalDbFindExpression(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *Environment) Object {
	var result Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *ReturnValue:
			return result.Value
		case *Error:
			return result
		}
	}

	return result
}

func evalImportStatement(is *ast.ImportStatement, env *Environment) Object {
	// Resolve and load the module
	module := resolveModule(is.Path, env.currentFile)
	if IsError(module) {
		return module
	}

	// Store module in environment under ModuleName
	env.Set(is.ModuleName, module)

	// Import statements don't produce values
	return NULL
}

func evalBlockStatement(block *ast.BlockStatement, env *Environment) Object {
	var result Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == BREAK_OBJ || rt == CONTINUE_OBJ {
				return result
			}
		}
	}

	return result
}

func evalWhileStatement(ws *ast.WhileStatement, env *Environment) Object {
	var result Object = NULL

	for {
		condition := Eval(ws.Condition, env)
		if IsError(condition) {
			return condition
		}

		if !IsTruthy(condition) {
			break
		}

		result = Eval(ws.Body, env)
		if IsError(result) {
			return result
		}

		if result.Type() == RETURN_VALUE_OBJ {
			return result
		}

		if result.Type() == BREAK_OBJ {
			break
		}

		if result.Type() == CONTINUE_OBJ {
			continue
		}
	}

	return result
}

func evalForStatement(fs *ast.ForStatement, env *Environment) Object {
	var result Object = NULL

	if fs.IsRangeStyle {
		// For-in style
		iterable := Eval(fs.Iterable, env)
		if IsError(iterable) {
			return iterable
		}

		switch iter := iterable.(type) {
		case *Array:
			for i, elem := range iter.Elements {
				forEnv := NewEnclosedEnvironment(env)
				if fs.Iterator != nil {
					forEnv.Set(fs.Iterator.Value, &Integer{Value: int64(i)})
				}
				forEnv.Set(fs.Value.Value, elem)

				result = Eval(fs.Body, forEnv)
				if IsError(result) {
					return result
				}
				if result.Type() == RETURN_VALUE_OBJ {
					return result
				}
				if result.Type() == BREAK_OBJ {
					break
				}
			}
		case *Map:
			for key, value := range iter.Pairs {
				forEnv := NewEnclosedEnvironment(env)
				if fs.Iterator != nil {
					forEnv.Set(fs.Iterator.Value, &String{Value: key})
				}
				forEnv.Set(fs.Value.Value, value)

				result = Eval(fs.Body, forEnv)
				if IsError(result) {
					return result
				}
				if result.Type() == RETURN_VALUE_OBJ {
					return result
				}
				if result.Type() == BREAK_OBJ {
					break
				}
			}
		default:
			return newError("cannot iterate over %s", iterable.Type())
		}
	} else {
		// C-style for
		forEnv := NewEnclosedEnvironment(env)

		// Init
		if fs.Init != nil {
			initResult := Eval(fs.Init, forEnv)
			if IsError(initResult) {
				return initResult
			}
		}

		// Loop
		for {
			// Condition
			if fs.Condition != nil {
				condition := Eval(fs.Condition, forEnv)
				if IsError(condition) {
					return condition
				}
				if !IsTruthy(condition) {
					break
				}
			}

			// Body
			result = Eval(fs.Body, forEnv)
			if IsError(result) {
				return result
			}
			if result.Type() == RETURN_VALUE_OBJ {
				return result
			}
			if result.Type() == BREAK_OBJ {
				break
			}

			// Post
			if fs.Post != nil {
				postResult := Eval(fs.Post, forEnv)
				if IsError(postResult) {
					return postResult
				}
			}
		}
	}

	return result
}

func evalSwitchStatement(ss *ast.SwitchStatement, env *Environment) Object {
	value := Eval(ss.Value, env)
	if IsError(value) {
		return value
	}

	for _, caseClause := range ss.Cases {
		for _, caseValue := range caseClause.Values {
			cv := Eval(caseValue, env)
			if IsError(cv) {
				return cv
			}
			if objectsEqual(value, cv) {
				return Eval(caseClause.Body, env)
			}
		}
	}

	if ss.Default != nil {
		return Eval(ss.Default, env)
	}

	return NULL
}

func evalPrefixExpression(operator string, right Object) Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right Object) Object {
	if IsTruthy(right) {
		return FALSE
	}
	return TRUE
}

func evalMinusPrefixOperatorExpression(right Object) Object {
	switch right := right.(type) {
	case *Integer:
		return &Integer{Value: -right.Value}
	case *Float:
		return &Float{Value: -right.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right Object) Object {
	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == FLOAT_OBJ || right.Type() == FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(objectsEqual(left, right))
	case operator == "!=":
		return nativeBoolToBooleanObject(!objectsEqual(left, right))
	case operator == "&&":
		return nativeBoolToBooleanObject(IsTruthy(left) && IsTruthy(right))
	case operator == "||":
		return nativeBoolToBooleanObject(IsTruthy(left) || IsTruthy(right))
	case operator == "is":
		return evalIsOperator(left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalAssignmentExpression(node *ast.InfixExpression, env *Environment) Object {
	// Evaluate the right side first
	val := Eval(node.Right, env)
	if IsError(val) {
		return val
	}

	// Handle different left-hand side types
	switch left := node.Left.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x = value
		return evalIdentifierAssignment(left, node.Operator, val, env)

	case *ast.IndexExpression:
		// Index assignment: arr[0] = value or map["key"] = value
		return evalIndexAssignment(left, node.Operator, val, env)

	case *ast.MemberExpression:
		// Member assignment: obj.field = value
		return evalMemberAssignment(left, node.Operator, val, env)

	default:
		return newError("cannot assign to %T", node.Left)
	}
}

func evalIdentifierAssignment(ident *ast.Identifier, operator string, val Object, env *Environment) Object {
	// Handle compound assignment operators
	if operator != "=" {
		// Get current value
		currentVal, ok := env.Get(ident.Value)
		if !ok {
			return newError("identifier not found: %s", ident.Value)
		}

		// Determine the arithmetic operator
		var op string
		switch operator {
		case "+=":
			op = "+"
		case "-=":
			op = "-"
		case "*=":
			op = "*"
		case "/=":
			op = "/"
		case "%=":
			op = "%"
		}

		// Perform the operation
		val = evalInfixExpression(op, currentVal, val)
		if IsError(val) {
			return val
		}
	}

	// Validate type if variable has a type annotation
	if typeExpr, ok := env.GetType(ident.Value); ok {
		if err := validateType(val, typeExpr, env); err != nil {
			return err
		}
		// Coerce value if needed (e.g., integer to float)
		val = coerceValue(val, typeExpr)
	}

	// Set the variable
	env.Set(ident.Value, val)
	return val
}

func evalIndexAssignment(indexExpr *ast.IndexExpression, operator string, val Object, env *Environment) Object {
	// Evaluate the object being indexed
	obj := Eval(indexExpr.Left, env)
	if IsError(obj) {
		return obj
	}

	// Evaluate the index
	index := Eval(indexExpr.Index, env)
	if IsError(index) {
		return index
	}

	// Handle compound assignment operators
	if operator != "=" {
		// Get current value at index
		currentVal := evalIndexExpression(obj, index)
		if IsError(currentVal) {
			return currentVal
		}

		// Determine the arithmetic operator
		var op string
		switch operator {
		case "+=":
			op = "+"
		case "-=":
			op = "-"
		case "*=":
			op = "*"
		case "/=":
			op = "/"
		case "%=":
			op = "%"
		}

		// Perform the operation
		val = evalInfixExpression(op, currentVal, val)
		if IsError(val) {
			return val
		}
	}

	// Handle array and map index assignment
	switch objType := obj.(type) {
	case *Array:
		// Array index assignment
		idx, ok := index.(*Integer)
		if !ok {
			return newError("array index must be integer, got %s", index.Type())
		}

		arrayLen := int64(len(objType.Elements))
		indexVal := idx.Value

		// Handle negative indices
		if indexVal < 0 {
			indexVal = arrayLen + indexVal
		}

		// Validate index bounds
		if indexVal < 0 || indexVal >= arrayLen {
			return newError("array index out of bounds: %d", idx.Value)
		}

		// Modify the array element in place
		objType.Elements[indexVal] = val
		return val

	case *Map:
		// Map index assignment
		keyStr, ok := index.(*String)
		if !ok {
			return newError("map key must be string, got %s", index.Type())
		}

		// Set the key-value pair
		objType.Pairs[keyStr.Value] = val
		return val

	default:
		return newError("index assignment not supported for %s", obj.Type())
	}
}

func evalMemberAssignment(memberExpr *ast.MemberExpression, operator string, val Object, env *Environment) Object {
	// Evaluate the object
	obj := Eval(memberExpr.Object, env)
	if IsError(obj) {
		return obj
	}

	// Only ModelInstance supports field assignment
	instance, ok := obj.(*ModelInstance)
	if !ok {
		return newError("member assignment only supported for model instances, got %s", obj.Type())
	}

	memberName := memberExpr.Member.Value

	// Check if field exists in model
	if _, exists := instance.Model.Fields[memberName]; !exists {
		return newError("model %s has no field '%s'", instance.Model.Name, memberName)
	}

	// Handle compound assignment operators
	if operator != "=" {
		// Get current field value
		currentVal, exists := instance.Fields[memberName]
		if !exists {
			return newError("field '%s' not initialized", memberName)
		}

		// Determine the arithmetic operator
		var op string
		switch operator {
		case "+=":
			op = "+"
		case "-=":
			op = "-"
		case "*=":
			op = "*"
		case "/=":
			op = "/"
		case "%=":
			op = "%"
		}

		// Perform the operation
		val = evalInfixExpression(op, currentVal, val)
		if IsError(val) {
			return val
		}
	}

	// Set the field value
	instance.Fields[memberName] = val
	return val
}

func evalIntegerInfixExpression(operator string, left, right Object) Object {
	leftInt, ok := left.(*Integer)
	if !ok {
		return newError("type error: expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*Integer)
	if !ok {
		return newError("type error: expected integer, got %s", right.Type())
	}
	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch operator {
	case "+":
		return &Integer{Value: leftVal + rightVal}
	case "-":
		return &Integer{Value: leftVal - rightVal}
	case "*":
		return &Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &Float{Value: float64(leftVal) / float64(rightVal)}
	case "//":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &Integer{Value: leftVal / rightVal}
	case "%":
		return &Integer{Value: leftVal % rightVal}
	case "**":
		result := int64(1)
		for i := int64(0); i < rightVal; i++ {
			result *= leftVal
		}
		return &Integer{Value: result}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right Object) Object {
	var leftVal, rightVal float64

	switch left := left.(type) {
	case *Float:
		leftVal = left.Value
	case *Integer:
		leftVal = float64(left.Value)
	default:
		return newError("type mismatch in float operation")
	}

	switch right := right.(type) {
	case *Float:
		rightVal = right.Value
	case *Integer:
		rightVal = float64(right.Value)
	default:
		return newError("type mismatch in float operation")
	}

	switch operator {
	case "+":
		return &Float{Value: leftVal + rightVal}
	case "-":
		return &Float{Value: leftVal - rightVal}
	case "*":
		return &Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &Float{Value: leftVal / rightVal}
	case "**":
		return &Float{Value: pow(leftVal, rightVal)}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right Object) Object {
	leftStr, ok := left.(*String)
	if !ok {
		return newError("type error: expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*String)
	if !ok {
		return newError("type error: expected string, got %s", right.Type())
	}
	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch operator {
	case "+":
		return &String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) Object {
	condition := Eval(ie.Condition, env)
	if IsError(condition) {
		return condition
	}

	if IsTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalIdentifier(node *ast.Identifier, env *Environment) Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return val
}

func evalExpressions(exps []ast.Expression, env *Environment) []Object {
	var result []Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if IsError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn Object, args []Object, callingEnv *Environment) Object {
	switch fn := fn.(type) {
	case *Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments: expected %d, got %d",
				len(fn.Parameters), len(args))
		}

		// Validate parameter types if annotations are present and coerce if needed
		for i, param := range fn.Parameters {
			if param.Type != nil {
				if err := validateType(args[i], param.Type, callingEnv); err != nil {
					errObj, ok := err.(*Error)
					if !ok {
						return newError("unexpected error type in parameter validation")
					}
					return newError("parameter '%s': %s", param.Name.Value, errObj.Message)
				}
				// Coerce value if needed (e.g., integer to float)
				args[i] = coerceValue(args[i], param.Type)
			}
		}

		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *Builtin:
		return fn.Fn(args...)

	case *BoundMethod:
		// Prepend the receiver as the first argument
		methodArgs := make([]Object, 0, len(args)+1)
		methodArgs = append(methodArgs, fn.Receiver)
		methodArgs = append(methodArgs, args...)
		return fn.Method(methodArgs...)

	case *Model:
		// Try custom constructors first
		for _, constructor := range fn.Constructors {
			if len(constructor.Parameters) == len(args) {
				// Validate parameter types if annotations are present and coerce if needed
				for i, param := range constructor.Parameters {
					if param.Type != nil {
						if err := validateType(args[i], param.Type, callingEnv); err != nil {
							errObj, ok := err.(*Error)
							if !ok {
								return newError("unexpected error type in constructor parameter validation")
							}
							return newError("constructor parameter '%s': %s", param.Name.Value, errObj.Message)
						}
						// Coerce value if needed (e.g., integer to float)
						args[i] = coerceValue(args[i], param.Type)
					}
				}

				// Found a matching constructor, call it
				extendedEnv := extendFunctionEnv(constructor, args)
				evaluated := Eval(constructor.Body, extendedEnv)
				if returnValue, ok := evaluated.(*ReturnValue); ok {
					return returnValue.Value
				}
				return evaluated
			}
		}

		// No matching custom constructor, use default constructor
		instance := &ModelInstance{
			Model:  fn,
			Fields: make(map[string]Object),
		}

		// Check argument count matches field count
		if len(args) != len(fn.FieldNames) {
			return newError("wrong number of arguments for %s: expected %d, got %d",
				fn.Name, len(fn.FieldNames), len(args))
		}

		// Assign arguments to fields in order, validating types
		// Use calling environment for type validation so user-defined types are accessible
		for i, fieldName := range fn.FieldNames {
			// Validate field type if annotation is present
			if fieldType, ok := fn.Fields[fieldName]; ok && fieldType != nil {
				if err := validateType(args[i], fieldType, callingEnv); err != nil {
					errObj, ok := err.(*Error)
					if !ok {
						return newError("unexpected error type in field validation")
					}
					return newError("field '%s': %s", fieldName, errObj.Message)
				}
				// Coerce value if needed (e.g., integer to float)
				args[i] = coerceValue(args[i], fieldType)
			}
			instance.Fields[fieldName] = args[i]
		}

		return instance

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *Function, args []Object) *Environment {
	env := NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Name.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index Object) Object {
	switch {
	case left.Type() == ARRAY_OBJ && index.Type() == INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == MAP_OBJ && index.Type() == STRING_OBJ:
		return evalMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index Object) Object {
	arrayObject, ok := array.(*Array)
	if !ok {
		return newError("type error: expected array, got %s", array.Type())
	}
	indexInt, ok := index.(*Integer)
	if !ok {
		return newError("index operator not supported: %s", index.Type())
	}
	idx := indexInt.Value
	maxIdx := int64(len(arrayObject.Elements) - 1)

	if idx < 0 {
		// Negative indexing
		idx = int64(len(arrayObject.Elements)) + idx
	}

	if idx < 0 || idx > maxIdx {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalMapIndexExpression(mapObj, index Object) Object {
	mapObject, ok := mapObj.(*Map)
	if !ok {
		return newError("type error: expected map, got %s", mapObj.Type())
	}
	keyStr, ok := index.(*String)
	if !ok {
		return newError("map key must be string, got %s", index.Type())
	}
	key := keyStr.Value

	value, ok := mapObject.Pairs[key]
	if !ok {
		return NULL
	}

	return value
}

func evalMapLiteral(node *ast.MapLiteral, env *Environment) Object {
	pairs := make(map[string]Object)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if IsError(key) {
			return key
		}

		keyStr, ok := key.(*String)
		if !ok {
			return newError("map key must be string, got %s", key.Type())
		}

		value := Eval(valueNode, env)
		if IsError(value) {
			return value
		}

		pairs[keyStr.Value] = value
	}

	return &Map{Pairs: pairs}
}

func evalMemberExpression(node *ast.MemberExpression, env *Environment) Object {
	object := Eval(node.Object, env)
	if IsError(object) {
		return object
	}

	memberName := node.Member.Value

	// Handle Enum member access (Enum.Value) and methods
	if enum, ok := object.(*Enum); ok {
		// Check for special methods
		if memberName == "values" {
			// Return a builtin function that returns all enum values
			return &Builtin{
				Name: "values",
				Fn: func(args ...Object) Object {
					if len(args) != 0 {
						return newError("values() takes no arguments")
					}
					elements := make([]Object, 0, len(enum.Values))
					for name, value := range enum.Values {
						elements = append(elements, &EnumValue{
							EnumName: enum.Name,
							Name:     name,
							Value:    value,
						})
					}
					return &Array{Elements: elements}
				},
			}
		}

		if memberName == "fromValue" {
			// Return a builtin function that finds enum by value
			return &Builtin{
				Name: "fromValue",
				Fn: func(args ...Object) Object {
					if len(args) != 1 {
						return newError("fromValue() takes exactly 1 argument")
					}
					searchValue := args[0]
					// Search for matching value
					for name, value := range enum.Values {
						if objectsEqual(value, searchValue) {
							return &EnumValue{
								EnumName: enum.Name,
								Name:     name,
								Value:    value,
							}
						}
					}
					return newError("no enum value found for %s", searchValue.Inspect())
				},
			}
		}

		// Check for regular enum values
		if value, exists := enum.Values[memberName]; exists {
			return &EnumValue{
				EnumName: enum.Name,
				Name:     memberName,
				Value:    value,
			}
		}
		return newError("enum %s has no member '%s'", enum.Name, memberName)
	}

	// Handle Module member access
	if module, ok := object.(*Module); ok {
		value, exists := module.Scope.Get(memberName)
		if !exists {
			return newError("module '%s' has no member '%s'", module.Name, memberName)
		}
		return value
	}

	// Handle ModelInstance member access
	if instance, ok := object.(*ModelInstance); ok {
		// Check if it's a field
		if value, exists := instance.Fields[memberName]; exists {
			return value
		}

		// Check if it's a method
		if method, exists := instance.Model.Methods[memberName]; exists {
			// Create a new environment with 'this' bound to the instance
			methodEnv := NewEnclosedEnvironment(method.Env)
			methodEnv.Set("this", instance)

			// Return a function that will use this environment
			return &Function{
				Parameters: method.Parameters,
				Body:       method.Body,
				Env:        methodEnv,
			}
		}

		return newError("model %s has no field or method '%s'", instance.Model.Name, memberName)
	}

	// Handle EnumValue.value access
	if enumValue, ok := object.(*EnumValue); ok {
		if memberName == "value" {
			return enumValue.Value
		}
		return newError("enum value only has 'value' member")
	}

	// Handle Map member access (same as index access)
	if mapObj, ok := object.(*Map); ok {
		value, exists := mapObj.Pairs[memberName]
		if !exists {
			return NULL
		}
		return value
	}

	// Look up method for this object type (for built-in types)
	method := getMethod(object.Type(), memberName)
	if method == nil {
		return newError("undefined method '%s' for type %s", memberName, object.Type())
	}

	// Return a bound method that includes the receiver
	return &BoundMethod{
		Receiver: object,
		Method:   method,
	}
}

func evalRangeExpression(node *ast.RangeExpression, env *Environment) Object {
	start := Eval(node.Start, env)
	if IsError(start) {
		return start
	}

	end := Eval(node.End, env)
	if IsError(end) {
		return end
	}

	startInt, ok := start.(*Integer)
	if !ok {
		return newError("range start must be integer")
	}

	endInt, ok := end.(*Integer)
	if !ok {
		return newError("range end must be integer")
	}

	elements := []Object{}
	endVal := endInt.Value
	if node.Inclusive {
		endVal++
	}

	for i := startInt.Value; i < endVal; i++ {
		elements = append(elements, &Integer{Value: i})
	}

	return &Array{Elements: elements}
}

func evalSliceExpression(obj Object, rangeExpr *ast.RangeExpression, env *Environment) Object {
	// Only arrays support slicing
	arr, ok := obj.(*Array)
	if !ok {
		return newError("slice operation only supported for arrays, got %s", obj.Type())
	}

	arrayLen := int64(len(arr.Elements))

	// Evaluate start index (default to 0 if not specified)
	var startIdx int64
	if rangeExpr.Start != nil {
		start := Eval(rangeExpr.Start, env)
		if IsError(start) {
			return start
		}
		startInt, ok := start.(*Integer)
		if !ok {
			return newError("slice start must be integer, got %s", start.Type())
		}
		startIdx = startInt.Value
		// Handle negative indices
		if startIdx < 0 {
			startIdx = arrayLen + startIdx
		}
	} else {
		startIdx = 0
	}

	// Evaluate end index (default to array length if not specified)
	var endIdx int64
	if rangeExpr.End != nil {
		end := Eval(rangeExpr.End, env)
		if IsError(end) {
			return end
		}
		endInt, ok := end.(*Integer)
		if !ok {
			return newError("slice end must be integer, got %s", end.Type())
		}
		endIdx = endInt.Value
		// Handle negative indices
		if endIdx < 0 {
			endIdx = arrayLen + endIdx
		}
		// Handle inclusive range
		if rangeExpr.Inclusive {
			endIdx++
		}
	} else {
		endIdx = arrayLen
	}

	// Clamp bounds to valid range
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > arrayLen {
		startIdx = arrayLen
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > arrayLen {
		endIdx = arrayLen
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}

	// Create the sliced array
	elements := make([]Object, endIdx-startIdx)
	copy(elements, arr.Elements[startIdx:endIdx])

	return &Array{Elements: elements}
}

func objectsEqual(left, right Object) bool {
	if left.Type() != right.Type() {
		return false
	}

	switch left := left.(type) {
	case *Integer:
		rightInt, ok := right.(*Integer)
		return ok && left.Value == rightInt.Value
	case *Float:
		rightFloat, ok := right.(*Float)
		return ok && left.Value == rightFloat.Value
	case *String:
		rightStr, ok := right.(*String)
		return ok && left.Value == rightStr.Value
	case *Boolean:
		rightBool, ok := right.(*Boolean)
		return ok && left.Value == rightBool.Value
	case *Null:
		return true
	default:
		return false
	}
}

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func pow(base, exponent float64) float64 {
	return math.Pow(base, exponent)
}

// newError creates a new error object.
func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

// NewError creates a new error object (exported for stdlib modules).
func NewError(format string, a ...interface{}) *Error {
	return newError(format, a...)
}

// methodRegistry stores methods for each object type
var methodRegistry = make(map[ObjectType]map[string]BuiltinFunction)

// RegisterMethod registers a method for a given object type.
func RegisterMethod(objType ObjectType, name string, method BuiltinFunction) {
	if methodRegistry[objType] == nil {
		methodRegistry[objType] = make(map[string]BuiltinFunction)
	}
	methodRegistry[objType][name] = method
}

// getMethod retrieves a method for a given object type and method name.
func getMethod(objType ObjectType, name string) BuiltinFunction {
	if methods, ok := methodRegistry[objType]; ok {
		return methods[name]
	}
	return nil
}

func evalModelLiteral(node *ast.ModelLiteral, env *Environment) Object {
	model := &Model{
		Name:         "", // Name will be set when assigned to a variable
		FieldNames:   make([]string, 0, len(node.Fields)),
		Fields:       make(map[string]*ast.TypeExpression),
		Annotations:  make(map[string][]string),
		Methods:      make(map[string]*Function),
		Constructors: make([]*Function, 0),
	}

	// Store field type information and annotations in order
	for _, field := range node.Fields {
		fieldName := field.Name.Value
		model.FieldNames = append(model.FieldNames, fieldName)
		model.Fields[fieldName] = field.Type
		if len(field.Annotations) > 0 {
			model.Annotations[fieldName] = field.Annotations
		}
	}

	// Store constructors
	for _, constructor := range node.Constructors {
		fn := &Function{
			Parameters: constructor.Parameters,
			Body:       constructor.Body,
			Env:        env,
		}
		model.Constructors = append(model.Constructors, fn)
	}

	// Evaluate and store methods
	for _, method := range node.Methods {
		fn := &Function{
			Parameters: method.Function.Parameters,
			Body:       method.Function.Body,
			Env:        env,
		}
		model.Methods[method.Name.Value] = fn
	}

	return model
}

func evalEnumLiteral(node *ast.EnumLiteral, env *Environment) Object {
	enum := &Enum{
		Name:   "", // Name will be set when assigned to a variable
		Values: make(map[string]Object),
	}

	// Evaluate all enum values
	for _, enumValue := range node.Values {
		value := Eval(enumValue.Value, env)
		if IsError(value) {
			return value
		}
		enum.Values[enumValue.Name.Value] = value
	}

	return enum
}

func evalThisExpression(env *Environment) Object {
	// Look for 'this' in the current environment
	val, ok := env.Get("this")
	if !ok {
		return newError("'this' can only be used inside a model method")
	}
	return val
}

func evalIsOperatorWithTypeName(left Object, typeName string, env *Environment) Object {
	// Check for built-in type names first
	switch typeName {
	case typeNameInteger:
		return nativeBoolToBooleanObject(left.Type() == INTEGER_OBJ)
	case typeNameFloat:
		return nativeBoolToBooleanObject(left.Type() == FLOAT_OBJ)
	case typeNameString:
		return nativeBoolToBooleanObject(left.Type() == STRING_OBJ)
	case typeNameBoolean:
		return nativeBoolToBooleanObject(left.Type() == BOOLEAN_OBJ)
	case typeNameNull:
		return nativeBoolToBooleanObject(left.Type() == NULL_OBJ)
	case typeNameArray:
		return nativeBoolToBooleanObject(left.Type() == ARRAY_OBJ)
	case typeNameMap:
		return nativeBoolToBooleanObject(left.Type() == MAP_OBJ)
	case typeNameFunction:
		return nativeBoolToBooleanObject(left.Type() == FUNCTION_OBJ)
	}

	// Not a built-in type name, try to evaluate as identifier (Model or Enum)
	right := evalIdentifier(&ast.Identifier{Value: typeName}, env)
	if IsError(right) {
		return newError("undefined type: %s", typeName)
	}

	// Use the regular is operator evaluation
	return evalIsOperator(left, right)
}

func evalIsOperator(left, right Object) Object {
	// The `is` operator checks if left is an instance of the type on the right
	// right should be a Model or Enum type

	switch rightType := right.(type) {
	case *Model:
		// Check if left is an instance of this model
		if instance, ok := left.(*ModelInstance); ok {
			return nativeBoolToBooleanObject(instance.Model == rightType)
		}
		return FALSE

	case *Enum:
		// Check if left is a value of this enum
		if enumValue, ok := left.(*EnumValue); ok {
			return nativeBoolToBooleanObject(enumValue.EnumName == rightType.Name)
		}
		return FALSE

	default:
		return newError("'is' operator requires a type name or type value on the right side")
	}
}

func evalDbFindExpression(node *ast.DbFindExpression, env *Environment) Object {
	// Get db module from environment
	dbModule, exists := env.Get("db")
	if !exists || dbModule == nil {
		return newError("db module not imported")
	}

	module, ok := dbModule.(*Module)
	if !ok {
		return newError("db is not a module")
	}

	// Get database state from module scope
	stateWrapper, exists := module.Scope.Get("__db_state__")
	if !exists || stateWrapper == nil {
		return newError("database state not found - db module may not be properly initialized")
	}

	wrapper, ok := stateWrapper.(*DBStateWrapper)
	if !ok {
		return newError("invalid database state")
	}
	state := wrapper.State

	// Ensure database is open
	state.mu.Lock()
	defer state.mu.Unlock()
	if err := state.ensureOpen(); err != nil {
		return newError("database error: %s", err.Error())
	}

	// Evaluate model expression
	modelObj := Eval(node.Model, env)
	if IsError(modelObj) {
		return modelObj
	}

	model, ok := modelObj.(*Model)
	if !ok {
		return newError("db.find() requires a model type, got %s", modelObj.Type())
	}

	// Build and execute SQL query
	sql := "SELECT DISTINCT entity_id FROM data WHERE model_type = ?"
	args := []interface{}{model.Name}

	// Add conditions to query
	for i, cond := range node.Conditions {
		// Evaluate the condition value
		value := Eval(cond.Value, env)
		if IsError(value) {
			return value
		}

		// Extract field path from condition
		fieldPath := extractFieldPath(cond.Field)
		operator := translateOperator(cond.Operator)

		// Determine if we need numeric casting for comparison
		fieldExpr := "field_value"
		if _, ok := value.(*Integer); ok {
			fieldExpr = "CAST(field_value AS REAL)"
		} else if _, ok := value.(*Float); ok {
			fieldExpr = "CAST(field_value AS REAL)"
		}

		// Build subquery for this condition
		subquery := fmt.Sprintf("(SELECT entity_id FROM data WHERE model_type = ? AND field_name = ? AND %s %s ?)", fieldExpr, operator)
		if i > 0 {
			if cond.LogicOp == "||" {
				sql += " OR entity_id IN " + subquery
			} else {
				sql += " AND entity_id IN " + subquery
			}
		} else {
			sql += " AND entity_id IN " + subquery
		}
		args = append(args, model.Name, fieldPath, serializeForQuery(value))
	}

	// Execute query to get entity IDs
	rows, err := state.execer().Query(sql, args...)
	if err != nil {
		return newError("query error: %s", err.Error())
	}
	defer func() { _ = rows.Close() }()

	// Collect entity IDs
	entityIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return newError("scan error: %s", err.Error())
		}
		entityIDs = append(entityIDs, id)
	}

	// Handle query modifiers
	if node.Modifiers != nil {
		if node.Modifiers.Count {
			return &Integer{Value: int64(len(entityIDs))}
		}
	}

	// Load full instances
	results := &Array{Elements: []Object{}}
	for _, id := range entityIDs {
		instance := loadInstance(state, model, id)
		if IsError(instance) {
			return instance
		}
		results.Elements = append(results.Elements, instance)
	}

	// Apply modifiers
	if node.Modifiers != nil {
		// orderBy sorting
		if node.Modifiers.OrderBy != "" {
			sortInstances(results, node.Modifiers.OrderBy, node.Modifiers.OrderDesc)
		}

		// offset and limit
		if node.Modifiers.Offset != nil {
			offset := Eval(node.Modifiers.Offset, env)
			if IsError(offset) {
				return offset
			}
			offsetInt, ok := offset.(*Integer)
			if !ok {
				return newError("offset must be an integer")
			}
			offsetVal := int(offsetInt.Value)
			if offsetVal < 0 || offsetVal >= len(results.Elements) {
				results.Elements = []Object{}
			} else {
				results.Elements = results.Elements[offsetVal:]
			}
		}

		if node.Modifiers.Limit != nil {
			limit := Eval(node.Modifiers.Limit, env)
			if IsError(limit) {
				return limit
			}
			limitInt, ok := limit.(*Integer)
			if !ok {
				return newError("limit must be an integer")
			}
			limitVal := int(limitInt.Value)
			if limitVal > 0 && limitVal < len(results.Elements) {
				results.Elements = results.Elements[:limitVal]
			}
		}

		// first modifier - return single result or null
		if node.Modifiers.First {
			if len(results.Elements) == 0 {
				return NULL
			}
			return results.Elements[0]
		}
	}

	return results
}

// extractFieldPath extracts the field path from a query condition expression.
// For "this.x" returns "x", for "this.center.x" returns "center" (top-level field).
func extractFieldPath(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.MemberExpression:
		// Handle this.field
		if ident, ok := e.Object.(*ast.Identifier); ok && ident.Value == "this" {
			return e.Member.Value
		}
		// For nested fields like this.center.x, just return the top-level field
		if mem, ok := e.Object.(*ast.MemberExpression); ok {
			return extractFieldPath(mem)
		}
		return e.Member.Value
	case *ast.Identifier:
		return e.Value
	default:
		return ""
	}
}

// translateOperator converts TotalScript comparison operators to SQL operators.
func translateOperator(op string) string {
	switch op {
	case "==":
		return "="
	case "!=":
		return "!="
	case ">":
		return ">"
	case ">=":
		return ">="
	case "<":
		return "<"
	case "<=":
		return "<="
	default:
		return "="
	}
}

// serializeForQuery converts a TotalScript object to a string value for SQL queries.
func serializeForQuery(obj Object) string {
	switch v := obj.(type) {
	case *Integer:
		return fmt.Sprintf("%d", v.Value)
	case *Float:
		return fmt.Sprintf("%f", v.Value)
	case *String:
		return v.Value
	case *Boolean:
		if v.Value {
			return stringTrue
		}
		return stringFalse
	case *Null:
		return ""
	default:
		return obj.Inspect()
	}
}

// loadInstance loads a model instance from the database by entity ID.
func loadInstance(state *DBState, model *Model, entityID string) Object {
	instance := &ModelInstance{
		Model:  model,
		Fields: make(map[string]Object),
	}

	rows, err := state.execer().Query(
		"SELECT field_name, field_value, field_type FROM data WHERE entity_id = ?",
		entityID,
	)
	if err != nil {
		return newError("failed to load instance: %s", err.Error())
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var fieldName, fieldValue, fieldType string
		if err := rows.Scan(&fieldName, &fieldValue, &fieldType); err != nil {
			return newError("failed to scan field: %s", err.Error())
		}
		instance.Fields[fieldName] = deserializeDBValue(fieldValue, fieldType)
	}

	// Store entity_id for future operations
	instance.Fields["_entity_id"] = &String{Value: entityID}

	return instance
}

// sortInstances sorts an array of model instances by a field name.
func sortInstances(arr *Array, fieldName string, descending bool) {
	// Simple bubble sort for now
	n := len(arr.Elements)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			inst1, ok1 := arr.Elements[j].(*ModelInstance)
			inst2, ok2 := arr.Elements[j+1].(*ModelInstance)
			if !ok1 || !ok2 {
				continue
			}

			val1, exists1 := inst1.Fields[fieldName]
			val2, exists2 := inst2.Fields[fieldName]
			if !exists1 || !exists2 {
				continue
			}

			shouldSwap := false
			// Compare values
			switch v1 := val1.(type) {
			case *Integer:
				if v2, ok := val2.(*Integer); ok {
					if descending {
						shouldSwap = v1.Value > v2.Value
					} else {
						shouldSwap = v1.Value < v2.Value
					}
				}
			case *Float:
				if v2, ok := val2.(*Float); ok {
					if descending {
						shouldSwap = v1.Value > v2.Value
					} else {
						shouldSwap = v1.Value < v2.Value
					}
				}
			case *String:
				if v2, ok := val2.(*String); ok {
					if descending {
						shouldSwap = v1.Value > v2.Value
					} else {
						shouldSwap = v1.Value < v2.Value
					}
				}
			}

			if shouldSwap {
				arr.Elements[j], arr.Elements[j+1] = arr.Elements[j+1], arr.Elements[j]
			}
		}
	}
}
