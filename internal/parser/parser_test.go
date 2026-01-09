package parser

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/ast"
	"github.com/mishankov/totalscript-lang/internal/lexer"
)

func TestVarStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple var", "var x = 5", "var x = 5"},
		{"var with type", "var x: integer = 5", "var x: integer = 5"},
		{"var without value", "var x: integer", "var x: integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt := program.Statements[0]
			if stmt.String() != tt.expected {
				t.Errorf("stmt.String() wrong. expected=%q, got=%q", tt.expected, stmt.String())
			}
		})
	}
}

func TestConstStatements(t *testing.T) {
	input := `const x = 5`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstStatement)
	if !ok {
		t.Fatalf("statement is not *ast.ConstStatement. got=%T", program.Statements[0])
	}

	if stmt.Name.Value != "x" {
		t.Errorf("stmt.Name.Value wrong. expected=%q, got=%q", "x", stmt.Name.Value)
	}
}

func TestImportStatements(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedPath   string
		expectedAlias  string
		expectedModule string
	}{
		{"stdlib module", `import "math"`, "math", "", "math"},
		{"file module", `import "./utils"`, "./utils", "", "utils"},
		{"nested file module", `import "./lib/helpers"`, "./lib/helpers", "", "helpers"},
		{"module with alias", `import "math" as m`, "math", "m", "m"},
		{"file with alias", `import "./geometry" as geo`, "./geometry", "geo", "geo"},
		{"file with .tsl extension", `import "./utils.tsl"`, "./utils.tsl", "", "utils"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ImportStatement)
			if !ok {
				t.Fatalf("statement is not *ast.ImportStatement. got=%T", program.Statements[0])
			}

			if stmt.Path != tt.expectedPath {
				t.Errorf("stmt.Path wrong. expected=%q, got=%q", tt.expectedPath, stmt.Path)
			}

			if stmt.Alias != tt.expectedAlias {
				t.Errorf("stmt.Alias wrong. expected=%q, got=%q", tt.expectedAlias, stmt.Alias)
			}

			if stmt.ModuleName != tt.expectedModule {
				t.Errorf("stmt.ModuleName wrong. expected=%q, got=%q", tt.expectedModule, stmt.ModuleName)
			}
		})
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"return with value", "return 5", "return 5"},
		{"return with expression", "return x + y", "return (x + y)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt := program.Statements[0]
			if stmt.String() != tt.expected {
				t.Errorf("stmt.String() wrong. expected=%q, got=%q", tt.expected, stmt.String())
			}
		})
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "3.14"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp not *ast.FloatLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 3.14 {
		t.Errorf("literal.Value not %f. got=%f", 3.14, literal.Value)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		boolean, ok := stmt.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("exp not *ast.BooleanLiteral. got=%T", stmt.Expression)
		}

		if boolean.Value != tt.expected {
			t.Errorf("boolean.Value not %t. got=%t", tt.expected, boolean.Value)
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true", "!", true},
		{"!false", "!", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp is not ast.InfixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x < y { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if exp.Condition.String() != "(x < y)" {
		t.Errorf("exp.Condition is not '(x < y)'. got=%s", exp.Condition.String())
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d",
			len(exp.Consequence.Statements))
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `function(x, y) { x + y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d",
			len(function.Parameters))
	}

	if function.Parameters[0].Name.Value != "x" {
		t.Fatalf("parameter 0 is not 'x'. got=%s", function.Parameters[0].Name.Value)
	}

	if function.Parameters[1].Name.Value != "y" {
		t.Fatalf("parameter 1 is not 'y'. got=%s", function.Parameters[1].Name.Value)
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statement. got=%d",
			len(function.Body.Statements))
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5)"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}
}

func TestIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp not *ast.IndexExpression. got=%T", stmt.Expression)
	}

	if indexExp.Left.String() != "myArray" {
		t.Fatalf("indexExp.Left not 'myArray'. got=%s", indexExp.Left.String())
	}

	if indexExp.Index.String() != "(1 + 1)" {
		t.Fatalf("indexExp.Index not '(1 + 1)'. got=%s", indexExp.Index.String())
	}
}

func TestMapLiterals(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	mapLit, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(mapLit.Pairs) != 3 {
		t.Errorf("map.Pairs has wrong length. got=%d", len(mapLit.Pairs))
	}
}

func TestWhileStatement(t *testing.T) {
	input := `while x < 10 { x = x + 1 }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.WhileStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Condition.String() != "(x < 10)" {
		t.Errorf("stmt.Condition wrong. got=%s", stmt.Condition.String())
	}
}

func TestRangeExpression(t *testing.T) {
	tests := []struct {
		input     string
		expected  string
		inclusive bool
	}{
		{"0..10", "0..10", false},
		{"0..=10", "0..=10", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		rangeExp, ok := stmt.Expression.(*ast.RangeExpression)
		if !ok {
			t.Fatalf("exp not *ast.RangeExpression. got=%T", stmt.Expression)
		}

		if rangeExp.String() != tt.expected {
			t.Errorf("rangeExp.String() wrong. expected=%q, got=%q",
				tt.expected, rangeExp.String())
		}

		if rangeExp.Inclusive != tt.inclusive {
			t.Errorf("rangeExp.Inclusive wrong. expected=%t, got=%t",
				tt.inclusive, rangeExp.Inclusive)
		}
	}
}

func TestMemberExpression(t *testing.T) {
	input := "obj.property"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	memberExp, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("exp not *ast.MemberExpression. got=%T", stmt.Expression)
	}

	if memberExp.Object.String() != "obj" {
		t.Errorf("memberExp.Object wrong. got=%s", memberExp.Object.String())
	}

	if memberExp.Member.Value != "property" {
		t.Errorf("memberExp.Member wrong. got=%s", memberExp.Member.Value)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}
