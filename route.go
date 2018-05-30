package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"regexp/syntax"
	"strings"
)

type Node struct {
	path    string
	capture *syntax.Regexp
	pattern *regexp.Regexp
	index   int
}

func (node *Node) Match(path string) bool {
	if node.pattern != nil {
		return node.pattern.MatchString(path)
	}

	return node.path == path
}

type Match struct {
	route      *Route
	parameters map[string]string
}

// Route represents URL pattern -> handler relation.
// Route implements Resolver interface.
type Route struct {
	nodes    []*Node
	captures []*Node
	handler  http.Handler
	name     string
}

// Name returns route name.
func (route *Route) Name() string {
	return route.name
}

// Reverse makes URL path using parameters as values for groups of regular expression.
func (route *Route) reverse(name string, parameters map[string]string) (path string, err error) {
	var parts = make([]string, 0, len(route.nodes))

	for _, node := range route.nodes {
		var part = node.path

		if node.capture != nil {
			if value, ok := parameters[node.capture.Name]; ok {
				part = value
			} else {
				return "", fmt.Errorf("No value for parameter '%s'", node.capture.Name)
			}
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, "/"), nil
}

// Resolve checks match URL path with pattern.
func (route *Route) resolve(parts []string) (*Match, bool) {
	if len(parts) != len(route.nodes) {
		return nil, false
	}

	for i, node := range route.nodes {
		if !node.Match(parts[i]) {
			return nil, false
		}
	}

	return &Match{route: route, parameters: route.GetGroups(parts)}, true
}

// GetGroups returns map of matched regular expression groups.
func (route *Route) GetGroups(parts []string) map[string]string {
	var groups = make(map[string]string, len(route.captures))

	for _, node := range route.captures {
		groups[node.capture.Name] = parts[node.index]
	}

	return groups
}

func (route *Route) tree() map[string]*Route {
	return map[string]*Route{route.name: route}
}

// NewRoute creates new Route instance.
func NewRoute(pattern string, handler http.Handler, name string) *Route {
	var ast, _ = syntax.Parse(pattern, syntax.Perl)
	var groups = make(map[string]*syntax.Regexp, len(ast.Sub)+1)

	pattern = strings.TrimPrefix(ast.String(), "/")

	for _, group := range append(ast.Sub, ast) {
		if group.Op == syntax.OpCapture {
			if group.Name == "" {
				panic("All groups in pattern should be named")
			}

			groups[group.Name] = group
			pattern = strings.Replace(pattern, group.String(), fmt.Sprintf(":%s", group.Name), -1)
		}
	}

	var parts = strings.Split(pattern, "/")
	var nodes = make([]*Node, 0, len(parts))
	var captures = make([]*Node, 0, len(groups))

	for i, part := range parts {
		var node = &Node{path: part, index: i}

		if strings.HasPrefix(part, ":") {
			if capture, ok := groups[strings.TrimPrefix(part, ":")]; ok {
				node.capture = capture
				node.pattern = regexp.MustCompile(capture.String())
				captures = append(captures, node)
			}
		}

		nodes = append(nodes, node)
	}

	return &Route{nodes, captures, handler, name}
}
