// Package parser implements the TotalScript parser.
package parser

import (
	"fmt"
	"strconv"

	"github.com/mishankov/totalscript-lang/internal/ast"
	"github.com/mishankov/totalscript-lang/internal/lexer"
	"github.com/mishankov/totalscript-lang/internal/token"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	ASSIGN      // =, +=, -=, etc.
	OR          // ||
	AND         // &&
	EQUALS      // ==, !=
	LESSGREATER // >, <, >=, <=
	IS          // is
	RANGE       // .., ..=
	SUM         // +, -
	PRODUCT     // *, /, //, %
	POWER       // **
	PREFIX      // -x, !x
	CALL        // myFunction(x)
	INDEX       // array[index], map[key]
	MEMBER      // obj.property
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:          ASSIGN,
	token.PLUS_ASSIGN:     ASSIGN,
	token.MINUS_ASSIGN:    ASSIGN,
	token.ASTERISK_ASSIGN: ASSIGN,
	token.SLASH_ASSIGN:    ASSIGN,
	token.PERCENT_ASSIGN:  ASSIGN,
	token.OR:              OR,
	token.AND:             AND,
	token.EQ:              EQUALS,
	token.NOT_EQ:          EQUALS,
	token.LT:              LESSGREATER,
	token.GT:              LESSGREATER,
	token.LT_EQ:           LESSGREATER,
	token.GT_EQ:           LESSGREATER,
	token.IS:              IS,
	token.DOTDOT:          RANGE,
	token.DOTDOTEQ:        RANGE,
	token.PLUS:            SUM,
	token.MINUS:           SUM,
	token.SLASH:           PRODUCT,
	token.SLASHSLASH:      PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.PERCENT:         PRODUCT,
	token.POWER:           POWER,
	token.LPAREN:          CALL,
	token.LBRACKET:        INDEX,
	token.DOT:             MEMBER,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser represents a TotalScript parser.
type Parser struct {
	l      *lexer.Lexer
	errors []*ParseError

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []*ParseError{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INTEGER, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)
	p.registerPrefix(token.MODEL, p.parseModelLiteral)
	p.registerPrefix(token.ENUM, p.parseEnumLiteral)
	p.registerPrefix(token.THIS, p.parseThisExpression)
	p.registerPrefix(token.DOTDOT, p.parsePrefixRangeExpression)
	p.registerPrefix(token.DOTDOTEQ, p.parsePrefixRangeExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.SLASHSLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.POWER, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.PLUS_ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.MINUS_ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK_ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.SLASH_ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.PERCENT_ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.DOTDOT, p.parseRangeExpression)
	p.registerInfix(token.DOTDOTEQ, p.parseRangeExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors.
func (p *Parser) Errors() []*ParseError {
	return p.errors
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.addError(msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(msg)
}

// ParseProgram parses the program and returns the AST.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for type annotation
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		stmt.Type = p.parseTypeExpression()
	}

	// Check for initialization
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // consume '='
		p.nextToken() // move to value
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for type annotation
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		stmt.Type = p.parseTypeExpression()
	}

	// Expect initialization
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// Check if there's a return value
	if !p.curTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	return &ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	return &ast.ContinueStatement{Token: p.curToken}
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	p.nextToken()

	// Check if it's a for-in loop
	if p.curTokenIs(token.IDENT) {
		firstIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()

		// Check for comma (index, value pattern)
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			if !p.curTokenIs(token.IDENT) {
				p.addError("expected identifier after comma in for loop")
				return nil
			}
			stmt.Iterator = firstIdent
			stmt.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			p.nextToken()
		} else {
			stmt.Value = firstIdent
		}

		// Expect 'in'
		if !p.curTokenIs(token.IN) {
			// Not a for-in loop, might be C-style
			// Rewind and parse as C-style
			return p.parseForStatementCStyle()
		}

		p.nextToken()
		stmt.Iterable = p.parseExpression(LOWEST)
		stmt.IsRangeStyle = true

	} else if p.curTokenIs(token.VAR) {
		// C-style for loop
		return p.parseForStatementCStyle()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForStatementCStyle() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken, IsRangeStyle: false}

	// Init statement
	stmt.Init = p.parseStatement()
	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken()
	stmt.Post = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	stmt := &ast.SwitchStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.CASE) {
			caseClause := p.parseCaseClause()
			if caseClause != nil {
				stmt.Cases = append(stmt.Cases, caseClause)
			}
		} else if p.curTokenIs(token.DEFAULT) {
			p.nextToken()
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			stmt.Default = p.parseBlockStatement()
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseCaseClause() *ast.CaseClause {
	clause := &ast.CaseClause{Token: p.curToken}

	p.nextToken()
	clause.Values = []ast.Expression{p.parseExpression(LOWEST)}

	// Check for comma-separated values
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		clause.Values = append(clause.Values, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	clause.Body = p.parseBlockStatement()

	return clause
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.addError(msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if p.peekTokenIs(token.IF) {
			// else if
			p.nextToken()
			elseIf := p.parseIfExpression()
			// Wrap in a block
			expression.Alternative = &ast.BlockStatement{
				Token: p.curToken,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      p.curToken,
						Expression: elseIf,
					},
				},
			}
		} else if p.expectPeek(token.LBRACE) {
			expression.Alternative = p.parseBlockStatement()
		}
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	// Check for return type
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		lit.ReturnType = p.parseTypeExpression()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := &ast.Parameter{
		Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// Check for type annotation
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		param.Type = p.parseTypeExpression()
	}

	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to identifier

		param := &ast.Parameter{
			Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			p.nextToken()
			param.Type = p.parseTypeExpression()
		}

		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	// After parsing the index expression, we should be at RBRACKET
	// (range expressions may leave us there) or the next token should be RBRACKET
	if !p.curTokenIs(token.RBRACKET) {
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	}

	return exp
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Token: p.curToken, Object: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	exp := &ast.RangeExpression{
		Token:     p.curToken,
		Start:     left,
		Inclusive: p.curTokenIs(token.DOTDOTEQ),
	}

	precedence := p.curPrecedence()
	p.nextToken()

	// Check for open-ended range (e.g., 5..)
	// If the next token is a delimiter like ], ), or comma, the range has no end
	if p.curTokenIs(token.RBRACKET) || p.curTokenIs(token.RPAREN) || p.curTokenIs(token.COMMA) {
		exp.End = nil
	} else {
		exp.End = p.parseExpression(precedence)
	}

	return exp
}

func (p *Parser) parsePrefixRangeExpression() ast.Expression {
	exp := &ast.RangeExpression{
		Token:     p.curToken,
		Start:     nil, // Open-ended start (e.g., ..5)
		Inclusive: p.curTokenIs(token.DOTDOTEQ),
	}

	p.nextToken()

	// Check for fully open-ended range (e.g., ..)
	if p.curTokenIs(token.RBRACKET) || p.curTokenIs(token.RPAREN) || p.curTokenIs(token.COMMA) {
		exp.End = nil
	} else {
		exp.End = p.parseExpression(RANGE)
	}

	return exp
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLiteral{Token: p.curToken}
	mapLit.Pairs = make(map[ast.Expression]ast.Expression)

	// Check for empty map
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return mapLit
	}

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		mapLit.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return mapLit
}

func (p *Parser) parseTypeExpression() *ast.TypeExpression {
	typeExpr := &ast.TypeExpression{Token: p.curToken}

	// Handle union types (integer | string)
	if p.curTokenIs(token.IDENT) {
		typeExpr.Name = p.curToken.Literal

		// Check for generic type parameter
		if p.peekTokenIs(token.LT) {
			p.nextToken() // consume '<'
			p.nextToken() // move to first type
			typeExpr.Generic = append(typeExpr.Generic, p.curToken.Literal)

			for p.peekTokenIs(token.COMMA) {
				p.nextToken() // consume comma
				p.nextToken() // move to next type
				typeExpr.Generic = append(typeExpr.Generic, p.curToken.Literal)
			}

			if !p.expectPeek(token.GT) {
				return nil
			}
		}

		// Check for union (|)
		if p.peekTokenIs(token.PIPE) {
			typeExpr.Union = []string{typeExpr.Name}
			typeExpr.Name = ""

			for p.peekTokenIs(token.PIPE) {
				p.nextToken() // consume '|'
				p.nextToken() // move to type
				typeExpr.Union = append(typeExpr.Union, p.curToken.Literal)
			}
		}

		// Check for optional (?)
		if p.peekTokenIs(token.QUESTION) {
			p.nextToken()
			typeExpr.Optional = true
		}
	}

	return typeExpr
}

func (p *Parser) parseModelLiteral() ast.Expression {
	model := &ast.ModelLiteral{Token: p.curToken}
	model.Fields = []*ast.ModelField{}
	model.Methods = []*ast.ModelMethod{}
	model.Constructors = []*ast.FunctionLiteral{}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Parse field, method, or constructor
		if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.CONSTRUCTOR) {
			msg := "expected field, method, or constructor name in model"
			p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
			return nil
		}

		name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// Check if it's a constructor, method, or field
		switch {
		case p.curTokenIs(token.CONSTRUCTOR) && p.peekTokenIs(token.ASSIGN):
			// It's a constructor (constructor = function)
			p.nextToken() // consume '='
			p.nextToken() // move to 'function'

			if !p.curTokenIs(token.FUNCTION) {
				msg := "expected function after '=' in constructor"
				p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
				return nil
			}

			fn := p.parseFunctionLiteral()
			if fn == nil {
				return nil
			}

			fnLit, ok := fn.(*ast.FunctionLiteral)
			if !ok {
				msg := "expected function literal"
				p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
				return nil
			}

			model.Constructors = append(model.Constructors, fnLit)

		case p.peekTokenIs(token.ASSIGN):
			// It's a method (name = function)
			p.nextToken() // consume '='
			p.nextToken() // move to 'function'

			if !p.curTokenIs(token.FUNCTION) {
				msg := "expected function after '=' in model method"
				p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
				return nil
			}

			fn := p.parseFunctionLiteral()
			if fn == nil {
				return nil
			}

			fnLit, ok := fn.(*ast.FunctionLiteral)
			if !ok {
				msg := "expected function literal"
				p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
				return nil
			}

			method := &ast.ModelMethod{
				Name:     name,
				Function: fnLit,
			}
			model.Methods = append(model.Methods, method)

		case p.peekTokenIs(token.COLON):
			// It's a field
			p.nextToken() // consume ':'
			p.nextToken() // move to type

			typeExpr := p.parseTypeExpression()
			if typeExpr == nil {
				return nil
			}

			field := &ast.ModelField{
				Name: name,
				Type: typeExpr,
			}
			model.Fields = append(model.Fields, field)

		default:
			msg := "expected ':' or '=' after field/method name in model"
			p.errors = append(p.errors, NewParseError(p.curToken.Line, p.curToken.Column, msg))
			return nil
		}

		p.nextToken()
	}

	return model
}

func (p *Parser) parseEnumLiteral() ast.Expression {
	enum := &ast.EnumLiteral{Token: p.curToken}
	enum.Values = []*ast.EnumValue{}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Parse enum value
		if !p.curTokenIs(token.IDENT) {
			p.peekError(token.IDENT)
			return nil
		}

		name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.ASSIGN) {
			return nil
		}

		p.nextToken() // move to value expression

		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}

		enumValue := &ast.EnumValue{
			Name:  name,
			Value: value,
		}
		enum.Values = append(enum.Values, enumValue)

		p.nextToken()
	}

	return enum
}

func (p *Parser) parseThisExpression() ast.Expression {
	return &ast.ThisExpression{Token: p.curToken}
}
