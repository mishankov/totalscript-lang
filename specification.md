# TotalScript language specification

## Concept

TotalScript is scripting language with batteries included. It includes stuff like database and http server for seamless scripting and creating internal and personal tools

## Files
Source code for TotalScript stored in files with `.tsl` extension

## Simple types
### Numbers
- integer
- float

### Other
- string
- boolean

## Examples

### Comments

```tsl
# Single line comment

var a = 3 # Comment after expression

###
Multiline comment
Second line
###
```

### Variables and constants

```tsl
var a: integer = 3    # Explicit type and initialization
var b: integer        # Declaration without initialization
var c = "some string" # Implicit type inference

const A: integer = 3  # Same thing with constants 
```
### Functions

Functions are first class objects in TotalScript

```tsl
const add = function (a: float, b: float): float {
  return a + b
}

# Implicit return type inference
const multiply = function (a: float, b: float) {
  return a * b
}
```

### Models

Models are representation of complex types in TotalScript. Models are first class objects in TotalScript

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

#### Model methods

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

#### Saving and model instances from database

```tsl

```
