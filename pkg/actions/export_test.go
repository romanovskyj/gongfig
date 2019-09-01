package actions

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func getTestServer(resourcePath, body string) (*httptest.Server, error) {
	if isJSONString(body) == false {
		return nil, errors.New("specifed body does not have json format")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := getResourcePath(request.URL.Path); path {
		case resourcePath:
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, body)

		default:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"data": []}`)
		}
	}))

	return ts, nil
}

func TestGetServicesAndRoutesPreparedConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := getResourcePath(request.URL.Path); path {

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

	ts, _ := getTestServer(CertificatesPath, answerBody)
	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	certificates := reflect.ValueOf(preparedConfig[CertificatesPath])

	if certificates.Len() != 2 {
		t.Fatalf("2 certificates should be exported")
	}

	certMap := certificates.Index(1).Interface()

	var certificate Certificate
	mapstructure.Decode(certMap, &certificate)

	if len(certificate.Snis) != 1 {
		t.Fatalf("Exported certificate should have 1 sni")
	}
}

func TestGetConsumersPreparedConfig(t *testing.T) {
	consumer1Id := "1"
	consumer1Username := "john"
	consumer1CustomId := "1"
	consumer1Key := "key1"

	consumer2Id := "2"
	consumer2Username := "alex"
	consumer2CustomId := "2"
	consumer2Key := "key2"

	consumerAnswerBody := fmt.Sprintf(`{"data": [
		{"id": "%s", "username": "%s", "created_at": 1422386534, "customId": "%s"},
		{"id": "%s", "username": "%s", "created_at": 1422386534, "custom_id": "%s"}
	]}`, consumer1Id, consumer1Username, consumer1CustomId,
		consumer2Id, consumer2Username, consumer2CustomId)

	keyAuthAnswerBody := fmt.Sprintf(`{"data": [
		{"consumer_id": "%s", "key": "%s"},
		{"consumer_id": "%s", "key": "%s"}
	]}`, consumer1Id, consumer1Key, consumer2Id, consumer2Key)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path := getResourcePath(request.URL.Path); path {

		case ConsumersPath:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, consumerAnswerBody)

		case KeyAuthsPath:
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, keyAuthAnswerBody)
		}
	}))

	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	consumers := reflect.ValueOf(preparedConfig[ConsumersPath])

	if consumers.Len() != 2 {
		t.Fatalf("2 consumers should be exported")
	}

	username := consumers.Index(0).Interface().(Consumer).Username
	if username != consumer1Username {
		t.Fatalf("First consumer should have name %s, but it has %s", consumer1Username, username)
	}

	key := consumers.Index(0).Interface().(Consumer).Key
	if key != consumer1Key {
		t.Fatalf("First consumer should have key %s, but it has %s", consumer1Key, key)
	}
}

func TestGetPluginsPreparedConfig(t *testing.T) {
	answerBody := `{"data": [
		{
          "id": "1", 
          "config": {
            "key": "value"
          }
        },
		{"id": "2"}
	]}`

	ts, _ := getTestServer(PluginsPath, answerBody)
	defer ts.Close()

	preparedConfig := getPreparedConfig(ts.URL)

	plugins := reflect.ValueOf(preparedConfig[PluginsPath])

	if plugins.Len() != 2 {
		t.Fatalf("2 plugins should be exported")
	}

	pluginMap := plugins.Index(0).Interface()
	var plugin Plugin

	mapstructure.Decode(pluginMap, &plugin)

	if plugin.Id != "1" {
		t.Fatalf("Exported plugin should have correct id")
	}
}
