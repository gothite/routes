package routes

import (
	"net/http"
	"strings"
	"testing"
)

type Handler struct {
	option string
}

func (handler Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(request.Context().Value(Key(handler.option)).(string)))
}

func TestRouteResolve(test *testing.T) {
	var pattern = `/(?P<id>\d+)`
	var path = "11"
	instance := NewRoute(pattern, Handler{"path"}, "test")

	if match, matched := instance.resolve(strings.Split(path, "/")); !matched {
		test.Fatalf("route not matched")
	} else if match.route != instance {
		test.Errorf("Resolve returned wrong route object: %v", match.route)
		return
	}

	if _, matched := instance.resolve([]string{"wrong"}); matched {
		test.Fatalf("route matched by wrong path")
	}
}

func TestRouteGetGroups(test *testing.T) {
	pattern := "/(?P<path>.*)"
	path := "path"
	route := NewRoute(pattern, Handler{"path"}, "test")

	matches := route.GetGroups([]string{path})

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

	NewRoute("(.*)", Handler{"path"}, "test")
}

func TestRouteReverse(test *testing.T) {
	pattern := "/prefix/(?P<path>.*)"
	name := "test"
	route := NewRoute(pattern, Handler{"path"}, name)

	if path, err := route.reverse(name, map[string]string{"path": "test"}); err != nil {
		test.Fatal(err)
	} else if want := "prefix/test"; path != want {
		test.Fatalf("route reversed wrong: got %v, want %v", path, want)
	}
}
