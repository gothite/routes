package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Key is a type for context keys.
type Key string

// Router is a group of resolvers.
// Router implements Resolver and http.Handler interface.
type Router struct {
	prefix       string
	namespace    string
	defaultRoute *Route
	resolvers    map[string]Resolver
}

// Name returns router name (namespace).
func (router *Router) Name() string {
	return router.namespace
}

// ServeHTTP impelements http.Handler.ServeHTTP.
func (router *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	router.Handle(response, request)
}

// Add adds new resolver to router.
// It's may replace existing resolver with same name.
func (router *Router) Add(resolver Resolver) {
	router.resolvers[resolver.Name()] = resolver
}

// Reverse returns URL path from matched resolver.
func (router *Router) Reverse(name string, parameters map[string]string) (path string, err error) {
	return router.reverse(name, parameters)
}

func (router *Router) reverse(name string, parameters map[string]string) (path string, err error) {
	var parts = strings.Split(name, ":")

	if resolver, exists := router.resolvers[parts[0]]; exists {
		path, err := resolver.reverse(strings.Join(parts[1:], ":"), parameters)

		if err != nil {
			return "", err
		}

		path = fmt.Sprintf("/%v/%v", router.prefix, strings.TrimPrefix(path, "/"))
		return strings.Replace(path, "//", "/", 1), nil
	}

	return "", fmt.Errorf("Namespace '%s' not found", parts[0])
}

// Resolve looking route by path.
func (router *Router) resolve(parts []string) (*Match, bool) {
	if len(parts) < 2 || parts[0] != router.prefix {
		return nil, false
	}

	for _, route := range router.resolvers {
		if match, matched := route.resolve(parts[1:]); matched {
			return match, matched
		}
	}

	if router.defaultRoute != nil {
		return &Match{route: router.defaultRoute}, true
	}

	return nil, false
}

// Handle looking for route by path and delegates request to handler.
// If route not found, Handle will write header http.StatusNotFound.
func (router *Router) Handle(response http.ResponseWriter, request *http.Request) {
	var parts = strings.Split(request.URL.Path, "/")

	if match, found := router.resolve(parts[1:]); found {
		var ctx = request.Context()

		for key, value := range match.parameters {
			ctx = context.WithValue(ctx, Key(key), value)
		}

		match.route.handler.ServeHTTP(response, request.WithContext(ctx))
	} else {
		response.WriteHeader(http.StatusNotFound)
	}
}

// NewRouter creates new Router instance.
func NewRouter(prefix string, namespace string, defaultRoute *Route, resolvers ...Resolver) *Router {
	router := &Router{}
	router.prefix = strings.Trim(prefix, "/")
	router.namespace = namespace
	router.defaultRoute = defaultRoute
	router.resolvers = make(map[string]Resolver, len(resolvers))

	for _, resolver := range resolvers {
		router.Add(resolver)
	}

	return router
}
