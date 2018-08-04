package actions

import (
	"testing"
	"os"
	"net/http"
	"time"
	"os/exec"
)

func TestCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		service := ServicePrepared{
			Name: "email-service",
			Host: "email.tld",
			Path: "/api/v1",
		}
		client := &http.Client{Timeout: 1 * time.Second}
		reqLimitChan := make(chan bool, 5)
		reqLimitChan <- true

		createServiceWithRoutes(client, DefaultUrl, service, reqLimitChan)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCannotConnect")
	cmd.Env = append(os.Environ(), "CHECK_EXIT=1")
	err := cmd.Run()

	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
