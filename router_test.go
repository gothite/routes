package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CodeHandler struct {
	code int
}

func (handler *CodeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(handler.code)
}

func TestRouterName(test *testing.T) {
	instance := NewRouter("/", "namespace", nil, NewRoute("/", Handler{"path"}, "test"))

	if want := instance.Name(); instance.namespace != want {
		test.Errorf("wrong name: got %v, want %v", instance.namespace, want)
		return
	}
}

func TestRouterResolve(test *testing.T) {
	var pattern = "(?P<path>path)"
	var prefix = "prefix"
	var name = "test"
	var instance = NewRouter(prefix, "test", nil, NewRoute(pattern, Handler{"path"}, name))

	if match, matched := instance.resolve([]string{prefix, "path"}); !matched {
		test.Fatalf("route not matched")
	} else if match.route != instance.resolvers[name] {
		test.Fatalf("Resolve returned wrong route object: %v", match.route)
	}

	if _, matched := instance.resolve([]string{"wrong"}); matched {
		test.Fatalf("route matched by wrong path")
	}
}

func TestRouterHandle(test *testing.T) {
	pattern := "(?P<path>path)"
	prefix := "prefix"
	option := "path"
	path := strings.Join([]string{"", prefix, option}, "/")
	request, _ := http.NewRequest("GET", path, nil)

	instance := NewRouter(prefix, "test", nil, NewRoute(pattern, Handler{"path"}, "test"))

	mock := httptest.NewRecorder()

	instance.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusOK {
		test.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}

	if mock.Body.String() != option {
		test.Errorf("handler returned unexpected body: got %v, want %v",
			mock.Body.String(), option)
		return
	}

	request, _ = http.NewRequest("GET", "wrong", nil)
	mock = httptest.NewRecorder()

	instance.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusNotFound {
		test.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}
}

func TestRouterHandleDefaultRoute(test *testing.T) {
	instance := NewRouter(
		"/api", "api",
		NewRoute("", &CodeHandler{http.StatusNotImplemented}, "default"),
		NewRouter("/v1", "v1", nil),
		NewRouter(
			"/v2", "v2",
			NewRoute("", &CodeHandler{http.StatusNotAcceptable}, "default"),
		),
	)

	request, _ := http.NewRequest("GET", "/api/", nil)
	mock := httptest.NewRecorder()

	instance.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusNotImplemented {
		test.Errorf("Incorrect status code!\n")
		test.Errorf("Expected: %d", http.StatusNotImplemented)
		test.Errorf("Actual: %d", status)
		return
	}

	request, _ = http.NewRequest("GET", "/api/v1/", nil)
	mock = httptest.NewRecorder()

	instance.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusNotImplemented {
		test.Errorf("Incorrect status code!\n")
		test.Errorf("Expected: %d", http.StatusNotImplemented)
		test.Errorf("Actual: %d", status)
		return
	}

	request, _ = http.NewRequest("GET", "/api/v2/", nil)
	mock = httptest.NewRecorder()

	instance.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusNotAcceptable {
		test.Errorf("Incorrect status code!\n")
		test.Errorf("Expected: %d", http.StatusNotAcceptable)
		test.Errorf("Actual: %d", status)
		return
	}
}

func TestRouterReverse(test *testing.T) {
	pattern := "(?P<path>.*)"
	prefix := "prefix"
	name := "test"
	namespace := "root"
	option := "path"

	instance := NewRouter(prefix, namespace, nil, NewRoute(pattern, Handler{"path"}, name))

	if path, err := instance.Reverse(name, map[string]string{"path": option}); err != nil {
		test.Fatal(err)
	} else if want := fmt.Sprintf("/%v/%v", prefix, option); path != want {
		test.Errorf("wrong prepared pattern: got %v, want %v", instance.prefix, want)
		return
	}

	if _, err := instance.Reverse("wrong", map[string]string{}); err == nil {
		test.Fatal("route reversed wrong")
	}

	if _, err := instance.Reverse(fmt.Sprintf("%v:%v", namespace, "wrong"), map[string]string{}); err == nil {
		test.Fatal("route reversed wrong")
	}
}

func TestRouterNamespaces(test *testing.T) {
	var request, _ = http.NewRequest("GET", "/api/v1/user/4545786125", nil)
	var mock = httptest.NewRecorder()
	var router = NewRouter(
		"/api", "api",
		nil, // default route
		NewRouter(
			"/v1", "v1",
			nil, // default route
			NewRoute(`(?P<path>\w+)/(?P<id>\d+)`, Handler{"id"}, "endpoint"),
		),
	)

	router.ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusOK {
		test.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}

	if mock.Body.String() != "4545786125" {
		test.Errorf("handler returned unexpected body: got %v, want %v",
			mock.Body.String(), "4545786125")
		return
	}
}

func BenchmarkRouter(benchmark *testing.B) {
	var request, _ = http.NewRequest("GET", "/api/v1/user/4545786125", nil)
	var mock = httptest.NewRecorder()
	var router = NewRouter(
		"/api", "api",
		nil, // default route
		NewRouter(
			"/v1", "v1",
			nil, // default route
			NewRoute(`(?P<path>\w+)/(?P<id>\d+)`, Handler{"id"}, "endpoint"),
		),
	)

	for i := 0; i < benchmark.N; i++ {
		router.ServeHTTP(mock, request)
	}
}
