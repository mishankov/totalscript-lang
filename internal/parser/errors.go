package parser

import "fmt"

// ParseError represents a parser error with location information.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Message)
}

// NewParseError creates a new ParseError.
func NewParseError(line, column int, message string) *ParseError {
	return &ParseError{
		Line:    line,
		Column:  column,
		Message: message,
	}
}
