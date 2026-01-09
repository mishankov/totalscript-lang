package stdlib

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

func TestStringLength(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	lengthFn := methods["length"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		expected int64
	}{
		{
			name:     "empty string",
			receiver: &interpreter.String{Value: ""},
			expected: 0,
		},
		{
			name:     "simple string",
			receiver: &interpreter.String{Value: "hello"},
			expected: 5,
		},
		{
			name:     "string with spaces",
			receiver: &interpreter.String{Value: "hello world"},
			expected: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := lengthFn(tt.receiver)

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

func TestStringUpper(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	upperFn := methods["upper"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		expected string
	}{
		{
			name:     "lowercase to uppercase",
			receiver: &interpreter.String{Value: "hello"},
			expected: "HELLO",
		},
		{
			name:     "mixed case",
			receiver: &interpreter.String{Value: "HeLLo WoRLd"},
			expected: "HELLO WORLD",
		},
		{
			name:     "already uppercase",
			receiver: &interpreter.String{Value: "HELLO"},
			expected: "HELLO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := upperFn(tt.receiver)

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

func TestStringLower(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	lowerFn := methods["lower"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		expected string
	}{
		{
			name:     "uppercase to lowercase",
			receiver: &interpreter.String{Value: "HELLO"},
			expected: "hello",
		},
		{
			name:     "mixed case",
			receiver: &interpreter.String{Value: "HeLLo WoRLd"},
			expected: "hello world",
		},
		{
			name:     "already lowercase",
			receiver: &interpreter.String{Value: "hello"},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := lowerFn(tt.receiver)

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

func TestStringTrim(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	trimFn := methods["trim"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		expected string
	}{
		{
			name:     "trim spaces",
			receiver: &interpreter.String{Value: "  hello  "},
			expected: "hello",
		},
		{
			name:     "trim tabs and newlines",
			receiver: &interpreter.String{Value: "\t\nhello\n\t"},
			expected: "hello",
		},
		{
			name:     "no whitespace",
			receiver: &interpreter.String{Value: "hello"},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := trimFn(tt.receiver)

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

func TestStringSplit(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	splitFn := methods["split"]

	tests := []struct {
		name      string
		receiver  *interpreter.String
		separator *interpreter.String
		expected  []string
	}{
		{
			name:      "split by comma",
			receiver:  &interpreter.String{Value: "a,b,c"},
			separator: &interpreter.String{Value: ","},
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split by space",
			receiver:  &interpreter.String{Value: "hello world"},
			separator: &interpreter.String{Value: " "},
			expected:  []string{"hello", "world"},
		},
		{
			name:      "no separator found",
			receiver:  &interpreter.String{Value: "hello"},
			separator: &interpreter.String{Value: ","},
			expected:  []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := splitFn(tt.receiver, tt.separator)

			arrObj, ok := result.(*interpreter.Array)
			if !ok {
				t.Errorf("expected ARRAY, got %s", result.Type())
				return
			}

			if len(arrObj.Elements) != len(tt.expected) {
				t.Errorf("expected %d elements, got %d", len(tt.expected), len(arrObj.Elements))
				return
			}

			for i, expected := range tt.expected {
				strObj, ok := arrObj.Elements[i].(*interpreter.String)
				if !ok {
					t.Errorf("element %d: expected STRING, got %s", i, arrObj.Elements[i].Type())
					continue
				}

				if strObj.Value != expected {
					t.Errorf("element %d: expected %s, got %s", i, expected, strObj.Value)
				}
			}
		})
	}
}

func TestStringContains(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	containsFn := methods["contains"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		search   *interpreter.String
		expected bool
	}{
		{
			name:     "contains substring",
			receiver: &interpreter.String{Value: "hello world"},
			search:   &interpreter.String{Value: "world"},
			expected: true,
		},
		{
			name:     "does not contain substring",
			receiver: &interpreter.String{Value: "hello world"},
			search:   &interpreter.String{Value: "foo"},
			expected: false,
		},
		{
			name:     "contains empty string",
			receiver: &interpreter.String{Value: "hello"},
			search:   &interpreter.String{Value: ""},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := containsFn(tt.receiver, tt.search)

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

func TestStringStartsWith(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	startsWithFn := methods["startsWith"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		prefix   *interpreter.String
		expected bool
	}{
		{
			name:     "starts with prefix",
			receiver: &interpreter.String{Value: "hello world"},
			prefix:   &interpreter.String{Value: "hello"},
			expected: true,
		},
		{
			name:     "does not start with prefix",
			receiver: &interpreter.String{Value: "hello world"},
			prefix:   &interpreter.String{Value: "world"},
			expected: false,
		},
		{
			name:     "empty prefix",
			receiver: &interpreter.String{Value: "hello"},
			prefix:   &interpreter.String{Value: ""},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := startsWithFn(tt.receiver, tt.prefix)

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

func TestStringEndsWith(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	endsWithFn := methods["endsWith"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		suffix   *interpreter.String
		expected bool
	}{
		{
			name:     "ends with suffix",
			receiver: &interpreter.String{Value: "hello world"},
			suffix:   &interpreter.String{Value: "world"},
			expected: true,
		},
		{
			name:     "does not end with suffix",
			receiver: &interpreter.String{Value: "hello world"},
			suffix:   &interpreter.String{Value: "hello"},
			expected: false,
		},
		{
			name:     "empty suffix",
			receiver: &interpreter.String{Value: "hello"},
			suffix:   &interpreter.String{Value: ""},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := endsWithFn(tt.receiver, tt.suffix)

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

func TestStringReplace(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	replaceFn := methods["replace"]

	tests := []struct {
		name        string
		receiver    *interpreter.String
		old         *interpreter.String
		replacement *interpreter.String
		expected    string
	}{
		{
			name:        "replace substring",
			receiver:    &interpreter.String{Value: "hello world"},
			old:         &interpreter.String{Value: "world"},
			replacement: &interpreter.String{Value: "there"},
			expected:    "hello there",
		},
		{
			name:        "replace multiple occurrences",
			receiver:    &interpreter.String{Value: "foo bar foo"},
			old:         &interpreter.String{Value: "foo"},
			replacement: &interpreter.String{Value: "baz"},
			expected:    "baz bar baz",
		},
		{
			name:        "substring not found",
			receiver:    &interpreter.String{Value: "hello world"},
			old:         &interpreter.String{Value: "foo"},
			replacement: &interpreter.String{Value: "bar"},
			expected:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := replaceFn(tt.receiver, tt.old, tt.replacement)

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

func TestStringSubstring(t *testing.T) {
	t.Parallel()

	methods := StringMethods()
	substringFn := methods["substring"]

	tests := []struct {
		name     string
		receiver *interpreter.String
		start    *interpreter.Integer
		end      *interpreter.Integer
		expected string
		isError  bool
	}{
		{
			name:     "basic substring",
			receiver: &interpreter.String{Value: "hello world"},
			start:    &interpreter.Integer{Value: 0},
			end:      &interpreter.Integer{Value: 5},
			expected: "hello",
			isError:  false,
		},
		{
			name:     "middle substring",
			receiver: &interpreter.String{Value: "hello world"},
			start:    &interpreter.Integer{Value: 6},
			end:      &interpreter.Integer{Value: 11},
			expected: "world",
			isError:  false,
		},
		{
			name:     "entire string",
			receiver: &interpreter.String{Value: "hello"},
			start:    &interpreter.Integer{Value: 0},
			end:      &interpreter.Integer{Value: 5},
			expected: "hello",
			isError:  false,
		},
		{
			name:     "start out of bounds",
			receiver: &interpreter.String{Value: "hello"},
			start:    &interpreter.Integer{Value: -1},
			end:      &interpreter.Integer{Value: 5},
			isError:  true,
		},
		{
			name:     "end out of bounds",
			receiver: &interpreter.String{Value: "hello"},
			start:    &interpreter.Integer{Value: 0},
			end:      &interpreter.Integer{Value: 10},
			isError:  true,
		},
		{
			name:     "start greater than end",
			receiver: &interpreter.String{Value: "hello"},
			start:    &interpreter.Integer{Value: 3},
			end:      &interpreter.Integer{Value: 1},
			isError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := substringFn(tt.receiver, tt.start, tt.end)

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
