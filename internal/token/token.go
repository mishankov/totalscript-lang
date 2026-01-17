// Package token defines the token types used in TotalScript lexical analysis.
package token

// TokenType represents the type of a lexical token.
type TokenType string

// Token represents a lexical token in TotalScript source code.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Token type constants.
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers and literals
	IDENT   TokenType = "IDENT"   // variable names, function names, etc.
	INTEGER TokenType = "INTEGER" // 42, -17, 0
	FLOAT   TokenType = "FLOAT"   // 3.14, -0.5
	STRING  TokenType = "STRING"  // "hello"

	// Operators
	ASSIGN     TokenType = "="
	PLUS       TokenType = "+"
	MINUS      TokenType = "-"
	ASTERISK   TokenType = "*"
	SLASH      TokenType = "/"
	SlashSlash TokenType = "//" // integer division
	PERCENT    TokenType = "%"
	POWER      TokenType = "**"

	// Comparison operators
	EQ    TokenType = "=="
	NotEq TokenType = "!="
	LT    TokenType = "<"
	GT    TokenType = ">"
	LtEq  TokenType = "<="
	GtEq  TokenType = ">="

	// Logical operators
	AND TokenType = "&&"
	OR  TokenType = "||"
	NOT TokenType = "!"

	// Compound assignment operators
	PlusAssign     TokenType = "+="
	MinusAssign    TokenType = "-="
	AsteriskAssign TokenType = "*="
	SlashAssign    TokenType = "/="
	PercentAssign  TokenType = "%="

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	DOT       TokenType = "."
	DotDot    TokenType = ".."  // range exclusive
	DotDotEq  TokenType = "..=" // range inclusive

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"

	// Type-related
	PIPE     TokenType = "|" // union type
	QUESTION TokenType = "?" // optional type
	AT       TokenType = "@" // annotation

	// Keywords
	VAR         TokenType = "VAR"
	CONST       TokenType = "CONST"
	FUNCTION    TokenType = "FUNCTION"
	MODEL       TokenType = "MODEL"
	ENUM        TokenType = "ENUM"
	IF          TokenType = "IF"
	ELSE        TokenType = "ELSE"
	SWITCH      TokenType = "SWITCH"
	CASE        TokenType = "CASE"
	DEFAULT     TokenType = "DEFAULT"
	FOR         TokenType = "FOR"
	WHILE       TokenType = "WHILE"
	IN          TokenType = "IN"
	RETURN      TokenType = "RETURN"
	BREAK       TokenType = "BREAK"
	CONTINUE    TokenType = "CONTINUE"
	IMPORT      TokenType = "IMPORT"
	AS          TokenType = "AS"
	THIS        TokenType = "THIS"
	IS          TokenType = "IS"
	TRUE        TokenType = "TRUE"
	FALSE       TokenType = "FALSE"
	NULL        TokenType = "NULL"
	CONSTRUCTOR TokenType = "CONSTRUCTOR"

	// Database query modifiers
	ORDERBY TokenType = "ORDERBY"
	LIMIT   TokenType = "LIMIT"
	OFFSET  TokenType = "OFFSET"
	FIRST   TokenType = "FIRST"
	COUNT   TokenType = "COUNT"
	DESC    TokenType = "DESC"
)

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"var":         VAR,
	"const":       CONST,
	"function":    FUNCTION,
	"model":       MODEL,
	"enum":        ENUM,
	"if":          IF,
	"else":        ELSE,
	"switch":      SWITCH,
	"case":        CASE,
	"default":     DEFAULT,
	"for":         FOR,
	"while":       WHILE,
	"in":          IN,
	"return":      RETURN,
	"break":       BREAK,
	"continue":    CONTINUE,
	"import":      IMPORT,
	"as":          AS,
	"this":        THIS,
	"is":          IS,
	"true":        TRUE,
	"false":       FALSE,
	"null":        NULL,
	"constructor": CONSTRUCTOR,
	"orderBy":     ORDERBY,
	"limit":       LIMIT,
	"offset":      OFFSET,
	"desc":        DESC,
}

// LookupIdent checks if the given identifier is a keyword.
// Returns the keyword token type if found, otherwise returns IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
