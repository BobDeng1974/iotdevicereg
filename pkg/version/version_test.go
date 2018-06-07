package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thingful/iotdevicereg/pkg/version"
)

func TestVersionString(t *testing.T) {
	expected := "UNKNOWN (linux/amd64). build date: UNKNOWN"
	got := version.VersionString()

	assert.Equal(t, expected, got)
}
