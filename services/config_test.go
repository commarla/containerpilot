package services

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestFragmentServices struct {
	Services []Service
}

func TestServiceConfigValidateExec(t *testing.T) {

	cfg := &ServiceConfig{
		Exec:        []string{"/bin/to/healthcheck/for/service/A.sh", "A1", "A2"},
		execTimeout: "1ms",
	}
	validateCommandParsed(t, cfg,
		"/bin/to/healthcheck/for/service/A.sh",
		[]string{"A1", "A2"})

	cfg = &ServiceConfig{
		Exec:        "/bin/to/healthcheck/for/service/B.sh B1 B2",
		execTimeout: "1ms",
	}
	validateCommandParsed(t, cfg,
		"/bin/to/healthcheck/for/service/B.sh",
		[]string{"B1", "B2"})

	cfg = &ServiceConfig{
		Name:        "myName",
		Exec:        "/bin/true",
		execTimeout: "xx",
	}
	_, err := NewService(cfg)
	expected := "could not parse `health` in service myName: time: invalid duration xx"
	if err.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, err)
	}
}

func TestServicesConfigValidation(t *testing.T) {
	var raw []interface{}
	json.Unmarshal([]byte(`[{"name": ""}]`), &raw)
	_, err := NewServices(raw, nil)
	validateServiceConfigError(t, err, "`name` must not be blank")
	raw = nil

	json.Unmarshal([]byte(`[{"name": "myName", "port": 80}]`), &raw)
	_, err = NewServices(raw, nil)
	validateServiceConfigError(t, err,
		"`poll` must be > 0 in service `myName` when `port` is set")

	json.Unmarshal([]byte(`[{"name": "myName", "port": 80, "poll": 1}]`), &raw)
	_, err = NewServices(raw, nil)
	validateServiceConfigError(t, err,
		"`ttl` must be > 0 in service `myName` when `port` is set")

	json.Unmarshal([]byte(`[{"name": "myName", "poll": 1, "ttl": 1}]`), &raw)
	_, err = NewServices(raw, nil)
	validateServiceConfigError(t, err,
		"`heartbeat` and `ttl` may not be set in service `myName` if `port` is not set")

	// no health check shouldn't return an error
	json.Unmarshal([]byte(`[{"name": "myName", "poll": 1, "ttl": 1, "port": 80, "interfaces": "inet"}]`), &raw)
	_, err = NewServices(raw, nil)
	validateServiceConfigError(t, err, "")

}

func TestServicesConsulExtrasEnableTagOverride(t *testing.T) {
	jsonFragment := []byte(`[
{
  "name": "serviceA",
  "port": 8080,
  "interfaces": "inet",
  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
  "poll": 30,
  "ttl": 19,
  "timeout": "1ms",
  "tags": ["tag1","tag2"],
  "consul": {
	  "enableTagOverride": true
  }
}
]`)

	if services, err := NewServices(decodeJSONRawService(t, jsonFragment), nil); err != nil {
		t.Fatalf("could not parse service JSON: %s", err)
	} else {
		if services[0].Definition.ConsulExtras.EnableTagOverride != true {
			t.Errorf("ConsulExtras should have had EnableTagOverride set to true.")
		}
	}
}

func TestInvalidServicesConsulExtrasEnableTagOverride(t *testing.T) {
	jsonFragment := []byte(`[
{
  "name": "serviceA",
  "port": 8080,
  "interfaces": "inet",
  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
  "poll": 30,
  "ttl": 19,
  "timeout": "1ms",
  "tags": ["tag1","tag2"],
  "consul": {
	  "enableTagOverride": "nope"
  }
}
]`)

	if _, err := NewServices(decodeJSONRawService(t, jsonFragment), nil); err == nil {
		t.Errorf("ConsulExtras should have thrown error about EnableTagOverride being a string.")
	}
}

func TestServicesConsulExtrasDeregisterCriticalServiceAfter(t *testing.T) {
	jsonFragment := []byte(`[
{
  "name": "serviceA",
  "port": 8080,
  "interfaces": "inet",
  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
  "poll": 30,
  "ttl": 19,
  "timeout": "1ms",
  "tags": ["tag1","tag2"],
  "consul": {
	  "deregisterCriticalServiceAfter": "40m"
  }
}
]`)

	if services, err := NewServices(decodeJSONRawService(t, jsonFragment), nil); err != nil {
		t.Fatalf("could not parse service JSON: %s", err)
	} else {
		if services[0].Definition.ConsulExtras.DeregisterCriticalServiceAfter != "40m" {
			t.Errorf("ConsulExtras should have had DeregisterCriticalServiceAfter set to '40m'.")
		}
	}
}

func TestInvalidServicesConsulExtrasDeregisterCriticalServiceAfter(t *testing.T) {
	jsonFragment := []byte(`[
{
  "name": "serviceA",
  "port": 8080,
  "interfaces": "inet",
  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
  "poll": 30,
  "ttl": 19,
  "timeout": "1ms",
  "tags": ["tag1","tag2"],
  "consul": {
	  "deregisterCriticalServiceAfter": "nope"
  }
}
]`)

	if _, err := NewServices(decodeJSONRawService(t, jsonFragment), nil); err == nil {
		t.Errorf("error should have been generated for duration 'nope'.")
	}
}

// ------------------------------------------
// test helpers

func decodeJSONRawService(t *testing.T, testJSON json.RawMessage) []interface{} {
	var raw []interface{}
	if err := json.Unmarshal(testJSON, &raw); err != nil {
		t.Fatalf("unexpected error decoding JSON:\n%s\n%v", testJSON, err)
	}
	return raw
}

func validateCommandParsed(t *testing.T, cfg *ServiceConfig,
	expectedExec string, expectedArgs []string) {
	service, err := NewService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(service.exec.Exec, expectedExec) {
		t.Errorf("executable not configured: %s != %s", service.exec.Exec, expectedExec)
	}
	if !reflect.DeepEqual(service.exec.Args, expectedArgs) {
		t.Errorf("arguments not configured: %s != %s", service.exec.Args, expectedArgs)
	}
}

func validateServiceConfigError(t *testing.T, err error, expected string) {
	if expected == "" {
		if err != nil {
			t.Fatalf("expected no error but got '%s'", err)
		}
	} else {
		if err == nil {
			t.Fatalf("expected '%s' but got nil error", expected)
		}
		if err.Error() != expected {
			t.Fatalf("expected '%s' but got '%s'", expected, err.Error())
		}
	}
}
