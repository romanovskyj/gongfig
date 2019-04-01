package actions

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"os"
	"sync"
	"time"
)

// ConcurrentStringMap - special map for synchronizing localIds with externals
type ConcurrentStringMap struct {
	sync.Mutex
	store map[string]string
}

// Add - Locking is implemented in order to avoid problems with accessing to ConcurrentStringMap
func (concurrentStringMap *ConcurrentStringMap) Add(key, value string) {
	concurrentStringMap.Lock()
	defer concurrentStringMap.Unlock()

	concurrentStringMap.store[key] = value
}

func createEntries(client *http.Client, adminURL string, configMap map[string][]interface{}) {
	// In order to not overload the server, limit concurrent post requests to 10
	reqLimitChan := make(chan bool, 10)
	servicesConnectionBundle := ConnectionBundle{client, adminURL, reqLimitChan}

	// Map local resource ids with newly created
	concurrentStringMap := ConcurrentStringMap{store: make(map[string]string)}

	// Create services and routes in separate cycle as they depend on each other
	// and services should be created before routes
	for _, item := range configMap[ServicesPath] {
		reqLimitChan <- true

		// Convert item to service object for further creating it at Kong
		var service Service
		mapstructure.Decode(item, &service)

		go createServiceWithRoutes(&servicesConnectionBundle, service, &concurrentStringMap)
	}

	// Create upstreams and targets in separate cycle as they also depend on each other
	// (as services and routes)
	upstreamsConnectionBundle := ConnectionBundle{client, adminURL, reqLimitChan}

	for _, item := range configMap[UpstreamsPath] {
		reqLimitChan <- true

		var upstream Upstream
		mapstructure.Decode(item, &upstream)

		go createUpstreamsWithTargets(&upstreamsConnectionBundle, upstream)
	}


	url := getFullPath(adminURL, []string{CertificatesPath}, map[string]string{})

	for _, item := range configMap[CertificatesPath] {
		reqLimitChan <- true

		var certificate Certificate
		mapstructure.Decode(item, &certificate)

		go addResource(
			&ConnectionBundle{client, url, reqLimitChan},
			certificate, certificate.Id, &concurrentStringMap)
	}

	url = getFullPath(adminURL, []string{ConsumersPath}, map[string]string{})

	for _, item := range configMap[ConsumersPath] {
		reqLimitChan <- true

		var consumer Consumer
		mapstructure.Decode(item, &consumer)

		bundle := &ConnectionBundle{client, url, reqLimitChan}

		go createConsumersWithKeyAuths(bundle, consumer, &concurrentStringMap)
	}

	// Be aware all left requests are finished prior creation of depending resources
	for i := 0; i < cap(reqLimitChan); i++ {
		reqLimitChan <- true
	}

	// Clean channel for further creation
	for i := 0; i < cap(reqLimitChan); i++ {
		<- reqLimitChan
	}

	pluginsURL := getFullPath(adminURL, []string{PluginsPath}, map[string]string{})

	//Create plugins
	for _, item := range configMap[PluginsPath] {
		reqLimitChan <- true

		var plugin Plugin
		mapstructure.Decode(item, &plugin)

		if plugin.ServiceId != "" {
			plugin.ServiceId = concurrentStringMap.store[plugin.ServiceId]
		}

		if plugin.RouteId != "" {
			plugin.RouteId = concurrentStringMap.store[plugin.RouteId]
		}

		if plugin.ConsumerId != "" {
			plugin.ConsumerId = concurrentStringMap.store[plugin.ConsumerId]
		}

		go addResource(
			&ConnectionBundle{client, pluginsURL, reqLimitChan},
			&plugin, plugin.Id, &concurrentStringMap)
	}

	// Be aware all requests are finished prior to program exit
	for i := 0; i < cap(reqLimitChan); i++ {
		reqLimitChan <- true
	}
}

func createConsumersWithKeyAuths(requestBundle *ConnectionBundle, consumer Consumer, idMap *ConcurrentStringMap) {
	defer func() { <-requestBundle.ReqLimitChan}()

	//save id for adding it into idMap but avoid pushing when create consumer
	id := consumer.Id
	consumer.Id = ""

	// Store key in separate variable in order to not propagate it in consumer
	// resource itself
	key := consumer.Key
	consumer.Key = ""

	// Firstly create consumer in order to create keyauth at the next step for it
	consumerExternalId, err := requestNewResource(requestBundle.Client, consumer, requestBundle.URL)

	if err != nil {
		logFatalf("Failed to create consumer, %v\n", err)
		return
	}

	idMap.Add(id, consumerExternalId)

	if key != "" {
		paths := []string{ConsumersPath, consumerExternalId, KeyAuthPath}

		url := getFullPath(requestBundle.URL, paths, map[string]string{})
		keyAuth := KeyAuth{Key: key}

		_, err := requestNewResource(requestBundle.Client, keyAuth, url)

		if err != nil {
			logFatalf("Failed to create key-auth, %v\n", err)
			return
		}
	}
}

func createServiceWithRoutes(requestBundle *ConnectionBundle, service Service, idMap *ConcurrentStringMap) {
	defer func() { <-requestBundle.ReqLimitChan}()

	// Get path to the services collection
	servicesURL := getFullPath(requestBundle.URL, []string{ServicesPath}, map[string]string{})

	// Clear routes field as it is created in separate request
	routes := service.Routes
	service.Routes = nil

	// Record and clear id as it is for internal purposes
	id := service.Id
	service.Id = ""


	// Create services first, as routes are nested resources
	serviceExternalId, err := requestNewResource(requestBundle.Client, service, servicesURL)

	if err != nil {
		logFatalf("Failed to create service, %v\n", err)
		return
	}

	idMap.Add(id, serviceExternalId)

	// Compose path to routes
	routesPathElements := []string{ServicesPath, service.Name, RoutesPath}
	routesURL := getFullPath(requestBundle.URL, routesPathElements, map[string]string{})

	// Create routes one by one
	for _, route := range routes {
		// Record and clear id as it is for internal purposes
		id := route.Id
		route.Id = ""

		routeExternalId, err := requestNewResource(requestBundle.Client, route, routesURL)

		if err != nil {
			logFatalf("Could not create new resource, %v\n", err)
			return
		}

		idMap.Add(id, routeExternalId)
	}

}

func createUpstreamsWithTargets(requestBundle *ConnectionBundle, upstream Upstream) {
	defer func() { <-requestBundle.ReqLimitChan}()

	// Clear routes field as it is created in separate request
	targets := upstream.Targets
	upstream.Targets = nil

	// Clear id
	upstream.Id = ""

	upstreamsURL := getFullPath(requestBundle.URL, []string{UpstreamsPath}, map[string]string{})
	_, err := requestNewResource(requestBundle.Client, upstream, upstreamsURL)

	if err != nil {
		logFatalf("Could not create new resource, %v\n", err)
		return
	}

	paths := []string{UpstreamsPath, upstream.Name, TargetsPath}

	targetsURL := getFullPath(requestBundle.URL, paths, map[string]string{})

	for _, target := range targets {
		_, err := requestNewResource(requestBundle.Client, target, targetsURL)

		if err != nil {
			logFatalf("Failed to create target, %v\n", err)
			return
		}
	}

}

// Import - main function that is called by CLI in order to create resources at Kong service
func Import(adminURL string, filePath string) {
	client := &http.Client{Timeout: Timeout * time.Second}

	configFile, err := os.OpenFile(filePath, os.O_RDONLY,0444)

	if err != nil {
		logFatalf("Failed to read config file. %v\n", err.Error())
		return
	}

	jsonParser := json.NewDecoder(configFile)
	var configMap = make(map[string][]interface{})

	if err :=  jsonParser.Decode(&configMap); err != nil {
		logFatalf("Failed to parse json file. %v\n", err)
		return
	}

	createEntries(client, adminURL, configMap)

	fmt.Println("Done")
}