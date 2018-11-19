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
	"io"
	"fmt"
)

func getHTTPRequestBundle(url string) *ConnectionBundle {
	client := &http.Client{Timeout: 1 * time.Second}
	reqLimitChan := make(chan bool, 5)

	return &ConnectionBundle{client, url,reqLimitChan}
}

// Create httpclient, service, chan and run CreateServiceWithRoutes with it
func prepareAndCreateService(url string, concurrentStringMap *ConcurrentStringMap){
	connectionBundle := getHTTPRequestBundle(url)
	connectionBundle.ReqLimitChan <- true

	createServiceWithRoutes(connectionBundle, TestEmailService, concurrentStringMap)
}

func TestImportCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		prepareAndCreateService(DefaultURL, &ConcurrentStringMap{store: make(map[string]string)})
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

		prepareAndCreateService(ts.URL, &ConcurrentStringMap{store: make(map[string]string)})
	}

	err := runExit("TestImportBadRequest")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestServiceWithRoutesCreated(t *testing.T) {
	routesPath := getRoutesURL()

	serviceCreated := false
	routeCreated := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case ServicesPath:
			var body Service
			json.NewDecoder(request.Body).Decode(&body)

			if body.Name != TestEmailService.Name {
				t.Error("service name is not correct")
			}

			serviceCreated = true

		case routesPath:
			var body Route
			json.NewDecoder(request.Body).Decode(&body)

			if body.Paths[0] != TestEmailService.Routes[0].Paths[0] {
				t.Error("route path is not correct")
			}

			routeCreated = true
		}
		
	}))

	defer ts.Close()

	prepareAndCreateService(ts.URL, &ConcurrentStringMap{store: make(map[string]string)})

	if !serviceCreated {
		t.Error("Service was not created")
	}

	if !routeCreated {
		t.Error("Route was not created")
	}
}

func TestCertificatesCreated(t *testing.T) {
	certificatesCreated := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case CertificatesPath:
			var body Certificate
			json.NewDecoder(request.Body).Decode(&body)

			if body.Cert != TestCertificate.Cert {
				t.Error("Certificate name is not correct")
			}

			certificatesCreated = true
		}

	}))
	defer ts.Close()

	connectionBundle := getHTTPRequestBundle(ts.URL)
	config := make(map[string][]interface{})

	config[CertificatesPath] = []interface{}{
		map[string]string{"cert": TestCertificate.Cert},
	}

	createEntries(connectionBundle.Client, ts.URL, config)

	if !certificatesCreated {
		t.Error("Certificate was not created")
	}
}

func TestPluginCreated(t *testing.T) {
	pluginCreated := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case PluginsPath:
			var body Plugin
			json.NewDecoder(request.Body).Decode(&body)

			if body.Name != TestPlugin.Name {
				t.Error("Plugin name is not correct")
			}

			pluginCreated = true
		}

	}))
	defer ts.Close()

	connectionBundle := getHTTPRequestBundle(ts.URL)
	config := make(map[string][]interface{})

	config[PluginsPath] = []interface{}{
		map[string]string{"name": TestPlugin.Name},
	}

	createEntries(connectionBundle.Client, ts.URL, config)

	if !pluginCreated {
		t.Error("Plugin was not created")
	}
}

func TestPluginCreatedForCorrespondingService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		serviceExternalId := "service1"

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case ServicesPath:
			body := fmt.Sprintf(`{"id": "%s"}`, serviceExternalId)
			io.WriteString(w, body)
		case PluginsPath:
			var body Plugin
			json.NewDecoder(request.Body).Decode(&body)

			if body.ServiceId != serviceExternalId {
				t.Error("Plugin created with wrong service id")
			}
		}
	}))
	defer ts.Close()

	connectionBundle := getHTTPRequestBundle(ts.URL)
	config := make(map[string][]interface{})

	serviceLocalId := "local-id"
	config[ServicesPath] = []interface{}{
		map[string]string{"id": serviceLocalId, "name": "test-service"},
	}
	config[PluginsPath] = []interface{}{
		map[string]string{"name": "test-plugin", "service_id": serviceLocalId},
	}

	createEntries(connectionBundle.Client, ts.URL, config)
}

func TestPluginCreatedForCorrespondingRoute(t *testing.T) {
	routesPath := getRoutesURL()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusCreated)

		routeExternalId := "route2"

		// Use path without slash ([1:])
		switch path := request.URL.Path[1:]; path {
		case ServicesPath:
			body := `{"id": "service2"}`
			io.WriteString(w, body)
		case routesPath:
			body := fmt.Sprintf(`{"id": "%s"}`, routeExternalId)
			io.WriteString(w, body)
		case PluginsPath:
			var body Plugin
			json.NewDecoder(request.Body).Decode(&body)

			if body.RouteId != routeExternalId{
				t.Error("Plugin created with wrong route id")
			}
		}
	}))
	defer ts.Close()

	connectionBundle := getHTTPRequestBundle(ts.URL)
	config := make(map[string][]interface{})

	config[ServicesPath] = []interface{}{
		TestEmailService,
	}
	config[PluginsPath] = []interface{}{
		map[string]string{"name": "test-plugin", "route_id": TestEmailService.Routes[0].Id},
	}

	createEntries(connectionBundle.Client, ts.URL, config)
}

func TestServiceCreatedRoutesFailed(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		routesPathElements := []string{ServicesPath, TestEmailService.Name, RoutesPath}
		routesPath := strings.Join(routesPathElements, "/")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			// Use path without slash ([1:])
			switch path := request.URL.Path[1:]; path {
			case ServicesPath:
				var body Service
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

		prepareAndCreateService(ts.URL, &ConcurrentStringMap{store: make(map[string]string)})
	}

	err := runExit("TestServiceCreatedRoutesFailed")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}
