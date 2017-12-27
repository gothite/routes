package main

import (
	"log"
	"net/http"

	"github.com/gothite/routes"
)

type handler struct{}

func (h handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	path := request.Context().Value(routes.Key("path")).(string)
	response.Write([]byte(path))
}

func main() {
	router := routes.NewRouter(
		"/api", "api",
		nil, // default route
		routes.NewRouter(
			"/v1", "v1",
			nil, // default route
			routes.NewRoute("/(?P<path>.*)", handler{}, "endpoint"),
		),
	)

	path, _ := router.Reverse("v1:endpoint", map[string]string{"path": "test"})
	log.Print(path) // /api/v1/test

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
