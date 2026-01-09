package stdlib

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/ast"
	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

func TestArrayLength(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	lengthFn := methods["length"]

	tests := []struct {
		name     string
		receiver *interpreter.Array
		expected int64
	}{
		{
			name:     "empty array",
			receiver: &interpreter.Array{Elements: []interpreter.Object{}},
			expected: 0,
		},
		{
			name: "array with elements",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 2},
					&interpreter.Integer{Value: 3},
				},
			},
			expected: 3,
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

func TestArrayPush(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	pushFn := methods["push"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
		},
	}

	result := pushFn(arr, &interpreter.Integer{Value: 3})

	if result != interpreter.NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}

	lastElem, ok := arr.Elements[2].(*interpreter.Integer)
	if !ok || lastElem.Value != 3 {
		t.Errorf("expected last element to be 3")
	}
}

func TestArrayPop(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	popFn := methods["pop"]

	t.Run("pop from non-empty array", func(t *testing.T) {
		t.Parallel()

		arr := &interpreter.Array{
			Elements: []interpreter.Object{
				&interpreter.Integer{Value: 1},
				&interpreter.Integer{Value: 2},
				&interpreter.Integer{Value: 3},
			},
		}

		result := popFn(arr)

		intObj, ok := result.(*interpreter.Integer)
		if !ok {
			t.Errorf("expected INTEGER, got %s", result.Type())
			return
		}

		if intObj.Value != 3 {
			t.Errorf("expected 3, got %d", intObj.Value)
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements after pop, got %d", len(arr.Elements))
		}
	})

	t.Run("pop from empty array", func(t *testing.T) {
		t.Parallel()

		arr := &interpreter.Array{Elements: []interpreter.Object{}}

		result := popFn(arr)

		if !interpreter.IsError(result) {
			t.Errorf("expected error, got %s", result.Type())
		}
	})
}

func TestArrayInsert(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	insertFn := methods["insert"]

	tests := []struct {
		name     string
		receiver *interpreter.Array
		index    int64
		value    interpreter.Object
		expected []int64
		isError  bool
	}{
		{
			name: "insert at beginning",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 2},
					&interpreter.Integer{Value: 3},
				},
			},
			index:    0,
			value:    &interpreter.Integer{Value: 1},
			expected: []int64{1, 2, 3},
			isError:  false,
		},
		{
			name: "insert in middle",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 3},
				},
			},
			index:    1,
			value:    &interpreter.Integer{Value: 2},
			expected: []int64{1, 2, 3},
			isError:  false,
		},
		{
			name: "insert at end",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 2},
				},
			},
			index:    2,
			value:    &interpreter.Integer{Value: 3},
			expected: []int64{1, 2, 3},
			isError:  false,
		},
		{
			name: "insert out of bounds",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
				},
			},
			index:   10,
			value:   &interpreter.Integer{Value: 2},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := insertFn(tt.receiver, &interpreter.Integer{Value: tt.index}, tt.value)

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			if result != interpreter.NULL {
				t.Errorf("expected NULL, got %s", result.Type())
			}

			if len(tt.receiver.Elements) != len(tt.expected) {
				t.Errorf("expected %d elements, got %d", len(tt.expected), len(tt.receiver.Elements))
				return
			}

			for i, expected := range tt.expected {
				intObj, ok := tt.receiver.Elements[i].(*interpreter.Integer)
				if !ok {
					t.Errorf("element %d: expected INTEGER", i)
					continue
				}

				if intObj.Value != expected {
					t.Errorf("element %d: expected %d, got %d", i, expected, intObj.Value)
				}
			}
		})
	}
}

func TestArrayRemove(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	removeFn := methods["remove"]

	tests := []struct {
		name     string
		receiver *interpreter.Array
		index    int64
		expected []int64
		isError  bool
	}{
		{
			name: "remove from beginning",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 2},
					&interpreter.Integer{Value: 3},
				},
			},
			index:    0,
			expected: []int64{2, 3},
			isError:  false,
		},
		{
			name: "remove from middle",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 2},
					&interpreter.Integer{Value: 3},
				},
			},
			index:    1,
			expected: []int64{1, 3},
			isError:  false,
		},
		{
			name: "remove from end",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
					&interpreter.Integer{Value: 2},
					&interpreter.Integer{Value: 3},
				},
			},
			index:    2,
			expected: []int64{1, 2},
			isError:  false,
		},
		{
			name: "remove out of bounds",
			receiver: &interpreter.Array{
				Elements: []interpreter.Object{
					&interpreter.Integer{Value: 1},
				},
			},
			index:   10,
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := removeFn(tt.receiver, &interpreter.Integer{Value: tt.index})

			if tt.isError {
				if !interpreter.IsError(result) {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			if result != interpreter.NULL {
				t.Errorf("expected NULL, got %s", result.Type())
			}

			if len(tt.receiver.Elements) != len(tt.expected) {
				t.Errorf("expected %d elements, got %d", len(tt.expected), len(tt.receiver.Elements))
				return
			}

			for i, expected := range tt.expected {
				intObj, ok := tt.receiver.Elements[i].(*interpreter.Integer)
				if !ok {
					t.Errorf("element %d: expected INTEGER", i)
					continue
				}

				if intObj.Value != expected {
					t.Errorf("element %d: expected %d, got %d", i, expected, intObj.Value)
				}
			}
		})
	}
}

func TestArrayContains(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	containsFn := methods["contains"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.String{Value: "hello"},
			&interpreter.Boolean{Value: true},
		},
	}

	tests := []struct {
		name     string
		value    interpreter.Object
		expected bool
	}{
		{
			name:     "contains integer",
			value:    &interpreter.Integer{Value: 1},
			expected: true,
		},
		{
			name:     "contains string",
			value:    &interpreter.String{Value: "hello"},
			expected: true,
		},
		{
			name:     "does not contain",
			value:    &interpreter.Integer{Value: 99},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := containsFn(arr, tt.value)

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

func TestArrayIndexOf(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	indexOfFn := methods["indexOf"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
			&interpreter.Integer{Value: 3},
		},
	}

	tests := []struct {
		name     string
		value    interpreter.Object
		expected int64
		isError  bool
	}{
		{
			name:     "find at beginning",
			value:    &interpreter.Integer{Value: 1},
			expected: 0,
			isError:  false,
		},
		{
			name:     "find at middle",
			value:    &interpreter.Integer{Value: 2},
			expected: 1,
			isError:  false,
		},
		{
			name:     "find at end",
			value:    &interpreter.Integer{Value: 3},
			expected: 2,
			isError:  false,
		},
		{
			name:    "not found",
			value:   &interpreter.Integer{Value: 99},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := indexOfFn(arr, tt.value)

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

func TestArrayMap(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	mapFn := methods["map"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
			&interpreter.Integer{Value: 3},
		},
	}

	// Create a function that doubles the input
	fn := &interpreter.Function{
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "x"}},
		},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					ReturnValue: &ast.InfixExpression{
						Left:     &ast.Identifier{Value: "x"},
						Operator: "*",
						Right:    &ast.IntegerLiteral{Value: 2},
					},
				},
			},
		},
		Env: interpreter.NewEnvironment(),
	}

	result := mapFn(arr, fn)

	arrObj, ok := result.(*interpreter.Array)
	if !ok {
		t.Errorf("expected ARRAY, got %s", result.Type())
		return
	}

	expected := []int64{2, 4, 6}
	if len(arrObj.Elements) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(arrObj.Elements))
		return
	}

	for i, exp := range expected {
		intObj, ok := arrObj.Elements[i].(*interpreter.Integer)
		if !ok {
			t.Errorf("element %d: expected INTEGER", i)
			continue
		}

		if intObj.Value != exp {
			t.Errorf("element %d: expected %d, got %d", i, exp, intObj.Value)
		}
	}
}

func TestArrayFilter(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	filterFn := methods["filter"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
			&interpreter.Integer{Value: 3},
			&interpreter.Integer{Value: 4},
		},
	}

	// Create a function that returns true for even numbers
	fn := &interpreter.Function{
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "x"}},
		},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					ReturnValue: &ast.InfixExpression{
						Left: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: "x"},
							Operator: "%",
							Right:    &ast.IntegerLiteral{Value: 2},
						},
						Operator: "==",
						Right:    &ast.IntegerLiteral{Value: 0},
					},
				},
			},
		},
		Env: interpreter.NewEnvironment(),
	}

	result := filterFn(arr, fn)

	arrObj, ok := result.(*interpreter.Array)
	if !ok {
		t.Errorf("expected ARRAY, got %s", result.Type())
		return
	}

	expected := []int64{2, 4}
	if len(arrObj.Elements) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(arrObj.Elements))
		return
	}

	for i, exp := range expected {
		intObj, ok := arrObj.Elements[i].(*interpreter.Integer)
		if !ok {
			t.Errorf("element %d: expected INTEGER", i)
			continue
		}

		if intObj.Value != exp {
			t.Errorf("element %d: expected %d, got %d", i, exp, intObj.Value)
		}
	}
}

func TestArrayReduce(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	reduceFn := methods["reduce"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
			&interpreter.Integer{Value: 3},
		},
	}

	// Create a function that sums accumulator and current value
	fn := &interpreter.Function{
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "acc"}},
			{Name: &ast.Identifier{Value: "x"}},
		},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					ReturnValue: &ast.InfixExpression{
						Left:     &ast.Identifier{Value: "acc"},
						Operator: "+",
						Right:    &ast.Identifier{Value: "x"},
					},
				},
			},
		},
		Env: interpreter.NewEnvironment(),
	}

	result := reduceFn(arr, &interpreter.Integer{Value: 0}, fn)

	intObj, ok := result.(*interpreter.Integer)
	if !ok {
		t.Errorf("expected INTEGER, got %s", result.Type())
		return
	}

	if intObj.Value != 6 {
		t.Errorf("expected 6, got %d", intObj.Value)
	}
}

func TestArrayEach(t *testing.T) {
	t.Parallel()

	methods := ArrayMethods()
	eachFn := methods["each"]

	arr := &interpreter.Array{
		Elements: []interpreter.Object{
			&interpreter.Integer{Value: 1},
			&interpreter.Integer{Value: 2},
			&interpreter.Integer{Value: 3},
		},
	}

	// Create a simple function that just returns the value (for testing)
	fn := &interpreter.Function{
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "x"}},
		},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ReturnStatement{
					ReturnValue: &ast.Identifier{Value: "x"},
				},
			},
		},
		Env: interpreter.NewEnvironment(),
	}

	result := eachFn(arr, fn)

	if result != interpreter.NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}
}
