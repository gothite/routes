package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"regexp/syntax"
	"strings"
	"sync"
)

type Parameter struct {
	Name  string
	Value string
}

type Match struct {
	Path  string
	Route *Route

	parameters []Parameter
	sync       sync.Once
}

func (match *Match) parse() {
	var matches = match.Route.pattern.FindStringSubmatch(match.Path)[1:]
	var groups = match.Route.pattern.SubexpNames()[1:]

	match.parameters = make([]Parameter, len(groups))

	for i, name := range groups {
		match.parameters[i].Name = name
		match.parameters[i].Value = matches[i]
	}
}

func (match *Match) Get(name string) string {
	match.sync.Do(match.parse)

	for _, parameter := range match.parameters {
		if parameter.Name == name {
			return parameter.Value
		}
	}

	return ""
}

// Route represents URL pattern -> handler relation.
// Route implements Resolver interface.
type Route struct {
	pattern        *regexp.Regexp
	re             *syntax.Regexp
	handler        http.Handler
	defaultHandler http.Handler
	name           string
}

// Name returns route name.
func (route *Route) Name() string {
	return route.name
}

// Reverse makes URL path using parameters as values for groups of regular expression.
func (route *Route) reverse(parameters map[string]string) (path string, err error) {
	path = route.re.String()

	for _, sub := range route.re.Sub[1:] {
		if value, exists := parameters[sub.Name]; exists {
			path = strings.Replace(path, sub.String(), value, -1)
		} else {
			return "", fmt.Errorf("Have no value for '%s' parameter", sub.Name)
		}
	}

	return path, nil
}

// Resolve checks match URL path with pattern.
func (route *Route) Resolve(path string) (*Match, bool) {
	if route.pattern.MatchString(path) {
		return &Match{Path: path, Route: route}, true
	}

	return nil, false
}

func (route *Route) Routes() []*Route {
	return []*Route{route}
}

func (route *Route) Defaults() *defaults {
	return &defaults{}
}

// NewRoute creates new Route instance.
func NewRoute(pattern string, handler http.Handler, name string) *Route {
	pattern = fmt.Sprintf("/%s", strings.TrimPrefix(pattern, "/"))

	var re = regexp.MustCompile(pattern)
	var ast, _ = syntax.Parse(pattern, syntax.Perl)

	for _, name := range re.SubexpNames()[1:] {
		if len(name) == 0 {
			panic("All groups in pattern should be named")
		}
	}

	return &Route{re, ast, handler, nil, name}
}
