package actions

import (
	"testing"
	"os"
	"net/http"
	"time"
	"os/exec"
	"strings"
	"net/http/httptest"
	"encoding/json"
)

// Run particular test in separate thread and test it exits with non zero value
func runExit(testName string) error {
	cmd := exec.Command(os.Args[0], strings.Join([]string{"-test.run=", testName}, ""))
	cmd.Env = append(os.Environ(), "CHECK_EXIT=1")
	err := cmd.Run()

	return err
}

// Create httpclient, service, chan and run CreateServiceWithRoutes with it
func prepareAndCreateService(url string){
	client := &http.Client{Timeout: 1 * time.Second}
	reqLimitChan := make(chan bool, 5)
	reqLimitChan <- true

	createServiceWithRoutes(client, url, TestService, reqLimitChan)
}

func TestCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		prepareAndCreateService(DefaultURL)
	}

	err := runExit("TestCannotConnect")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestBadRequest(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		prepareAndCreateService(ts.URL)
	}

	err := runExit("TestBadRequest")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestServiceWithRoutesCreated(t *testing.T) {
	//Create path /services/<service name>/routes
	routesPathElements := []string{ServicesKey, TestService.Name, RoutesKey}
	routesPath := strings.Join(routesPathElements, "/")

	serviceCreated := false
	routeCreated := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case ServicesKey:
			serviceCreated = true
			var body ServicePrepared
			json.NewDecoder(request.Body).Decode(&body)

			if body.Name != TestService.Name {
				t.Error("service name is not correct")
			}

		case routesPath:
			routeCreated = true
			var body RoutePrepared
			json.NewDecoder(request.Body).Decode(&body)

			if body.Paths[0] != TestService.Routes[0].Paths[0] {
				t.Error("route path is not correct")
			}
		}
		
	}))

	defer ts.Close()

	prepareAndCreateService(ts.URL)

	if !serviceCreated {
		t.Error("Service was not created")
	}

	if !routeCreated {
		t.Error("Route was not created")
	}
}