package routes

import (
	"net/http"
	"testing"
)

type Handler struct {
	option string
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var match = request.Context().Value(Key("match")).(*Match)

	response.Write([]byte(match.Get(handler.option)))
}

func TestRouteResolve(test *testing.T) {
	var pattern = `(?P<id>\d+)`
	var path = "/11"
	instance := NewRoute(pattern, &Handler{"path"}, "test")

	if match, matched := instance.Resolve(path); !matched {
		test.Fatalf("route not matched")
	} else if match.Route != instance {
		test.Errorf("Resolve returned wrong route object: %v", match.Route)
		return
	}

	if _, matched := instance.Resolve("wrong"); matched {
		test.Fatalf("route matched by wrong path")
	}
}

func TestNewRoutePositionalGroups(test *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			test.Errorf("NewRoute didn't panic")
		}
	}()

	NewRoute("(.*)", &Handler{"path"}, "test")
}

func TestRouteReverse(test *testing.T) {
	pattern := "/prefix/(?P<path>.*)"
	name := "test"
	route := NewRoute(pattern, &Handler{"path"}, name)

	if path, err := route.reverse(map[string]string{"path": "test"}); err != nil {
		test.Fatal(err)
	} else if want := "/prefix/test"; path != want {
		test.Fatalf("route reversed wrong: got %v, want %v", path, want)
	}
}
