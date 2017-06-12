package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"regexp/syntax"
	"strings"
)

// Route represents URL pattern -> handler relation.
// Route implements Resolver interface.
type Route struct {
	pattern *regexp.Regexp
	re      *syntax.Regexp
	handler func(http.ResponseWriter, *http.Request, map[string]string)
	name    string
}

// Name returns route name.
func (route *Route) Name() string {
	return route.name
}

// Reverse makes URL path using parameters as values for groups of regular expression.
func (route *Route) Reverse(name string, parameters map[string]string) (path string, found bool) {
	path = route.re.String()

	for _, sub := range route.re.Sub {
		if value, exists := parameters[sub.Name]; exists {
			path = strings.Replace(path, sub.String(), value, -1)
		}
	}

	return path, true
}

// Resolve checks match URL path with pattern.
func (route *Route) Resolve(path string) (*Route, bool) {
	if route.pattern.MatchString(path) {
		return route, true
	}

	return nil, false
}

// GetGroups returns map of matched regular expression groups.
func (route *Route) GetGroups(path string) map[string]string {
	matches := route.pattern.FindStringSubmatch(path)[1:]
	matched := make(map[string]string)

	for i, name := range route.pattern.SubexpNames()[1:] {
		matched[name] = matches[i]
	}

	return matched
}

// NewRoute creates new Route instance.
func NewRoute(pattern string, handler func(http.ResponseWriter, *http.Request, map[string]string), name string) *Route {
	pattern = fmt.Sprintf("/%v", strings.TrimPrefix(pattern, "/"))
	re := regexp.MustCompile(pattern)

	for _, name := range re.SubexpNames()[1:] {
		if len(name) == 0 {
			panic("All groups in pattern should be named")
		}
	}

	ast, _ := syntax.Parse(re.String(), syntax.Perl)

	return &Route{re, ast, handler, name}
}
