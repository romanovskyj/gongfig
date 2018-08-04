package actions

import (
	"net/url"
	"net/http"
	"log"
	"encoding/json"
	"os"
)

// Get url and path and return concatenation
// e.g http://localhost:8001, services will return http://localhost:8001/services
func getFullPath(adminUrl string, path string) string {
	uri, _ := url.Parse(adminUrl)
	uri.Path = path
	return uri.String()
}

// Get list of resources by http and pass it to the channel where it will handled further
func getResourceList(client *http.Client, writeData chan *resourceAnswer, fullPath string, resource string) {
	response, err := client.Get(fullPath)

	if err != nil {
		log.Fatal("Request to Kong admin failed")
		os.Exit(1)
	}

	defer response.Body.Close()

	var body resourceConfig
	json.NewDecoder(response.Body).Decode(&body)

	// send only data field for writing in order to write { "service": [items...] } instead of
	// { "service": {"data": [items...] }}
	writeData <- &resourceAnswer{resource, body.Data}
}