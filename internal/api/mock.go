package api

import (
	"fmt"
	"strings"
)

// MockAPI is a sample implementation of APIHandler
type MockAPI struct {
	Name string
}

func (m *MockAPI) HandleQuery(query string) string {
	return fmt.Sprintf("Response from %s: %s", m.Name, strings.ToUpper(query))
}
