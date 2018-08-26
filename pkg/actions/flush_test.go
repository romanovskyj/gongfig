package actions

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"io"
	"os"
	"os/exec"
)

func TestConfigFlushed(t *testing.T) {
	serviceDeleted := false
	routeDeleted := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := request.URL.Path[1:]; path {

		case ServicesKey:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, `{"data": [{"id": "1"}]}`)

		case RoutesKey:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, `{"data": [{"id": "2"}]}`)

		case "services/1":
			w.WriteHeader(http.StatusNoContent)
			serviceDeleted = true

			case "routes/2":
			w.WriteHeader(http.StatusNoContent)
			routeDeleted = true
		}
	}))

	defer ts.Close()

	flushAll(ts.URL)

	if !serviceDeleted {
		t.Error("Service was not deleted")
	}

	if !routeDeleted {
		t.Error("Route was not deleted")
	}
}

func TestFlushCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		flushAll(DefaultURL)
	}

	err := runExit("TestFlushCannotConnect")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}
