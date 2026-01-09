package interpreter

import (
	"fmt"

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
		env.Set(node.Name.Value, val)
		return val

	case *ast.ConstStatement:
		val := Eval(node.Value, env)
		if IsError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val

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
		return applyFunction(function, args)

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
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalAssignmentExpression(node *ast.InfixExpression, env *Environment) Object {
	// Get the identifier name from the left side
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return newError("cannot assign to %T", node.Left)
	}

	// Evaluate the right side
	val := Eval(node.Right, env)
	if IsError(val) {
		return val
	}

	// Handle compound assignment operators
	if node.Operator != "=" {
		// Get current value
		currentVal, ok := env.Get(ident.Value)
		if !ok {
			return newError("identifier not found: %s", ident.Value)
		}

		// Determine the arithmetic operator
		var op string
		switch node.Operator {
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

	// Set the variable
	env.Set(ident.Value, val)
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

func applyFunction(fn Object, args []Object) Object {
	function, ok := fn.(*Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	if len(args) != len(function.Parameters) {
		return newError("wrong number of arguments: expected %d, got %d",
			len(function.Parameters), len(args))
	}

	extendedEnv := extendFunctionEnv(function, args)
	evaluated := Eval(function.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
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

	// For now, we don't support member access
	// This would require implementing methods on types
	return newError("member access not yet supported")
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

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
