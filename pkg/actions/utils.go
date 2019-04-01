package actions

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Data - general interface for storing json body answers
type Data []interface{}

// All items are contained of data property of json answer
type resourceConfig struct {
	Data Data `json:"data"`
}

// Get url, path items, query params and return concatenation
// e.g http://localhost:8001, services will return http://localhost:8001/services
func getFullPath(adminURL string, pathElements []string, params map[string]string) string {
	uri, _ := url.Parse(adminURL)
	path := strings.Join(pathElements, "/")

	if len(params) > 0 {
		q := uri.Query()
		for key, value := range params {
			q.Set(key, value)
		}

		uri.RawQuery = q.Encode()
	}

	uri.Path = path
	return uri.String()
}

func getResourceList(client *http.Client, fullPath string) resourceConfig {
	response, err := client.Get(fullPath)

	if err != nil {
		logFatal("Request to Kong admin failed")
		return resourceConfig{}
	}

	defer response.Body.Close()

	var body resourceConfig
	json.NewDecoder(response.Body).Decode(&body)

	return body
}

// Get list of resources by http and pass it to the channel where it will handled further
func getResourceListToChan(client *http.Client, writeData chan *resourceAnswer, fullPath string, resource string) {
	body := getResourceList(client, fullPath)

	// send only data field for writing in order to write { "service": [items...] } instead of
	// { "service": {"data": [items...] }}
	writeData <- &resourceAnswer{resource, body.Data}
}

func requestNewResource(client *http.Client, resource interface{}, url string) (string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(resource)

	// Create services first, as routes are nested resources
	response, err := client.Post(url, "application/json;charset=utf-8", body)

	if err != nil {
		logFatal("Request to Kong admin failed")
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 201 {
		message := Message{}
		json.NewDecoder(response.Body).Decode(&message)

		log.Println(message.Message)
		logFatal("Was not able to create resource")
		return "", err
	}

	createdResource := ResourceInstance{}

	json.NewDecoder(response.Body).Decode(&createdResource)

	return createdResource.Id, nil
}

func addResource(connectionBundle *ConnectionBundle, resource interface{}, resourceId string, idMap *ConcurrentStringMap) {
	defer func() { <-connectionBundle.ReqLimitChan}()

	externalId, err := requestNewResource(connectionBundle.Client, resource, connectionBundle.URL)

	if err != nil {
		logFatalf("Failed to create resource, %v\n", err)
	}

	idMap.Add(resourceId, externalId)
}

func isJSONString(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil

}
