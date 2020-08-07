package eureka

import (
	"testing"
)

// This is an integration test and eureka must be up and running
func testFetchAlreadyRegisteredInstanceFromEureka(t *testing.T) {

	got, err := GetInstanceURL("127.0.0.1:8761", "foo-service")
	want := "http://localhost:8017"

	if err != nil {
		t.Error(err)
	}

	if got != want {
		t.Errorf(" got: %s want: %s", got, want)
	}
}

// This is an integration test and eureka must be up and running
func testFetchConfigInstanceFromEureka(t *testing.T) {
	got, err := GetInstanceURL("127.0.0.1:8761", "config")
	want := "http://localhost:8762"

	if err != nil {
		t.Error(err)
	}

	if got != want {
		t.Errorf(" got: %s want: %s", got, want)
	}
}

// This is an integration test and eureka must be up and running
func testNormalizeLocalhost(t *testing.T) {
	got, err := GetInstanceURL("localhost", "config")
	want := "http://localhost:8762"

	if err != nil {
		t.Error(err)
	}

	if got != want {
		t.Errorf(" got: %s want: %s", got, want)
	}
}
