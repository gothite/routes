package routes

import (
	"net/http"
	"strings"
	"testing"
)

func TestRouteResolve(test *testing.T) {
	pattern := "(?P<path>.*)"
	path := "/path"
	instance := NewRoute(pattern, func(response http.ResponseWriter, request *http.Request, options map[string]string) {}, "test")

	if want := strings.Join([]string{"/", pattern}, ""); instance.pattern.String() != want {
		test.Errorf("wrong prepared pattern: got %v, want %v", instance.pattern, want)
		return
	}

	route, matched := instance.Resolve(path)

	if !matched {
		test.Errorf("route not matched")
		return
	}

	if route != instance {
		test.Errorf("Resolve returned wrong route object: %v", route)
		return
	}

	if _, matched = instance.Resolve("wrong"); matched {
		test.Errorf("route matched by wrong path")
		return
	}
}

func TestRouteGetGroups(test *testing.T) {
	pattern := "(?P<path>.*)"
	path := "/path"
	route := NewRoute(pattern, func(response http.ResponseWriter, request *http.Request, options map[string]string) {}, "test")

	matches := route.GetGroups(path)

	if want := strings.TrimPrefix(path, "/"); matches["path"] != want {
		test.Errorf("route matched wrong: got %v, want %v", matches["path"], want)
		return
	}
}

func TestNewRoutePositionalGroups(test *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			test.Errorf("NewRoute didn't panic")
		}
	}()

	NewRoute("(.*)", func(response http.ResponseWriter, request *http.Request, options map[string]string) {}, "test")
}

func TestRouteReverse(test *testing.T) {
	pattern := "/prefix/(?P<path>.*)"
	name := "test"
	route := NewRoute(pattern, func(response http.ResponseWriter, request *http.Request, options map[string]string) {}, name)

	path, found := route.Reverse(name, map[string]string{"path": "test"})

	if !found {
		test.Error("route not reversed")
		return
	}

	if want := "/prefix/test"; path != want {
		test.Errorf("route reversed wrong: got %v, want %v", path, want)
		return
	}
}
