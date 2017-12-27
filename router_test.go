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
	pattern := "(?P<path>path)"
	prefix := "prefix"
	name := "test"
	path := strings.Join([]string{"", prefix, "path"}, "/")
	instance := NewRouter(prefix, "test", nil, NewRoute(pattern, Handler{"path"}, name))

	if want := strings.Join([]string{"/", prefix}, ""); instance.prefix != want {
		test.Errorf("wrong prepared pattern: got %v, want %v", instance.prefix, want)
		return
	}

	route, matched := instance.Resolve(path)

	if !matched {
		test.Errorf("route not matched")
		return
	}

	if route != instance.resolvers[name] {
		test.Errorf("Resolve returned wrong route object: %v", route)
		return
	}

	if _, matched = instance.Resolve("wrong"); matched {
		test.Errorf("route matched by wrong path")
		return
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

	path, found := instance.Reverse(name, map[string]string{"path": option})

	if !found {
		test.Error("route not reversed")
		return
	}

	if want := fmt.Sprintf("/%v/%v", prefix, option); path != want {
		test.Errorf("wrong prepared pattern: got %v, want %v", instance.prefix, want)
		return
	}

	_, found = instance.Reverse("wrong", map[string]string{})

	if found {
		test.Error("route reversed wrong")
		return
	}

	_, found = instance.Reverse(fmt.Sprintf("%v:%v", namespace, "wrong"), map[string]string{})

	if found {
		test.Error("route reversed wrong")
		return
	}
}
