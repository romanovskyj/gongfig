package actions

import (
	"testing"
	"os"
	"net/http"
	"time"
	"os/exec"
	"strings"
	"net/http/httptest"
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
		prepareAndCreateService(DefaultUrl)
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
