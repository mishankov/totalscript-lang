# TotalScript language specification

## Concept

TotalScript is a scripting language with batteries included. It includes built-in database and HTTP server for seamless scripting and creating internal and personal tools.

## Files
Source code for TotalScript is stored in files with `.tsl` extension.

## Types

### Primitive Types
| Type | Description | Min | Max | Example |
|------|-------------|-----|-----|---------|
| `integer` | 64-bit signed integer | -9,223,372,036,854,775,808 | 9,223,372,036,854,775,807 | `42`, `-17`, `0` |
| `float` | 64-bit floating point (IEEE 754) | ≈ -1.8 × 10³⁰⁸ | ≈ 1.8 × 10³⁰⁸ | `3.14`, `-0.5`, `1.0` |
| `string` | UTF-8 string | - | - | `"hello"`, `"world"` |
| `boolean` | True or false | - | - | `true`, `false` |

### Collection Types
| Type | Description | Example |
|------|-------------|---------|
| `array` | Dynamic ordered collection | `[1, 2, 3]`, `["a", "b"]` |
| `map` | Key-value pairs | `{"name": "Alice", "age": 30}` |

### Special Types
| Type | Description |
|------|-------------|
| `function` | First-class function |
| `model` | User-defined structured type |
| `result` | Success or error wrapper |

### Optional Types
Any type can be made nullable by adding `?` suffix:
```tsl
var name: string? = null      # Can be string or null
var age: integer = 42         # Cannot be null

name = "Alice"                # OK
name = null                   # OK
age = null                    # Error: integer cannot be null
```

### Union Types
Multiple types can be combined with `|`:
```tsl
var id: integer | string = 42
id = "abc-123"                # OK, string is allowed

var value: integer | float | string = 3.14
```

## Comments

```tsl
# Single line comment

var a = 3 # Comment after expression

###
Multiline comment
Second line
###
```

## Variables and Constants

```tsl
var a: integer = 3    # Explicit type and initialization
var b: integer        # Declaration without initialization
var c = "some string" # Implicit type inference

const A: integer = 3  # Same thing with constants
```

## Functions

Functions are first class objects in TotalScript.

```tsl
const add = function (a: float, b: float): float {
  return a + b
}

# Implicit return type inference
const multiply = function (a: float, b: float) {
  return a * b
}
```

## Models

Models are representations of complex types in TotalScript. Models are first class objects.

```tsl
const Point = model {
  x: float
  y: float
}

const Circle = model {
  center: Point
  radius: float
}

var myPoint = Point(2, 3)
```

### Model Methods

```tsl
const Rectangle = model {
  a: float
  b: float

  square = function() {
    return this.a * this.b
  }
}

var s = Rectangle(3, 4).square() # 12
```

## Operators

### Arithmetic Operators
| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `5 + 3` → `8` |
| `-` | Subtraction | `5 - 3` → `2` |
| `*` | Multiplication | `5 * 3` → `15` |
| `/` | Division | `5 / 2` → `2.5` |
| `//` | Integer division | `5 // 2` → `2` |
| `%` | Modulo | `5 % 3` → `2` |
| `**` | Power | `2 ** 3` → `8` |

### Comparison Operators
| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `5 == 5` → `true` |
| `!=` | Not equal | `5 != 3` → `true` |
| `<` | Less than | `3 < 5` → `true` |
| `>` | Greater than | `5 > 3` → `true` |
| `<=` | Less or equal | `3 <= 3` → `true` |
| `>=` | Greater or equal | `5 >= 3` → `true` |

### Logical Operators
| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `true && false` → `false` |
| `\|\|` | Logical OR | `true \|\| false` → `true` |
| `!` | Logical NOT | `!true` → `false` |

### Assignment Operators
| Operator | Description | Equivalent |
|----------|-------------|------------|
| `=` | Assignment | - |
| `+=` | Add and assign | `a = a + b` |
| `-=` | Subtract and assign | `a = a - b` |
| `*=` | Multiply and assign | `a = a * b` |
| `/=` | Divide and assign | `a = a / b` |
| `%=` | Modulo and assign | `a = a % b` |

### String Operators
| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Concatenation | `"hello" + " world"` → `"hello world"` |

## Control Flow

### Conditionals
```tsl
if x > 0 {
  print("positive")
} else if x < 0 {
  print("negative")
} else {
  print("zero")
}

# Ternary expression
var status = if age >= 18 { "adult" } else { "minor" }
```

### Switch-Case
```tsl
switch value {
  case 1 {
    print("one")
  }
  case 2, 3 {
    print("two or three")
  }
  case 4..10 {
    print("four to nine")
  }
  default {
    print("something else")
  }
}

# Switch as expression
var name = switch code {
  case 200 { "OK" }
  case 404 { "Not Found" }
  case 500 { "Server Error" }
  default { "Unknown" }
}
```

### Loops

#### For-in loop
```tsl
# Iterate over array (value only)
for item in [1, 2, 3] {
  print(item)
}

# Iterate over array with index
for index, item in [1, 2, 3] {
  print(index, item)     # 0 1, 1 2, 2 3
}

# Iterate over range
for i in 0..10 {        # 0 to 9 (exclusive end)
  print(i)
}

for i in 0..=10 {       # 0 to 10 (inclusive end)
  print(i)
}

# Iterate over map
for key, value in {"a": 1, "b": 2} {
  print(key, value)
}
```

#### C-style for loop
```tsl
for var i = 0; i < 10; i += 1 {
  print(i)
}
```

#### While loop
```tsl
var i = 0
while i < 10 {
  print(i)
  i += 1
}
```

#### Loop control
```tsl
for i in 0..100 {
  if i == 5 {
    continue        # Skip to next iteration
  }
  if i == 10 {
    break           # Exit loop
  }
  print(i)
}
```

## Error Handling

TotalScript uses union types with the built-in `Error` model for error handling.

### Error Model
`Error` is a built-in model for representing errors:
```tsl
# Built-in Error model (provided by TotalScript)
const Error = model {
  message: string
}
```

### Functions Returning Errors
Use union types to indicate a function can return either a value or an error:
```tsl
const divide = function(a: float, b: float): float | Error {
  if b == 0 {
    return Error("division by zero")
  }
  return a / b
}
```

### Handling Errors
Use type checking to handle the result:
```tsl
var result = divide(10, 2)

if result is Error {
  print("Error: " + result.message)
  return
}

# Here result is narrowed to float
print(result)
```

## Collections

### Arrays

#### Creating Arrays
```tsl
var numbers: array<integer> = [1, 2, 3, 4, 5]
var names = ["Alice", "Bob"]                      # inferred as array<string>
var mixed: array<integer | string> = [1, "two", 3]
var empty: array<integer> = []
```

#### Indexing and Slicing
```tsl
var arr = [10, 20, 30, 40, 50]

arr[0]          # 10 (first element)
arr[-1]         # 50 (last element)
arr[1..3]       # [20, 30] (slice, exclusive end)
arr[1..=3]      # [20, 30, 40] (slice, inclusive end)
arr[2..]        # [30, 40, 50] (from index to end)
arr[..3]        # [10, 20, 30] (from start to index)
```

#### Array Methods
```tsl
var arr = [1, 2, 3]

arr.length()              # 3
arr.push(4)               # arr is now [1, 2, 3, 4]
arr.pop()                 # returns 4, arr is now [1, 2, 3]
arr.insert(1, 10)         # arr is now [1, 10, 2, 3]
arr.remove(1)             # removes at index 1, arr is now [1, 2, 3]
arr.contains(2)           # true
arr.indexOf(2)            # 1

# Functional methods
arr.map(function(x) { return x * 2 })           # [2, 4, 6]
arr.filter(function(x) { return x > 1 })        # [2, 3]
arr.reduce(0, function(acc, x) { return acc + x })  # 6
arr.each(function(x) { print(x) })              # prints each element
```

### Maps

#### Creating Maps
```tsl
var user: map<string, integer | string> = {"name": "Alice", "age": 30}
var counts: map<string, integer> = {"a": 1, "b": 2}
var empty: map<string, integer> = {}
```

#### Accessing Values
```tsl
var user = {"name": "Alice", "age": 30}

user["name"]          # "Alice"
user["missing"]       # null (key doesn't exist)
```

#### Map Methods
```tsl
var m = {"a": 1, "b": 2}

m.length()            # 2
m.keys()              # ["a", "b"]
m.values()            # [1, 2]
m.contains("a")       # true
m.remove("a")         # removes key "a"

# Add or update
m["c"] = 3            # m is now {"a": 1, "b": 2, "c": 3}
```

## Built-in Functions

### Output
```tsl
print("Hello")              # Prints without newline
println("Hello")            # Prints with newline
println("Name:", name)      # Multiple arguments separated by space
```

### Type Conversions
Type conversion functions return a union type that includes `Error` for invalid conversions:
```tsl
# integer() returns integer | Error
integer(3.14)               # 3
integer("42")               # 42
integer("not a number")     # Error("cannot convert 'not a number' to integer")

# float() returns float | Error
float(42)                   # 42.0
float("3.14")               # 3.14
float("invalid")            # Error("cannot convert 'invalid' to float")

# string() always succeeds
string(42)                  # "42"
string(3.14)                # "3.14"
string(true)                # "true"

# boolean() returns boolean | Error
boolean(0)                  # false
boolean(1)                  # true
boolean("")                 # false
boolean("any")              # true
```

### Type Checking
```tsl
typeof(42)                  # "integer"
typeof("hello")             # "string"
typeof([1, 2, 3])           # "array"

value is integer            # true if value is integer
value is string             # true if value is string
value is Error              # true if value is Error model
```

### String Methods
```tsl
var s = "Hello, World!"

s.length()                  # 13
s.upper()                   # "HELLO, WORLD!"
s.lower()                   # "hello, world!"
s.trim()                    # removes whitespace from both ends
s.split(",")                # ["Hello", " World!"]
s.contains("World")         # true
s.startsWith("Hello")       # true
s.endsWith("!")             # true
s.replace("World", "TotalScript")  # "Hello, TotalScript!"
s.substring(0, 5)           # "Hello"
```

### Math Functions
```tsl
abs(-5)                     # 5
min(1, 2, 3)                # 1
max(1, 2, 3)                # 3
floor(3.7)                  # 3
ceil(3.2)                   # 4
round(3.5)                  # 4
sqrt(16)                    # 4.0
pow(2, 3)                   # 8.0
```

