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

// Create httpclient, service, chan and run CreateServiceWithRoutes with it
func prepareAndCreateService(url string){
	client := &http.Client{Timeout: 1 * time.Second}
	reqLimitChan := make(chan bool, 5)
	reqLimitChan <- true

	createServiceWithRoutes(client, url, TestEmailService, reqLimitChan)
}

func TestImportCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		prepareAndCreateService(DefaultURL)
	}

	err := runExit("TestImportCannotConnect")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestImportBadRequest(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		prepareAndCreateService(ts.URL)
	}

	err := runExit("TestImportBadRequest")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestServiceWithRoutesCreated(t *testing.T) {
	//Create path /services/<service name>/routes
	routesPathElements := []string{ServicesKey, TestEmailService.Name, RoutesKey}
	routesPath := strings.Join(routesPathElements, "/")

	serviceCreated := false
	routeCreated := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case ServicesKey:
			var body ServicePrepared
			json.NewDecoder(request.Body).Decode(&body)

			if body.Name != TestEmailService.Name {
				t.Error("service name is not correct")
			}

			serviceCreated = true

		case routesPath:
			var body RoutePrepared
			json.NewDecoder(request.Body).Decode(&body)

			if body.Paths[0] != TestEmailService.Routes[0].Paths[0] {
				t.Error("route path is not correct")
			}

			routeCreated = true
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

func TestServiceCreatedRoutesFailed(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		routesPathElements := []string{ServicesKey, TestEmailService.Name, RoutesKey}
		routesPath := strings.Join(routesPathElements, "/")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			// Use path without slash ([1:])
			switch path := request.URL.Path[1:]; path {
			case ServicesKey:
				var body ServicePrepared
				json.NewDecoder(request.Body).Decode(&body)

				if body.Name != TestEmailService.Name {
					t.Error("service name is not correct")
				}

				w.WriteHeader(http.StatusCreated)

			case routesPath:
				w.WriteHeader(http.StatusBadRequest)
			}

		}))

		defer ts.Close()

		prepareAndCreateService(ts.URL)
	}

	err := runExit("TestServiceCreatedRoutesFailed")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}
