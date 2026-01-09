// Package lexer implements the lexical analyzer for TotalScript.
package lexer

import (
	"github.com/mishankov/totalscript-lang/internal/token"
)

// Lexer performs lexical analysis on TotalScript source code.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number (1-indexed)
	column       int  // current column number (1-indexed)
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position in the input string.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing the position.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharN returns the character n positions ahead without advancing.
func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.EQ)
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.PLUS_ASSIGN)
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.MINUS_ASSIGN)
		} else {
			tok = l.newToken(token.MINUS, l.ch)
		}
	case '*':
		if l.peekChar() == '*' {
			tok = l.makeTwoCharToken(token.POWER)
		} else if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.ASTERISK_ASSIGN)
		} else {
			tok = l.newToken(token.ASTERISK, l.ch)
		}
	case '/':
		if l.peekChar() == '/' {
			tok = l.makeTwoCharToken(token.SLASHSLASH)
		} else if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.SLASH_ASSIGN)
		} else {
			tok = l.newToken(token.SLASH, l.ch)
		}
	case '%':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.PERCENT_ASSIGN)
		} else {
			tok = l.newToken(token.PERCENT, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.NOT_EQ)
		} else {
			tok = l.newToken(token.NOT, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.LT_EQ)
		} else {
			tok = l.newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.GT_EQ)
		} else {
			tok = l.newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			tok = l.makeTwoCharToken(token.AND)
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			tok = l.makeTwoCharToken(token.OR)
		} else {
			tok = l.newToken(token.PIPE, l.ch)
		}
	case '.':
		if l.peekChar() == '.' {
			if l.peekCharN(2) == '=' {
				tok = l.makeThreeCharToken(token.DOTDOTEQ)
			} else {
				tok = l.makeTwoCharToken(token.DOTDOT)
			}
		} else {
			tok = l.newToken(token.DOT, l.ch)
		}
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case '?':
		tok = l.newToken(token.QUESTION, l.ch)
	case '@':
		tok = l.newToken(token.AT, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = l.readNumber()
			return tok
		}
		tok = l.newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

// newToken creates a new token with the given type and character.
func (l *Lexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    l.line,
		Column:  l.column,
	}
}

// makeTwoCharToken creates a two-character token.
func (l *Lexer) makeTwoCharToken(tokenType token.TokenType) token.Token {
	ch := l.ch
	l.readChar()
	literal := string(ch) + string(l.ch)
	return token.Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column - 1,
	}
}

// makeThreeCharToken creates a three-character token.
func (l *Lexer) makeThreeCharToken(tokenType token.TokenType) token.Token {
	first := l.ch
	l.readChar()
	second := l.ch
	l.readChar()
	literal := string(first) + string(second) + string(l.ch)
	return token.Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column - 2,
	}
}

// readIdentifier reads an identifier and advances the lexer's position.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float) and returns its type and literal.
func (l *Lexer) readNumber() (token.TokenType, string) {
	position := l.position
	tokenType := token.INTEGER

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = token.FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return tokenType, l.input[position:l.position]
}

// readString reads a string literal, handling escape sequences.
func (l *Lexer) readString() string {
	var result []byte
	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				result = append(result, '\\', l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	// Skip the closing quote
	l.readChar()

	return string(result)
}

// skipWhitespace skips whitespace characters (space, tab, carriage return, newline).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips single-line (#) and multi-line (###) comments.
func (l *Lexer) skipComment() {
	if l.ch != '#' {
		return
	}

	// Check for multi-line comment (###)
	if l.peekChar() == '#' && l.peekCharN(2) == '#' {
		l.readChar() // skip first #
		l.readChar() // skip second #
		l.readChar() // skip third #

		// Read until closing ###
		for {
			if l.ch == 0 {
				break
			}
			if l.ch == '#' && l.peekChar() == '#' && l.peekCharN(2) == '#' {
				l.readChar() // skip first #
				l.readChar() // skip second #
				l.readChar() // skip third #
				break
			}
			l.readChar()
		}
	} else {
		// Single-line comment
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}

	// Skip any whitespace after comment and check for more comments
	l.skipWhitespace()
	if l.ch == '#' {
		l.skipComment()
	}
}

// isLetter returns true if the character is a letter or underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit returns true if the character is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
