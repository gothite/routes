# gothite/routes #
*Advanced URL routing for Go*

## Installation ##

`go get github.com/gothite/routes`


## Usage ##
```go
package main

import (
        "log"
        "net/http"

        "github.com/gothite/routes"
)

func handler(response http.ResponseWriter, request *http.Request) {
        path := request.Context().Value(routes.Key("path")).(string)
        response.Write([]byte(path))
}

func main() {
        router := routes.NewRouter(
                "/api", "api",
                routes.NewRouter(
                        "/v1", "v1",
                        routes.NewRoute("/(?P<path>.*)", handler, "endpoint"),
                ),
        )

        path, _ := router.Reverse("v1:endpoint", map[string]string{"path": "test"})
        log.Print(path) // /api/v1/test

        http.HandleFunc("/", router.Handle)
        log.Fatal(http.ListenAndServe(":8000", nil))
}
```
