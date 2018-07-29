package actions

import (
	"fmt"
	"bufio"
	"os"
	"net/http"
	"time"
)

func flushResources(config map[string]Data) {
	// Firstly we need delete routes and only then services,
	// as routes are nested resources of services


}

func Flush(adminUrl string) {
	fmt.Println("All services and routes will be deleted from kong, are you sure? Write yes or no:")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')

	// Delete \n at the end
	answer = answer[0:len(answer)-1]

	if answer== "yes" {
		client := &http.Client{Timeout: 10 * time.Second}

		// We obtain resources data concurrently and push them to the channel that
		// will be handled by services and routes deleting logic
		flushData := make(chan *resourceAnswer)

		// Collect representation of all resources
		for _, resource := range Apis {
			fullPath := getFullPath(adminUrl, resource)

			go getResourceList(client, flushData, fullPath, resource)

		}

		resourcesNum := len(Apis)
		config := map[string]Data{}

		for {
			resource := <- flushData
			config[resource.resourceName] = resource.config

			resourcesNum--

			if resourcesNum == 0 {
				flushResources(config)
				break
			}
		}

	} else {
		fmt.Println("Configuration was not flushed")
	}
}
