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

func getRoutesUrl() string {
	//Create path /services/<service name>/routes
	routesPathElements := []string{ServicesPath, TestEmailService.Name, RoutesPath}
	return strings.Join(routesPathElements, "/")
}
