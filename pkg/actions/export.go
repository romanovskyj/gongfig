package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/getlantern/deepcopy.v1"
	"sort"
)

// resourceAnswer contains resource name and its configuration so
// file writer can compose json with name as a key and complete resource configuration as a value
type resourceAnswer struct {
	resourceName string
	config       Data
}

// Prepare config for writing: put routes as nested resources of services, omit unnecessary fields etc
func composeConfig(config map[string]Data, client *http.Client, url string) map[string]interface{} {
	preparedConfig := make(map[string]interface{})
	serviceMap := make(map[string]*Service)

	// Create a map of services where key is service id in order to effectively
	// search services for pasting there corresponding routes
	for _, item := range config[ServicesPath] {
		var service Service
		mapstructure.Decode(item, &service)
		serviceMap[service.Id] = &service
	}

	// Add routes to services as nested files so further it will be written to a file
	for _, item := range config[RoutesPath] {
		var route Route
		mapstructure.Decode(item, &route)

		var routePrepared Route
		mapstructure.Decode(item, &routePrepared)

		// Wipe service field as route located already inside of this service (nested)
		// so no need to duplicate it
		routePrepared.Service = nil

		serviceMap[route.Service.Id].Routes = append(serviceMap[route.Service.Id].Routes, routePrepared)
	}

	var services []Service

	// Rework serviceMap to a slice for writing it to the config file
	// as service entity already has an id field and it does not need to duplicate it
	for _, item := range serviceMap {
		service := Service{
			item.Id,
			item.Name,
			item.Host,
			item.Path,
			item.Port,
			item.Protocol,
			item.ConnectTimeout,
			item.ReadTimeout,
			item.WriteTimeout,
			item.Routes,
		}

		services = append(services, service)
	}

	//Sort services by name
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	preparedConfig[ServicesPath] = services

	// Obtain upstreams separately as it needs to do additional queries
	// for obtaining nested targets (there is no /targets collection so
	// only nested handling is possible)
	var upstreams []Upstream

	for _, item := range config[UpstreamsPath] {
		var upstream Upstream

		mapstructure.Decode(item, &upstream)

		// Compose path to particular target
		instancePathElements := []string{UpstreamsPath, upstream.Id, TargetsPath}
		upstreamTargetsURL := getFullPath(url, instancePathElements, map[string]string{"size": "500"})

		// Obtain targets
		var target Target
		targets := getResourceList(client, upstreamTargetsURL)

		for _, item := range targets.Data {
			mapstructure.Decode(item, &target)
			upstream.Targets = append(upstream.Targets, target)
		}

		upstreams = append(upstreams, upstream)
	}

	preparedConfig[UpstreamsPath] = upstreams

	// Handle Consumers separately as it needs to match it with key-auth if it exists
	consumerMap := make(map[string]*Consumer)

	// Create a map of consumers where key is consumer id in order to effectively
	// search consumers for pasting there corresponding api-keys
	for _, item := range config[ConsumersPath] {
		var consumer Consumer
		mapstructure.Decode(item, &consumer)
		consumerMap[consumer.Id] = &consumer
	}

	// Add api-key to consumers if it exists
	for _, item := range config[KeyAuthsPath] {
		var keyAuth KeyAuth
		mapstructure.Decode(item, &keyAuth)

		consumerMap[keyAuth.ConsumerId].Key = keyAuth.Key
	}

	var consumers []Consumer

	// Rework serviceMap to a slice for writing it to the config file
	// as consumer entity already has an id field and it does not need to duplicate it
	for _, item := range consumerMap {
		consumer := Consumer{
			item.Id,
			item.CustomId,
			item.Username,
			item.Key,
		}

		consumers = append(consumers, consumer)
	}

	preparedConfig[ConsumersPath] = consumers

	for _, resourceBundle := range ExportResourceBundles {
		var collection []interface{}
		for _, item := range config[resourceBundle.Path] {
			mapstructure.Decode(item, &resourceBundle.Struct)

			var resource interface{}
			deepcopy.Copy(&resource, &resourceBundle.Struct)

			collection = append(collection, resource)
		}

		preparedConfig[resourceBundle.Path] = collection
	}

	return preparedConfig
}

func getPreparedConfig(adminURL string) map[string]interface{} {
	client := &http.Client{Timeout: Timeout * time.Second}

	// We obtain resources data concurrently and push them to the channel that
	// will be handled by file writer
	writeData := make(chan *resourceAnswer)

	// Collect representation of all resources
	for _, resource := range Apis {
		//size means limit for number of elements that will be obtained within one request
		fullPath := getFullPath(adminURL, []string{resource}, map[string]string{"size": "500"})

		go getResourceListToChan(client, writeData, fullPath, resource)

	}

	resourcesNum := len(Apis)
	config := map[string]Data{}
	var preparedConfig map[string]interface{}

	// Before writing to a file the program composes json
	// It waits to obtain from channel exactly the same amount as number of resources
	// After that it composes the data in proper format, writes to a file and closes
	for {
		resource := <-writeData
		config[resource.resourceName] = resource.config

		resourcesNum--

		// resourcesNum is 0 means all needed resources are collected
		// and we can prepare config for writing it to a file
		if resourcesNum == 0 {
			preparedConfig = composeConfig(config, client, adminURL)
			break
		}
	}

	return preparedConfig
}

// Export - main function that is called by CLI in order to collect Kong config
func Export(adminURL string, filePath string) {
	preparedConfig := getPreparedConfig(adminURL)

	jsonAnswer, _ := json.MarshalIndent(preparedConfig, "", "    ")
	ioutil.WriteFile(filePath, jsonAnswer, 0644)
	fmt.Println("Done")
}
