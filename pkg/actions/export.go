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

type resourceConfig interface {}

// resourceAnswer contains resource name and its configuration so
// file writer can compose json with name as a key and complete resource configuration as a value
type resourceAnswer struct {
	resourceName string
	config *resourceConfig
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

		go func(fullPath string, resource string) {
			response, err := client.Get(uri.String())

			if err != nil {
				log.Fatal("Request to Kong admin failed")
				return
			}

			defer response.Body.Close()

			var body resourceConfig
			json.NewDecoder(response.Body).Decode(&body)

			writeData <- &resourceAnswer{resource, &body}
		}(fullPath, resource)

	}

	resourcesNum := len(Apis)
	config := map[string]resourceConfig{}

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

	fmt.Println(config)
	jsonAnswer, _ := json.MarshalIndent(config, "", "    ")
	ioutil.WriteFile(filePath, jsonAnswer, 0644)
}