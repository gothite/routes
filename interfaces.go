package routes

// Resolver is a URL routing interface.
// It's common interface for Route and Router.
type Resolver interface {
	// Name returns name of resolver.
	// Name uses for resolver identification.
	Name() string

	// Resolve searches and returns route by passed URL path splitted by slash.
	resolve([]string) (*Match, bool)

	// Reverse makes URL path by resolver name and URL parameters (regular expression groups).
	// Name may be nested like "route", "router:route", "api:v1:endpoint".
	reverse(string, map[string]string) (string, error)
}
