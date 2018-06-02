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
	var v1 = routes.New()
	v1.Add("/:param", &handler{}, "endpoint")

	var api = routes.New()
	api.AddRouter("/v1", v1, "v1")

	var router = routes.New()
	router.AddRouter("/api", api, "api")

	path, _ := router.Reverse("api:v1:endpoint", "param")
	log.Print(path) // /api/v1/param

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
