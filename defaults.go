package routes

import (
	"fmt"
	"sort"
	"strings"
)

type byLength []string

func (s byLength) Len() int {
	return len(s)
}

func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type defaults struct {
	prefixes []string
	routes   map[string]*Route
}

func (defaults *defaults) add(prefix string, route *Route) {
	defaults.prefixes = append(defaults.prefixes, prefix)
	defaults.routes[prefix] = route

	sort.Sort(sort.Reverse(byLength(defaults.prefixes)))
}

func (defaults *defaults) get(path string) (*Route, bool) {
	for _, prefix := range defaults.prefixes {
		if strings.HasPrefix(path, prefix) {
			return defaults.routes[prefix], true
		}
	}

	return nil, false
}

func (defaults *defaults) merge(prefix string, other *defaults) {
	for _, otherPrefix := range other.prefixes {
		var newPrefix = fmt.Sprintf("%s%s", prefix, otherPrefix)
		defaults.routes[newPrefix] = other.routes[otherPrefix]
		defaults.prefixes = append(defaults.prefixes, newPrefix)
	}

	sort.Sort(sort.Reverse(byLength(defaults.prefixes)))
}
