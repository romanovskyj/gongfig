package actions

import "net/url"

// Get url and path and return concatenation
// e.g http://localhost:8001, services will return http://localhost:8001/services
func getFullPath(adminUrl string, path string) string {
	uri, _ := url.Parse(adminUrl)
	uri.Path = path
	return uri.String()
}