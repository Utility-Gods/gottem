package types

// APIInfo holds information about an API
type APIInfo struct {
	Name     string
	Shortcut string
	Handler  APIHandler
}

// App represents the main application structure
type App struct {
	APIs map[string]APIHandler
}

// APIHandler interface defines the method that all API handlers must implement
type APIHandler interface {
	HandleQuery(query string) string
}
