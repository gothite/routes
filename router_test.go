package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterName(test *testing.T) {
	handler := func(response http.ResponseWriter, request *http.Request, options map[string]string) {}
	instance := NewRouter("/", "namespace", NewRoute("/", handler, "test"))

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
	handler := func(response http.ResponseWriter, request *http.Request, options map[string]string) {}
	instance := NewRouter(prefix, "test", NewRoute(pattern, handler, name))

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

	if _, matched = instance.Resolve("/prefix/wrong"); matched {
		test.Errorf("route matched by wrong path")
		return
	}
}

func TestRouterHandle(test *testing.T) {
	pattern := "(?P<path>path)"
	prefix := "prefix"
	path := strings.Join([]string{"", prefix, "path"}, "/")
	request, _ := http.NewRequest("GET", path, nil)

	handler := func(response http.ResponseWriter, request *http.Request, options map[string]string) {
		response.Write([]byte(request.URL.Path))
	}

	instance := NewRouter(prefix, "test", NewRoute(pattern, handler, "test"))

	mock := httptest.NewRecorder()

	http.HandlerFunc(instance.Handle).ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusOK {
		test.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}

	if mock.Body.String() != request.URL.Path {
		test.Errorf("handler returned unexpected body: got %v, want %v",
			mock.Body.String(), request.URL.Path)
		return
	}

	request, _ = http.NewRequest("GET", "wrong", nil)
	mock = httptest.NewRecorder()

	http.HandlerFunc(instance.Handle).ServeHTTP(mock, request)

	if status := mock.Code; status != http.StatusNotFound {
		test.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
		return
	}
}

func TestRouterReverse(test *testing.T) {
	pattern := "(?P<path>.*)"
	prefix := "prefix"
	name := "test"
	namespace := "root"
	handler := func(response http.ResponseWriter, request *http.Request, options map[string]string) {}
	instance := NewRouter(prefix, namespace, NewRoute(pattern, handler, name))

	path, found := instance.Reverse(name, map[string]string{"path": "test"})

	if !found {
		test.Error("route not reversed")
		return
	}

	if want := "/prefix/test"; path != want {
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
