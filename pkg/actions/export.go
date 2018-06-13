package actions

import (
	"net/http"
	"net/url"
	"time"
	"log"
	"encoding/json"
	"io/ioutil"
	"fmt"
)

type Data interface {}

// All items are contained of data property of json answer
type resourceConfig struct {
	Data *Data `json:"data"`
}

// resourceAnswer contains resource name and its configuration so
// file writer can compose json with name as a key and complete resource configuration as a value
type resourceAnswer struct {
	resourceName string
	config *Data
}

// Get list of resources by http and pass it to the channel where it will be writed to a disk
func getResourceList(client *http.Client, writeData chan *resourceAnswer, fullPath string, resource string) {
	response, err := client.Get(fullPath)

	if err != nil {
		log.Fatal("Request to Kong admin failed")
		return
	}

	defer response.Body.Close()

	var body resourceConfig
	json.NewDecoder(response.Body).Decode(&body)

	// send only data field for writing in order to write { "service": [items...] } instead of
	// { "service": {"data": [items...] }}
	writeData <- &resourceAnswer{resource, body.Data}
}

func Export(adminUrl string, filePath string) {
	client := &http.Client{Timeout: 10 * time.Second}

	// We obtain resources data concurrently and push them to the channel that
	// will be handled by file writer
	writeData := make(chan *resourceAnswer)
	uri, _ := url.Parse(adminUrl)

	for _, resource := range Apis {
		uri.Path = resource
		fullPath := uri.String()

		go getResourceList(client, writeData, fullPath, resource)

	}

	resourcesNum := len(Apis)
	config := map[string]Data{}

	// Before writing to a file the program composes json
	// It waits to obtain from channel exactly the same amount as number of resources
	// After that it writes the data to a file and closes
	for {
		resource := <- writeData
		config[resource.resourceName] = resource.config

		resourcesNum--

		if resourcesNum == 0 {
			break
		}
	}

	jsonAnswer, _ := json.MarshalIndent(config, "", "    ")
	ioutil.WriteFile(filePath, jsonAnswer, 0644)
	fmt.Println("Done")
}