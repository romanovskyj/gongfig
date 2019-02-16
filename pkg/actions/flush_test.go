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
	certificateDeleted := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := getResourcePath(request.URL.Path); path {

		case ServicesPath:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": [{"id": "1"}]}`)

		case RoutesPath:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": [{"id": "2"}]}`)

		case CertificatesPath:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": [{"id": "3"}]}`)

		case "services/1":
			w.WriteHeader(http.StatusNoContent)
			serviceDeleted = true

		case "routes/2":
			w.WriteHeader(http.StatusNoContent)
			routeDeleted = true

		case "certificates/3":
			w.WriteHeader(http.StatusNoContent)
			certificateDeleted = true
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
	if !certificateDeleted {
		t.Error("Certificate was not deleted")
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
