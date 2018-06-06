package version_test

import (
	"testing"

	"github.com/thingful/iotdevicereg/pkg/version"
)

func TestVersionString(t *testing.T) {
	expected := "UNKNOWN (linux/amd64). build date: UNKNOWN"
	got := version.VersionString()

	if got != expected {
		t.Errorf("Unexpected value, expected '%s', got '%s'", expected, got)
	}
}
