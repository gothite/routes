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
	prefix    string
	namespace string
	defaults  *defaults
	routes    map[string]*Route
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
	for _, route := range resolver.Routes() {
		var path = fmt.Sprintf("%s%s", router.prefix, route.pattern.String())

		route = NewRoute(path, route.handler, route.name)

		router.routes[route.name] = route
	}
}

// Reverse returns URL path from matched resolver.
func (router *Router) Reverse(name string, parameters map[string]string) (path string, err error) {
	if route, exists := router.routes[name]; exists {
		if path, err := route.reverse(parameters); err != nil {
			return "", err
		} else {
			return path, nil
		}
	}

	return "", fmt.Errorf("Route named '%s' not found", name)
}

// Resolve looking route by path.
func (router *Router) Resolve(path string) (*Match, bool) {
	for _, route := range router.routes {
		if match, matched := route.Resolve(path); matched {
			return match, matched
		}
	}

	if route, found := router.defaults.get(path); found {
		return &Match{Path: path, Route: route}, true
	}

	return nil, false
}

// Handle looking for route by path and delegates request to handler.
// If route not found, Handle will write header http.StatusNotFound.
func (router *Router) Handle(response http.ResponseWriter, request *http.Request) {
	if match, found := router.Resolve(request.URL.Path); found {
		var ctx = context.WithValue(request.Context(), Key("match"), match)
		match.Route.handler.ServeHTTP(response, request.WithContext(ctx))
	} else {
		response.WriteHeader(http.StatusNotFound)
	}
}

func (router *Router) Routes() []*Route {
	var routes = make([]*Route, 0, len(router.routes))

	for _, route := range router.routes {
		route.name = fmt.Sprintf("%s:%s", router.namespace, route.name)
		routes = append(routes, route)
	}

	return routes
}

func (router *Router) Defaults() *defaults {
	return router.defaults
}

// NewRouter creates new Router instance.
func NewRouter(prefix string, namespace string, defaultRoute *Route, resolvers ...Resolver) *Router {
	router := &Router{}
	router.prefix = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))
	router.namespace = namespace
	router.routes = make(map[string]*Route)
	router.defaults = &defaults{
		prefixes: make([]string, 0, len(resolvers)),
		routes:   make(map[string]*Route, len(resolvers)),
	}

	if defaultRoute != nil {
		router.defaults.add(prefix, defaultRoute)
	}

	for _, resolver := range resolvers {
		router.Add(resolver)
		router.defaults.merge(prefix, resolver.Defaults())
	}

	return router
}
