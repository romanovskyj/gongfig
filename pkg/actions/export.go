package actions

import (
	"net/http"
	"net/url"
	"time"
	"log"
	"encoding/json"
	"io/ioutil"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type Service struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Routes []Route
}

type Route struct {
	Paths []string `json:"paths"`
	Service Service `json:"service"`
}

type Data []interface{}

// All items are contained of data property of json answer
type resourceConfig struct {
	Data Data `json:"data"`
}

// resourceAnswer contains resource name and its configuration so
// file writer can compose json with name as a key and complete resource configuration as a value
type resourceAnswer struct {
	resourceName string
	config Data
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


// Prepare config for writing: put routes as nested resources of services, omit unnecessary field etc
func composeConfig(config map[string]Data) map[string]interface{} {
	preparedConfig := make(map[string]interface{})
	serviceMap := make(map[string]*Service)

	// Create a map of services where key is service id in order to effectively
	// search services for pasting there corresponding routes
	for _, item := range config[ServicesKey] {
		var service Service
		mapstructure.Decode(item, &service)
		serviceMap[service.Id] = &service
	}

	// Add routes to services as nested files so futher it will be written to a file
	for _, item := range config[RoutesKey] {
		var route Route
		mapstructure.Decode(item, &route)
		serviceMap[route.Service.Id].Routes = append(serviceMap[route.Service.Id].Routes, route)
	}

	services := []Service{}

	// Rework serviceMap to a slice for writing it to the config file
	// as service entity already has an id field and it does not need to duplicate it
	for _, value := range serviceMap{
		services = append(services, *value)
	}

	preparedConfig[ServicesKey] = services

	return preparedConfig
}

func Export(adminUrl string, filePath string) {
	client := &http.Client{Timeout: 10 * time.Second}

	// We obtain resources data concurrently and push them to the channel that
	// will be handled by file writer
	writeData := make(chan *resourceAnswer)
	uri, _ := url.Parse(adminUrl)

	// Collect representation of all resources
	for _, resource := range Apis {
		uri.Path = resource
		fullPath := uri.String()

		go getResourceList(client, writeData, fullPath, resource)

	}

	resourcesNum := len(Apis)
	config := map[string]Data{}
	var preparedConfig map[string]interface{}

	// Before writing to a file the program composes json
	// It waits to obtain from channel exactly the same amount as number of resources
	// After that it composes the data in proper format, writes to a file and closes
	for {
		resource := <- writeData
		config[resource.resourceName] = resource.config

		resourcesNum--

		// resourcesNum is 0 means all needed resources are collected
		// and we can prepare config for writing it to a file
		if resourcesNum == 0 {
			preparedConfig = composeConfig(config)
			break
		}
	}

	jsonAnswer, _ := json.MarshalIndent(preparedConfig, "", "    ")
	ioutil.WriteFile(filePath, jsonAnswer, 0644)
	fmt.Println("Done")
}