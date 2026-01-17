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
| `enum` | Enumeration with named values |

### Built-in Models
| Model | Description |
|-------|-------------|
| `Error` | Error with `message: string` field |
| `Request` | HTTP request (see HTTP section) |
| `Response` | HTTP response (see HTTP section) |

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

  area = function() {
    return this.a * this.b
  }
}

var s = Rectangle(3, 4).area() # 12
```

### Multiple Constructors

Models can define multiple constructors with different signatures:

```tsl
const Point = model {
  x: float
  y: float

  # Default constructor is auto-generated: Point(x, y)

  # Additional constructor: create from single value
  constructor = function(value: float) {
    return Point(value, value)
  }

  # Additional constructor: create origin point
  constructor = function() {
    return Point(0, 0)
  }
}

var p1 = Point(3, 4)      # Default constructor
var p2 = Point(5)         # Single value, creates Point(5, 5)
var p3 = Point()          # Origin, creates Point(0, 0)
```

```tsl
const Color = model {
  r: integer
  g: integer
  b: integer

  # Constructor from hex string
  constructor = function(hex: string) {
    # Parse hex and create color
    return Color(255, 128, 0)
  }
}

var c1 = Color(255, 128, 0)   # RGB values
var c2 = Color("#FF8000")     # From hex string
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
  println("positive")
} else if x < 0 {
  println("negative")
} else {
  println("zero")
}

# Ternary expression
var status = if age >= 18 { "adult" } else { "minor" }
```

### Switch-Case
```tsl
switch value {
  case 1 {
    println("one")
  }
  case 2, 3 {
    println("two or three")
  }
  case 4..10 {
    println("four to nine")
  }
  default {
    println("something else")
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
  println(item)
}

# Iterate over array with index
for index, item in [1, 2, 3] {
  println(index, item)     # 0 1, 1 2, 2 3
}

# Iterate over range
for i in 0..10 {        # 0 to 9 (exclusive end)
  println(i)
}

for i in 0..=10 {       # 0 to 10 (inclusive end)
  println(i)
}

# Iterate over map
for key, value in {"a": 1, "b": 2} {
  println(key, value)
}
```

#### C-style for loop
```tsl
for var i = 0; i < 10; i += 1 {
  println(i)
}
```

#### While loop
```tsl
var i = 0
while i < 10 {
  println(i)
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
  println(i)
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
  println("Error: " + result.message)
  return
}

# Here result is narrowed to float
println(result)
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
arr.indexOf(2)            # 1 (returns integer | Error, Error if not found)
arr.join(", ")            # "1, 2, 3" (converts elements to strings and joins with separator)

# Functional methods
arr.map(function(x) { return x * 2 })           # [2, 4, 6]
arr.filter(function(x) { return x > 1 })        # [2, 3]
arr.reduce(0, function(acc, x) { return acc + x })  # 6
arr.each(function(x) { println(x) })              # prints each element
```

### Maps

#### Creating Maps
```tsl
var user: map<string, integer | string> = {"name": "Alice", "age": 30}
var counts: map<string, integer> = {"a": 1, "b": 2}
var empty: map<string, integer> = {}
```

#### Accessing Values
Accessing a map key returns an optional type (`V?`). Missing keys return `null`.
```tsl
var user = {"name": "Alice", "age": 30}

user["name"]          # "Alice"
user["missing"]       # null (key doesn't exist)

# Type: map<string, integer> access returns integer?
var counts = {"a": 1, "b": 2}
var val = counts["c"]   # val is integer?, value is null
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

# boolean() always succeeds (truthiness-based)
boolean(0)                  # false
boolean(1)                  # true (any non-zero number)
boolean("")                 # false
boolean("any")              # true (any non-empty string)
boolean(null)               # false
boolean(Point(0, 0))        # true (model instances are always truthy)
```

### Type Checking
```tsl
typeof(42)                  # "integer"
typeof("hello")             # "string"
typeof(true)                # "boolean"
typeof([1, 2, 3])           # "array"
typeof({"a": 1})            # "map"
typeof(Point(1, 2))         # "Point" (returns model name for model instances)

value is integer            # true if value is integer
value is string             # true if value is string
value is Point              # true if value is Point model instance
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

## Database

TotalScript includes built-in SQLite database support through the `db` module. Every model can be persisted to its own table.

```tsl
import db
```

The database file is created automatically (default: `data.db`) or can be specified via CLI argument `--db=myapp.db`. You can also configure it programmatically:

```tsl
db.configure("myapp.db")    # Set database file path
```

### Model Annotations

#### Primary Key (@id)
Mark a field as the primary key with `@id`. If no field is marked, an `_id` field with UUID is automatically added.

```tsl
const User = model {
  @id email: string     # email is the primary key
  name: string
  age: integer
}

const Point = model {
  x: float
  y: float
}
# Point automatically gets _id: string (UUID) as primary key
```

### Saving Data

```tsl
var user = User("alice@example.com", "Alice", 30)
db.save(user)

var point = Point(3.0, 4.0)
db.save(point)      # _id is auto-generated

# Save multiple instances
db.saveAll([point1, point2, point3])
```

### Querying Data

Use pattern matching syntax with `db.find()`. Use `this.` prefix for model fields to distinguish from variables:

```tsl
# Find all points where x > 5
var points = db.find(Point) {
  this.x > 5
}

# Multiple conditions (AND)
var points = db.find(Point) {
  this.x > 0
  this.y > 0
  this.x < 100
}

# OR conditions
var points = db.find(Point) {
  this.x > 100 || this.y > 100
}

# Comparison operators: ==, !=, <, >, <=, >=
var adults = db.find(User) {
  this.age >= 18
}

# String matching
var users = db.find(User) {
  this.name.startsWith("A")
  this.email.contains("@gmail.com")
}

# Using variables (no prefix)
var minAge = 18
var adults = db.find(User) {
  this.age >= minAge
}
```

### Query Modifiers

```tsl
# Order results
var sorted = db.find(Point) { this.x > 0 } orderBy x
var desc = db.find(Point) { this.x > 0 } orderBy x desc

# Limit results
var top10 = db.find(User) { this.age >= 18 } orderBy age limit 10

# Skip results (pagination)
var page2 = db.find(User) {} orderBy name limit 10 offset 10

# Get first match only (returns Model?, null if no match)
var first = db.find(Point) { this.x > 5 } first

# Count matches
var count = db.find(User) { this.age >= 18 } count
```

### Querying Nested Models

```tsl
const Circle = model {
  center: Point
  radius: float
}

# Query through nested model fields
var circles = db.find(Circle) {
  this.center.x > 0
  this.center.y > 0
  this.radius >= 5
}
```

### Updating Data

```tsl
# Find and modify
var user = db.find(User) { this.email == "alice@example.com" } first
user.age = 31
db.save(user)       # Updates existing record (same primary key)
```

### Deleting Data

```tsl
# Delete single instance
db.delete(user)

# Delete all instances of a model
db.deleteAll(User)
```

### Transactions

```tsl
db.transaction(function() {
  db.save(user1)
  db.save(user2)
  # All operations succeed or all fail
})
```

## HTTP

TotalScript provides built-in HTTP server and client through the `http` module.

```tsl
import http
```

The module exports:
- `http.Server()` - Server model constructor for creating HTTP server instances
- `http.client` - HTTP client object for making requests (module-level functions)
- `http.Request` - Request model type
- `http.Response()` - Response constructor function

### HTTP Server

The HTTP server is a model that you instantiate and configure:

```tsl
import http

var server = http.Server()
```

#### Defining Routes

```tsl
import http
import db

var server = http.Server()

server.get("/", function(req: http.Request): http.Response {
  return http.Response(200, "Hello, World!")
})

server.get("/users", function(req: http.Request): http.Response {
  var users = db.find(User) {}
  return http.Response(200, users)    # Auto-converts to JSON
})

server.post("/users", function(req: http.Request): http.Response {
  var data = req.json()
  if data is Error {
    return http.Response(400, {"error": data.message})
  }
  var user = User(data["email"], data["name"], data["age"])
  db.save(user)
  return http.Response(201, user)
})

server.put("/users/:id", function(req: http.Request): http.Response {
  var id = req.params["id"]
  # ...
  return http.Response(200, user)
})

server.delete("/users/:id", function(req: http.Request): http.Response {
  var id = req.params["id"]
  db.find(User) { this.email == id } delete
  return http.Response(204)
})
```

#### Starting the Server

```tsl
server.start(8080)              # Blocks and listens on port 8080
```

#### Request Object

```tsl
req.method                      # "GET", "POST", etc.
req.path                        # "/users/123"
req.params                      # Route params: map<string, string>
req.query                       # Query params: map<string, array<string>>
req.headers                     # Headers: map<string, array<string>>
req.body                        # Raw body as string
req.json()                      # Parse body as JSON, returns map | Error

# Accessing multi-value fields
req.query["tag"]                # ["a", "b"] for ?tag=a&tag=b
req.query["page"][0]            # "1" for ?page=1
req.headers["Accept"][0]        # "application/json"
```

#### Response Object

```tsl
http.Response(status)                          # Response with status only
http.Response(status, body)                    # Response with body (string or model)
http.Response(status, body, headers)           # Response with custom headers

# Examples
http.Response(200, "OK")
http.Response(200, {"message": "success"})     # Map auto-converts to JSON
http.Response(200, user)                       # Model auto-converts to JSON
http.Response(301, "", {"Location": ["/new-path"]})
```

### HTTP Client

#### Making Requests

All client methods return `http.Response | Error` (network errors return Error).

```tsl
import http

# GET request
var res = http.client.get("https://api.example.com/users")
if res is Error {
  println("Network error:", res.message)
  return
}
println(res.status)   # 200

# POST request with JSON body
var res = http.client.post("https://api.example.com/users", {
  "name": "Alice",
  "email": "alice@example.com"
})

# Other methods
var res = http.client.put(url, body)
var res = http.client.patch(url, body)
var res = http.client.delete(url)
```

#### Request with Headers

```tsl
var res = http.client.get("https://api.example.com/users", {
  "Authorization": ["Bearer token123"]
})

var res = http.client.post(url, body, {
  "Content-Type": ["application/json"],
  "Authorization": ["Bearer token123"]
})
```

#### Client Response

```tsl
res.status                      # Status code: 200, 404, etc.
res.body                        # Raw body as string
res.headers                     # Response headers: map<string, array<string>>
res.json()                      # Parse body as JSON, returns map | Error
res.ok                          # true if status is 2xx
```

### Static Files

```tsl
import http

var server = http.Server()
server.static("/assets", "./public")    # Serve ./public at /assets
server.static("/", "./dist")            # Serve ./dist at root
```

### Middleware

```tsl
import http

var server = http.Server()
server.use(function(req: http.Request, next: function): http.Response {
  println("Request:", req.method, req.path)
  var res = next(req)
  println("Response:", res.status)
  return res
})
```

## Enums

Enums define a type with a fixed set of named values. Each enum has an underlying simple type and all values must be explicitly specified.

### Defining Enums

```tsl
# Integer enum
const HttpStatus = enum {
  OK = 200
  NotFound = 404
  ServerError = 500
}

# String enum
const Direction = enum {
  North = "N"
  South = "S"
  East = "E"
  West = "W"
}

# Boolean enum
const Feature = enum {
  Enabled = true
  Disabled = false
}
```

### Using Enums

```tsl
var status = HttpStatus.OK
var dir = Direction.North

# Comparison
if status == HttpStatus.OK {
  println("Success")
}

# Get underlying value
println(status.value)       # 200
println(dir.value)          # "N"

# Switch on enum
switch status {
  case HttpStatus.OK {
    println("OK")
  }
  case HttpStatus.NotFound {
    println("Not found")
  }
  default {
    println("Other status")
  }
}
```

### Enum Methods

```tsl
Direction.values()              # [Direction.North, Direction.South, Direction.East, Direction.West]

# Get enum from value
var s = HttpStatus.fromValue(404)    # HttpStatus.NotFound | Error
var d = Direction.fromValue("N")     # Direction.North | Error
```

## Modules and Imports

Each `.tsl` file is a module. All top-level declarations (variables, constants, functions, models) are automatically exported.

### Importing Modules

```tsl
# Import standard library module
import math

# Import local file (relative path with ./)
import ./utils
import ./lib/helpers

# Import with alias
import math as m
import ./utils as u
```

### Accessing Imported Items

Imported items are accessed via the module name (qualified access):

```tsl
import math
import ./geometry as geo

var x = math.sin(3.14)
var y = math.cos(0)
var area = geo.circleArea(5.0)
```

### Module File Example

```tsl
# File: ./geometry.tsl

const PI = 3.14159

const circleArea = function(radius: float): float {
  return PI * radius ** 2
}

const Rectangle = model {
  width: float
  height: float

  area = function(): float {
    return this.width * this.height
  }
}
```

```tsl
# File: main.tsl

import ./geometry as geo

var area = geo.circleArea(10.0)
var rect = geo.Rectangle(5.0, 3.0)
println(rect.area())              # 15.0
println(geo.PI)                   # 3.14159
```

### Standard Library Modules

| Module | Description |
|--------|-------------|
| `math` | Mathematical functions and constants |
| `json` | JSON parsing and serialization |
| `time` | Date, time, and duration utilities |
| `fs` | File system operations |
| `os` | Operating system utilities, environment variables |

#### math module
```tsl
import math

# Constants
math.PI                     # 3.14159...
math.E                      # 2.71828...

# Basic functions
math.abs(-5)                # 5
math.min(1, 2, 3)           # 1
math.max(1, 2, 3)           # 3
math.floor(3.7)             # 3
math.ceil(3.2)              # 4
math.round(3.5)             # 4
math.sqrt(16)               # 4.0
math.pow(2, 3)              # 8.0

# Trigonometric functions
math.sin(math.PI / 2)       # 1.0
math.cos(0)                 # 1.0
math.tan(0)                 # 0.0

# Logarithmic functions
math.log(math.E)            # 1.0
math.log10(100)             # 2.0
```

#### json module
```tsl
import json

# Parse JSON string (returns map | Error)
var data = json.parse("{\"key\": \"value\"}")
if data is Error {
  println("Invalid JSON")
}

# Serialize to JSON string
var str = json.stringify({"name": "Alice", "age": 30})
```

#### fs module
```tsl
import fs

# Read file (returns string | Error)
var content = fs.readFile("./data.txt")
if content is Error {
  println("File not found")
}

# Write file (returns null | Error)
var err = fs.writeFile("./output.txt", "Hello")

# Check if file exists
fs.exists("./data.txt")     # true or false

# List directory
var files = fs.listDir("./")  # array<string> | Error
```

#### time module
```tsl
import time

var now = time.now()        # Current timestamp
time.sleep(1000)            # Sleep for 1000 milliseconds
```
