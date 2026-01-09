package stdlib

import (
	"testing"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
)

func TestMapLength(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	lengthFn := methods["length"]

	tests := []struct {
		name     string
		receiver *interpreter.Map
		expected int64
	}{
		{
			name:     "empty map",
			receiver: &interpreter.Map{Pairs: make(map[string]interpreter.Object)},
			expected: 0,
		},
		{
			name: "map with elements",
			receiver: &interpreter.Map{
				Pairs: map[string]interpreter.Object{
					"a": &interpreter.Integer{Value: 1},
					"b": &interpreter.Integer{Value: 2},
					"c": &interpreter.Integer{Value: 3},
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

func TestMapKeys(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	keysFn := methods["keys"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
			"age":  &interpreter.Integer{Value: 30},
		},
	}

	result := keysFn(m)

	arrObj, ok := result.(*interpreter.Array)
	if !ok {
		t.Errorf("expected ARRAY, got %s", result.Type())
		return
	}

	if len(arrObj.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrObj.Elements))
		return
	}

	// Check that all keys are present (order doesn't matter)
	keySet := make(map[string]bool)
	for _, elem := range arrObj.Elements {
		strObj, ok := elem.(*interpreter.String)
		if !ok {
			t.Errorf("expected STRING key, got %s", elem.Type())
			continue
		}
		keySet[strObj.Value] = true
	}

	if !keySet["name"] || !keySet["age"] {
		t.Errorf("expected keys 'name' and 'age'")
	}
}

func TestMapValues(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	valuesFn := methods["values"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"a": &interpreter.Integer{Value: 1},
			"b": &interpreter.Integer{Value: 2},
		},
	}

	result := valuesFn(m)

	arrObj, ok := result.(*interpreter.Array)
	if !ok {
		t.Errorf("expected ARRAY, got %s", result.Type())
		return
	}

	if len(arrObj.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arrObj.Elements))
		return
	}

	// Check that all values are present (order doesn't matter)
	valueSet := make(map[int64]bool)
	for _, elem := range arrObj.Elements {
		intObj, ok := elem.(*interpreter.Integer)
		if !ok {
			t.Errorf("expected INTEGER value, got %s", elem.Type())
			continue
		}
		valueSet[intObj.Value] = true
	}

	if !valueSet[1] || !valueSet[2] {
		t.Errorf("expected values 1 and 2")
	}
}

func TestMapContains(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	containsFn := methods["contains"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
			"age":  &interpreter.Integer{Value: 30},
		},
	}

	tests := []struct {
		name     string
		key      *interpreter.String
		expected bool
	}{
		{
			name:     "contains existing key",
			key:      &interpreter.String{Value: "name"},
			expected: true,
		},
		{
			name:     "does not contain key",
			key:      &interpreter.String{Value: "email"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := containsFn(m, tt.key)

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

func TestMapRemove(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	removeFn := methods["remove"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
			"age":  &interpreter.Integer{Value: 30},
			"city": &interpreter.String{Value: "NYC"},
		},
	}

	result := removeFn(m, &interpreter.String{Value: "city"})

	if result != interpreter.NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if len(m.Pairs) != 2 {
		t.Errorf("expected 2 elements after remove, got %d", len(m.Pairs))
	}

	if _, exists := m.Pairs["city"]; exists {
		t.Errorf("key 'city' should have been removed")
	}

	if _, exists := m.Pairs["name"]; !exists {
		t.Errorf("key 'name' should still exist")
	}

	if _, exists := m.Pairs["age"]; !exists {
		t.Errorf("key 'age' should still exist")
	}
}

func TestMapRemoveNonExistentKey(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	removeFn := methods["remove"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
		},
	}

	result := removeFn(m, &interpreter.String{Value: "nonexistent"})

	if result != interpreter.NULL {
		t.Errorf("expected NULL, got %s", result.Type())
	}

	if len(m.Pairs) != 1 {
		t.Errorf("expected 1 element (unchanged), got %d", len(m.Pairs))
	}
}

func TestMapContainsNonStringKey(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	containsFn := methods["contains"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
		},
	}

	result := containsFn(m, &interpreter.Integer{Value: 123})

	if !interpreter.IsError(result) {
		t.Errorf("expected error when using non-string key, got %s", result.Type())
	}
}

func TestMapRemoveNonStringKey(t *testing.T) {
	t.Parallel()

	methods := MapMethods()
	removeFn := methods["remove"]

	m := &interpreter.Map{
		Pairs: map[string]interpreter.Object{
			"name": &interpreter.String{Value: "Alice"},
		},
	}

	result := removeFn(m, &interpreter.Integer{Value: 123})

	if !interpreter.IsError(result) {
		t.Errorf("expected error when using non-string key, got %s", result.Type())
	}
}
