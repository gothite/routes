package routes

// Resolver is a URL routing interface.
// It's common interface for Route and Router.
type Resolver interface {
	// Name returns name of resolver.
	// Name uses for resolver identification.
	Name() string

	// Resolve searches and returns route by passed URL path splitted by slash.
	Resolve(string) (*Match, bool)

	Routes() []*Route
	Defaults() *defaults
}
