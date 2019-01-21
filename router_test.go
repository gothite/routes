package routes

import (
	"net/http"
	"reflect"
	"testing"
)

type Handler struct{}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

type Result struct {
	found      bool
	parameters []string
}

func (result *Result) Check(handler http.Handler, parameters []string) bool {
	return (handler != nil) == result.found && reflect.DeepEqual(result.parameters, parameters)
}

func TestRouterResolve(test *testing.T) {
	var router = New()
	var paths = []string{
		":user/activity/:activity",
		"api/v1/user/:user/",
		"api/v1/:user",
		"static/*path",
	}
	var tests = map[string]Result{
		"/1/activity/2":          Result{true, []string{"1", "2"}},
		"api/v1/user":            Result{false, nil},
		"/api/v1/user/1":         Result{true, []string{"1"}},
		"api/v2/user":            Result{false, nil},
		"/static/path/to/static": Result{true, []string{"path/to/static"}},
	}

	for _, path := range paths {
		router.Add(path, &Handler{}, "")
	}

	for path, result := range tests {
		if h, parameters := router.Resolve(path); !result.Check(h, parameters) {
			test.Errorf("Test '%s' failed!", path)
			test.Errorf("Handler: %v", h)
			test.Fatalf("Parameters: %v", parameters)
		}
	}
}

func TestRouterReverse(test *testing.T) {
	var api = New()
	api.Add("/:name/endpoint/:id", &Handler{}, "endpoint")

	var router = New()
	router.AddRouter("api", api, "api")

	if path, err := router.Reverse("api:endpoint", "n", "1"); err != nil {
		test.Fatal(err)
	} else if path != "/api/n/endpoint/1" {
		test.Error("Incorrect path!")
		test.Error("Expected: /api/n/endpoint/1")
		test.Fatalf("Got: %s", path)
	}
}

func TestRouterNotFoundHandler(test *testing.T) {
	var api = New()
	api.NotFoundHandler = &Handler{}

	var router = New()
	router.AddRouter("api", api, "api")

	if handler, _ := router.Resolve("/api/v1/path"); handler == nil {
		test.Fatal("Expected handler!")
	}
}

func TestOverrideRoutes(test *testing.T) {
	var v1 = New()
	var api = New()
	var root = New()

	v1.Add("/path/add", &Handler{}, "path.add")
	v1.Add("/path", &Handler{}, "path")
	v1.Add("/path/:id", &Handler{}, "path.id")
	v1.Add("/path/new", &Handler{}, "path.new")

	api.AddRouter("/v1", v1, "v1")
	root.AddRouter("/api", api, "api")

	if handler, _ := root.Resolve("/api/v1/path/add"); handler == nil {
		test.Fatal("/api/v1/path/add not found!")
	}

	if handler, _ := root.Resolve("/api/v1/path/new"); handler == nil {
		test.Fatal("/api/v1/path/new not found!")
	}

	if path, err := root.Reverse("api:v1:path.id", "1"); err != nil {
		test.Fatal(err)
	} else if path != "/api/v1/path/1" {
		test.Fatal(path)
	}
}

func BenchmarkRouterTwoParameters(benchmark *testing.B) {
	var router = New()

	router.Add(":user/activity/:activity", &Handler{}, "")

	for i := 0; i < benchmark.N; i++ {
		router.Resolve("/1/activity/2")
	}
}

func BenchmarkRouterGreedy(benchmark *testing.B) {
	var router = New()

	router.Add("static/*path", &Handler{}, "")

	for i := 0; i < benchmark.N; i++ {
		router.Resolve("/static/path/to/static")
	}
}

func BenchmarkRouterStatic(benchmark *testing.B) {
	var router = New()

	router.Add("/some/static/path", &Handler{}, "")
	router.Add("/some/static2/path", &Handler{}, "")
	router.Add("/some/static/path2", &Handler{}, "")

	for i := 0; i < benchmark.N; i++ {
		router.Resolve("/some/static/path")
	}
}

func BenchmarkRouterReverse(benchmark *testing.B) {
	var api = New()
	api.Add("/:name/endpoint/:id", &Handler{}, "endpoint")

	var router = New()
	router.AddRouter("api", api, "api")

	var parameters = []string{"n", "1"}

	for i := 0; i < benchmark.N; i++ {
		router.Reverse("api:endpoint", parameters...)
	}
}

func BenchmarkRouterServeHTTP(benchmark *testing.B) {
	var request, _ = http.NewRequest("GET", "/1/activity/2", nil)
	var router = New()

	router.Add(":user/activity/:activity", &Handler{}, "")

	for i := 0; i < benchmark.N; i++ {
		router.ServeHTTP(nil, request)
	}
}
