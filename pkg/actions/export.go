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

// Service - for obtaining data from the server
// ServicePrepared - service object without Id field as id will be different
// for importing configuration every time so name is enough for identifying it
type Service struct {
	Id string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Path string `mapstructure:"path"`
	Port int `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
	ConnectTimeout int `mapstructure:"connect_timeout"`
	ReadTimeout int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	Routes []RoutePrepared
}

type ServicePrepared struct {
	Name string
	Host string
	Path string
	Port int
	Protocol string
	ConnectTimeout int
	ReadTimeout int
	WriteTimeout int
	Routes []RoutePrepared
}

// Route - for obtaining data from the server
// RoutePrepared - route object without Service field as route is already nested inside the server
type Route struct {
	Paths []string `mapstructure:"paths"`
	Service Service `mapstructure:"service"`
	StripPath bool `mapstructure:"strip_path"`
	PreserveHost bool `mapstructure:"preserve_host"`
	RegexPriority int `mapstructure:"regex_priority"`
	Hosts []string `mapstructure:"hosts"`
	Protocols []string `mapstructure:"protocols"`
	Methods []string `mapstructure:"methods"`
}

type RoutePrepared struct {
	Paths []string
	StripPath bool
	PreserveHost bool
	RegexPriority int
	Hosts []string
	Protocols []string
	Methods []string
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

		var routePrepared RoutePrepared
		mapstructure.Decode(item, &routePrepared)

		serviceMap[route.Service.Id].Routes = append(serviceMap[route.Service.Id].Routes, routePrepared)
	}

	services := []ServicePrepared{}

	// Rework serviceMap to a slice for writing it to the config file
	// as service entity already has an id field and it does not need to duplicate it
	for _, service := range serviceMap{
		servicePrepared := ServicePrepared{
			service.Name,
			service.Host,
			service.Path,
			service.Port,
			service.Protocol,
			service.ConnectTimeout,
			service.ReadTimeout,
			service.WriteTimeout,
			service.Routes,
		}

		services = append(services, servicePrepared)
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