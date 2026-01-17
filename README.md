# SLUG
**Scripting Language** *(that is)* **usualy great**

## Examples

### HTTP server + database

```tsl
import http
import db

var server = http.Server()

const AddOp = model {
    @id a: float
    @id b: float
    result: float
}

server.post("/add", function(req) {
    var data = req.json()

    var dbResults = db.find(AddOp) {
        this.a = data["a"]
        this.b = data["b"]
    }

    if dbResults.length() > 0 {
        return http.Response(
            200, 
            {"result": dbResults[0].result, "source": "db"}, 
            {"Content-Type": "application/json"}
        )
    }

    var result = data["a"] + data["b"]
    db.save(AddOp(data["a"], data["b"], result))

    return http.Response(
        200, 
        {"result": result, "source": "calculated"}, 
        {"Content-Type": "application/json"}
    )
})

server.start(8080)
```
