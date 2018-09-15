package actions

import (
	"os"
	"log"
	"encoding/json"
	"net/http"
	"bytes"
	"time"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func createEntries(client *http.Client, adminURL string, configMap map[string][]interface{}) {
	// In order to not overload the server, limit concurrent post requests to 10
	reqLimitChan := make(chan bool, 10)

	// Create services and routes in separate cycle as they depend on each other
	// and services should be created before routes
	for _, item := range configMap[ServicesPath] {
		reqLimitChan <- true

		// Convert item to service object for further creating it at Kong
		var service ServicePrepared
		mapstructure.Decode(item, &service)

		go createServiceWithRoutes(client, adminURL, service, reqLimitChan)
	}

	for _, resourceBundle := range ResourceBundles {
		url := getFullPath(adminURL, resourceBundle.Path)

		for _, item := range configMap[resourceBundle.Path] {
			reqLimitChan <- true
			mapstructure.Decode(item, &resourceBundle.Struct)
			go createResource(client, url, resourceBundle.Struct, reqLimitChan)
		}
	}

	//Be aware all left requests are finished
	for i := 0; i < cap(reqLimitChan); i++ {
		reqLimitChan <- true
	}
}

func createServiceWithRoutes(client *http.Client, url string, service ServicePrepared, reqLimitChan <-chan bool) {
	defer func() { <-reqLimitChan}()

	// Get path to the services collection
	servicesURL := getFullPath(url, ServicesPath)

	// Clear routes field as it is created in separate request
	routes := service.Routes
	service.Routes = nil

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(service)

	// Create services first, as routes are nested resources
	err := makePost(client, service, servicesURL)

	if err != nil {
		log.Fatalf("Failed to create service, %v\n", err)
		os.Exit(1)
	}

	// Compose path to routes
	routesPathElements := []string{ServicesPath, service.Name, RoutesPath}
	routesPath := strings.Join(routesPathElements, "/")
	routesURL := getFullPath(url, routesPath)

	// Create routes one by one
	for _, route := range routes {
		err := makePost(client, route, routesURL)

		if err != nil {
			log.Fatalf("Failed to create route, %v\n", err)
			os.Exit(1)
		}
	}

}

// Import - main function that is called by CLI in order to create resources at Kong service
func Import(adminURL string, filePath string) {
	client := &http.Client{Timeout: Timeout * time.Second}

	configFile, err := os.OpenFile(filePath, os.O_RDONLY,0444)

	if err != nil {
		log.Fatalf("Failed to read config file. %v\n", err.Error())
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(configFile)
	var configMap = make(map[string][]interface{})

	if err :=  jsonParser.Decode(&configMap); err != nil {
		log.Fatalf("Failed to parse json file. %v\n", err)
		os.Exit(1)
	}

	createEntries(client, adminURL, configMap)

	fmt.Println("Done")
}