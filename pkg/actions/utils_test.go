package actions

import (
	"strings"
)

func getRoutesURL() string {
	//Create path /services/<service name>/routes
	routesPathElements := []string{ServicesPath, TestEmailService.Name, RoutesPath}
	return strings.Join(routesPathElements, "/")
}

func getConsumerKeyAuthURL(consumerId string) string {
	//Create path /consumers/<consumer name>/key-auth
	routesPathElements := []string{ConsumersPath, consumerId, KeyAuthPath}
	return strings.Join(routesPathElements, "/")
}

func getResourcePath(path string) string {
	// The function that turns "/resource?size=10" into "resource"
	resourcePath := path[1:]

	questionMarkIndex := strings.IndexByte(resourcePath, '?')

	if questionMarkIndex > -1 {
		resourcePath = resourcePath[:questionMarkIndex]
	}

	return resourcePath
}
