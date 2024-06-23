package types

// APIHandler interface defines the method that all API handlers must implement
type APIHandler interface {
	HandleQuery(query string) string
}

// API represents an API with its handler and keyboard shortcut
type API struct {
	Name     string
	Shortcut string
	Handler  APIHandler
}
