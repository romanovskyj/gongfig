package actions

import (
	"os/exec"
	"os"
	"strings"
)

// Run particular test in separate thread and test it exits with non zero value
// such test implementation is needed as tested function does not return with error
// but simply stops the execution of the whole program (os.Exit)
func runExit(testName string) error {
	cmd := exec.Command(os.Args[0], strings.Join([]string{"-test.run=", testName}, ""))
	cmd.Env = append(os.Environ(), "CHECK_EXIT=1")
	err := cmd.Run()

	return err
}

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
