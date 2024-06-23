package api

import "fmt"

// WeatherAPI is a sample implementation of APIHandler for weather
type OpenAiAPI struct{}

func (w *OpenAiAPI) HandleQuery(query string) string {
	// In a real app, you'd make an API call here
	return fmt.Sprintf("Weather for %s: Sunny and 72°F", query)
}
