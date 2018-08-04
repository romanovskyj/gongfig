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


func createServiceWithRoutes(client *http.Client, url string, service ServicePrepared, reqLimitChan <-chan bool) {
	defer func() { <-reqLimitChan}()

	// Get path to the services collection
	servicesUrl := getFullPath(url, ServicesKey)

	// Clear routes field as it is created in separate request
	routes := service.Routes
	service.Routes = nil

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(service)

	// Create services first, as routes are nested resources
	response, err := client.Post(servicesUrl, "application/json;charset=utf-8", body)

	if err != nil {
		log.Fatal("Request to Kong admin failed")
		os.Exit(1)
	}

	if response.StatusCode != 201 {
		log.Fatal("Was not able to create service ", service.Name)
		os.Exit(1)
	}

	// Compose path to routes
	routesPathElements := []string{ServicesKey, service.Name, RoutesKey}
	routesPath := strings.Join(routesPathElements, "/")
	routesUrl := getFullPath(url, routesPath)

	// Create routes one by one
	for _, route := range routes {
		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(route)

		response, err = client.Post(routesUrl, "application/json;charset=utf-8", body)

		if response.StatusCode != 201 {
			log.Fatal("Was not able to create route for ", service.Name)
			os.Exit(1)
		}
	}

}

func Import(adminUrl string, filePath string) {
	client := &http.Client{Timeout: 10 * time.Second}

	configFile, err := os.OpenFile(filePath, os.O_RDONLY,0444)

	if err != nil {
		log.Fatal("Failed to read config file", err.Error())
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(configFile)
	var configMap = make(map[string][]interface{})

	if err :=  jsonParser.Decode(&configMap); err != nil {
		log.Fatal("Failed to parse json file")
		os.Exit(1)
	}

	// In order to not overload the server, limit concurrent post requests to 10
	reqLimitChan := make(chan bool, 10)

	// Current implementation imports services and nested routes only
	for _, item := range configMap[ServicesKey] {
		reqLimitChan <- true

		// Convert item to service object for further creating it at Kong
		var service ServicePrepared
		mapstructure.Decode(item, &service)

		go createServiceWithRoutes(client, adminUrl, service, reqLimitChan)
	}

	//Be aware all left requests are finished
	for i := 0; i < cap(reqLimitChan); i++ {
		reqLimitChan <- true
	}

	fmt.Println("Done")
}