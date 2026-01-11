- [ ] Imports without "s. `import "math"` -> `import math`

- [ ] Fix parse error for functions

This code parses successfuly:

```tsl
import "http"

var server = http.Server()

server.post("/add", function(r: Request): Response {
    var data = r.json()

    println(typeof(data), data)

    return http.Response(200, data["a"] + data["b"])
})

```

but this do not: 

```tsl
import "http"

var server = http.Server()

server.post("/add", function(r: http.Request): http.Response {
    var data = r.json()

    println(typeof(data), data)

    return http.Response(200, data["a"] + data["b"])
})
```

Looks like parser does not support types from other modules in function signatures

- [ ] Live reloading when code changes by default
