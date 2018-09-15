package actions

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"io"
	"os"
	"os/exec"
	"reflect"
)

func TestGetPreparedConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := request.URL.Path[1:]; path {

		case ServicesPath:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, `{"data": [{"id": "1"}]}`)

		case RoutesPath:
			w.WriteHeader(http.StatusOK)

			answerBody := `{"data": [
				{
					"id": "2", 
					"service": {"id": "1"},
					"paths": ["/rest/path"]
				}
			]}`

			io.WriteString(w, answerBody)
		}
	}))

	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)
	services := preparedConfig[ServicesPath].([]ServicePrepared)

	if len(services) != 1 {
		t.Fatalf("1 service should be exported")
	}

	if len(services[0].Routes) != 1 {
		t.Fatalf("Exported service should have 1 route")
	}

	if len(services[0].Routes[0].Paths) != 1 {
		t.Fatalf("Exported route should have 1 path")
	}
}

// Test resources export that is handled through ResourceBundle slice
func TestResourceBundlesPreparedConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := request.URL.Path[1:]; path {
		case CertificatesPath:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, `{"data": [
				{"id": "1", "snis": ["domain.tld"]},
				{"id": "2"}
			]}`)

		default:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": []}`)
		}
	}))
	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	certificates := reflect.ValueOf(preparedConfig[CertificatesPath])

	if certificates.Len() != 2 {
		t.Fatalf("2 certificates should be exported")
	}

	if len(certificates.Index(0).Interface().(*CertificatePrepared).Snis) != 1 {
		t.Fatalf("Exported certificate should have 1 sni")
	}
}

func TestExportCannotConnect(t *testing.T) {
	if os.Getenv("CHECK_EXIT") == "1" {
		flushAll(DefaultURL)
	}

	err := runExit("TestExportCannotConnect")
	e, ok := err.(*exec.ExitError)

	if ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}
