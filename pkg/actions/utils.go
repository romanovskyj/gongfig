package actions

import (
	"net/url"
	"net/http"
	"log"
	"encoding/json"
	"os"
	"bytes"
)

// Data - general interface for storing json body answers
type Data []interface{}

// All items are contained of data property of json answer
type resourceConfig struct {
	Data Data `json:"data"`
}

// Get url and path and return concatenation
// e.g http://localhost:8001, services will return http://localhost:8001/services
func getFullPath(adminURL string, path string) string {
	uri, _ := url.Parse(adminURL)
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

func makePost(client *http.Client, resource interface{}, url string) error {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(resource)

	// Create services first, as routes are nested resources
	response, err := client.Post(url, "application/json;charset=utf-8", body)

	if err != nil {
		log.Fatal("Request to Kong admin failed")
		return err
	}

	if response.StatusCode != 201 {
		log.Fatal("Was not able to create resource")
		return err
	}

	return nil
}

func createResource(client *http.Client, url string, resource interface{}, reqLimitChan <-chan bool) {
	defer func() { <-reqLimitChan}()

	err := makePost(client, resource, url)

	if err != nil {
		log.Fatalf("Failed to create resource, %v\n", err)
		os.Exit(1)
	}

}