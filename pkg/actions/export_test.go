package actions

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"io"
	"os"
	"os/exec"
	"reflect"
	"fmt"
)

func getTestServer(resourcePath, body string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := request.URL.Path[1:]; path {
		case resourcePath:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, body)

		default:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": []}`)
		}
	}))

	return ts
}

func TestGetServicesAndRoutesPreparedConfig(t *testing.T) {
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
	services := preparedConfig[ServicesPath].([]Service)

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

func TestGetCertificatesPreparedConfig(t *testing.T) {
	answerBody := `{"data": [
		{"id": "1", "snis": ["domain.tld"]},
		{"id": "2"}
	]}`

	ts := getTestServer(CertificatesPath, answerBody)
	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	certificates := reflect.ValueOf(preparedConfig[CertificatesPath])

	if certificates.Len() != 2 {
		t.Fatalf("2 certificates should be exported")
	}

	if len(certificates.Index(0).Interface().(Certificate).Snis) != 1 {
		t.Fatalf("Exported certificate should have 1 sni")
	}
}

func TestGetConsumersPreparedConfig(t *testing.T) {
	consumer1Username := "john"
	consumer1CustomId := "1"

	consumer2Username := "alex"
	consumer2CustomId := "2"

	answerBody := fmt.Sprintf(`{"data": [
		{"id": "1", "username": "%s", "created_at": 1422386534, "customId": "%s"},
		{"id": "2", "username": "%s", "created_at": 1422386534, "custom_id": "%s"}
	]}`, consumer1Username, consumer1CustomId, consumer2Username, consumer2CustomId)

	ts := getTestServer(ConsumersPath, answerBody)
	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	consumers := reflect.ValueOf(preparedConfig[ConsumersPath])

	if consumers.Len() != 2 {
		t.Fatalf("2 consumers should be exported")
	}

	if consumers.Index(0).Interface().(Consumer).Username != consumer1Username {
		t.Fatalf("First consumer should have name john")
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
