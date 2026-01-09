package stdlib

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

func TestPrintln(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []interpreter.Object
	}{
		{
			name: "no arguments",
			args: []interpreter.Object{},
		},
		{
			name: "single integer",
			args: []interpreter.Object{&interpreter.Integer{Value: 42}},
		},
		{
			name: "multiple arguments",
			args: []interpreter.Object{
				&interpreter.String{Value: "hello"},
				&interpreter.Integer{Value: 42},
				&interpreter.Boolean{Value: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Println.Fn(tt.args...)
			if result != interpreter.NULL {
				t.Errorf("expected NULL, got %s", result.Type())
			}
		})
	}
}

func TestTypeof(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []interpreter.Object
		expected string
		isError  bool
	}{
		{
			name:     "integer",
			args:     []interpreter.Object{&interpreter.Integer{Value: 42}},
			expected: "INTEGER",
			isError:  false,
		},
		{
			name:     "float",
			args:     []interpreter.Object{&interpreter.Float{Value: 3.14}},
			expected: "FLOAT",
			isError:  false,
		},
		{
			name:     "string",
			args:     []interpreter.Object{&interpreter.String{Value: "hello"}},
			expected: "STRING",
			isError:  false,
		},
		{
			name:     "boolean",
			args:     []interpreter.Object{interpreter.TRUE},
			expected: "BOOLEAN",
			isError:  false,
		},
		{
			name:     "null",
			args:     []interpreter.Object{interpreter.NULL},
			expected: "NULL",
			isError:  false,
		},
		{
			name:     "array",
			args:     []interpreter.Object{&interpreter.Array{Elements: []interpreter.Object{}}},
			expected: "ARRAY",
			isError:  false,
		},
		{
			name:    "no arguments",
			args:    []interpreter.Object{},
			isError: true,
		},
		{
			name: "too many arguments",
			args: []interpreter.Object{
				&interpreter.Integer{Value: 1},
				&interpreter.Integer{Value: 2},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Typeof.Fn(tt.args...)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			str, ok := result.(*interpreter.String)
			if !ok {
				t.Errorf("expected STRING, got %s", result.Type())
				return
			}

			if str.Value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, str.Value)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []interpreter.Object
		expected int64
		isError  bool
	}{
		{
			name:     "integer to integer",
			args:     []interpreter.Object{&interpreter.Integer{Value: 42}},
			expected: 42,
			isError:  false,
		},
		{
			name:     "float to integer",
			args:     []interpreter.Object{&interpreter.Float{Value: 3.14}},
			expected: 3,
			isError:  false,
		},
		{
			name:     "string to integer",
			args:     []interpreter.Object{&interpreter.String{Value: "123"}},
			expected: 123,
			isError:  false,
		},
		{
			name:     "negative string to integer",
			args:     []interpreter.Object{&interpreter.String{Value: "-456"}},
			expected: -456,
			isError:  false,
		},
		{
			name:     "true to integer",
			args:     []interpreter.Object{interpreter.TRUE},
			expected: 1,
			isError:  false,
		},
		{
			name:     "false to integer",
			args:     []interpreter.Object{interpreter.FALSE},
			expected: 0,
			isError:  false,
		},
		{
			name:    "invalid string",
			args:    []interpreter.Object{&interpreter.String{Value: "abc"}},
			isError: true,
		},
		{
			name:    "null to integer",
			args:    []interpreter.Object{interpreter.NULL},
			isError: true,
		},
		{
			name:    "no arguments",
			args:    []interpreter.Object{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Integer.Fn(tt.args...)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			intObj, ok := result.(*interpreter.Integer)
			if !ok {
				t.Errorf("expected INTEGER, got %s", result.Type())
				return
			}

			if intObj.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, intObj.Value)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []interpreter.Object
		expected float64
		isError  bool
	}{
		{
			name:     "float to float",
			args:     []interpreter.Object{&interpreter.Float{Value: 3.14}},
			expected: 3.14,
			isError:  false,
		},
		{
			name:     "integer to float",
			args:     []interpreter.Object{&interpreter.Integer{Value: 42}},
			expected: 42.0,
			isError:  false,
		},
		{
			name:     "string to float",
			args:     []interpreter.Object{&interpreter.String{Value: "3.14"}},
			expected: 3.14,
			isError:  false,
		},
		{
			name:     "negative string to float",
			args:     []interpreter.Object{&interpreter.String{Value: "-2.5"}},
			expected: -2.5,
			isError:  false,
		},
		{
			name:     "true to float",
			args:     []interpreter.Object{interpreter.TRUE},
			expected: 1.0,
			isError:  false,
		},
		{
			name:     "false to float",
			args:     []interpreter.Object{interpreter.FALSE},
			expected: 0.0,
			isError:  false,
		},
		{
			name:    "invalid string",
			args:    []interpreter.Object{&interpreter.String{Value: "abc"}},
			isError: true,
		},
		{
			name:    "null to float",
			args:    []interpreter.Object{interpreter.NULL},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Float.Fn(tt.args...)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			floatObj, ok := result.(*interpreter.Float)
			if !ok {
				t.Errorf("expected FLOAT, got %s", result.Type())
				return
			}

			if floatObj.Value != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, floatObj.Value)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []interpreter.Object
		expected string
		isError  bool
	}{
		{
			name:     "string to string",
			args:     []interpreter.Object{&interpreter.String{Value: "hello"}},
			expected: "hello",
			isError:  false,
		},
		{
			name:     "integer to string",
			args:     []interpreter.Object{&interpreter.Integer{Value: 42}},
			expected: "42",
			isError:  false,
		},
		{
			name:     "float to string",
			args:     []interpreter.Object{&interpreter.Float{Value: 3.14}},
			expected: "3.14",
			isError:  false,
		},
		{
			name:     "true to string",
			args:     []interpreter.Object{interpreter.TRUE},
			expected: "true",
			isError:  false,
		},
		{
			name:     "false to string",
			args:     []interpreter.Object{interpreter.FALSE},
			expected: "false",
			isError:  false,
		},
		{
			name:     "null to string",
			args:     []interpreter.Object{interpreter.NULL},
			expected: "null",
			isError:  false,
		},
		{
			name:    "no arguments",
			args:    []interpreter.Object{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := String.Fn(tt.args...)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			strObj, ok := result.(*interpreter.String)
			if !ok {
				t.Errorf("expected STRING, got %s", result.Type())
				return
			}

			if strObj.Value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, strObj.Value)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []interpreter.Object
		expected bool
		isError  bool
	}{
		{
			name:     "true to boolean",
			args:     []interpreter.Object{interpreter.TRUE},
			expected: true,
			isError:  false,
		},
		{
			name:     "false to boolean",
			args:     []interpreter.Object{interpreter.FALSE},
			expected: false,
			isError:  false,
		},
		{
			name:     "zero integer to boolean",
			args:     []interpreter.Object{&interpreter.Integer{Value: 0}},
			expected: false,
			isError:  false,
		},
		{
			name:     "non-zero integer to boolean",
			args:     []interpreter.Object{&interpreter.Integer{Value: 42}},
			expected: true,
			isError:  false,
		},
		{
			name:     "zero float to boolean",
			args:     []interpreter.Object{&interpreter.Float{Value: 0.0}},
			expected: false,
			isError:  false,
		},
		{
			name:     "non-zero float to boolean",
			args:     []interpreter.Object{&interpreter.Float{Value: 3.14}},
			expected: true,
			isError:  false,
		},
		{
			name:     "empty string to boolean",
			args:     []interpreter.Object{&interpreter.String{Value: ""}},
			expected: false,
			isError:  false,
		},
		{
			name:     "non-empty string to boolean",
			args:     []interpreter.Object{&interpreter.String{Value: "hello"}},
			expected: true,
			isError:  false,
		},
		{
			name:     "null to boolean",
			args:     []interpreter.Object{interpreter.NULL},
			expected: false,
			isError:  false,
		},
		{
			name:    "no arguments",
			args:    []interpreter.Object{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Boolean.Fn(tt.args...)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			boolObj, ok := result.(*interpreter.Boolean)
			if !ok {
				t.Errorf("expected BOOLEAN, got %s", result.Type())
				return
			}

			if boolObj.Value != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, boolObj.Value)
			}
		})
	}
}
