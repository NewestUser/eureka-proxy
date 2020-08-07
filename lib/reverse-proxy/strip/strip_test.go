package strip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStrip(t *testing.T) {

	stringStrip, err := New("foo:bar")

	assert.Nil(t, err)

	result := stringStrip.apply("/service/foo-api/gar")

	assert.Equal(t, "/service/bar-api/gar", result)
}

func TestRemovePartOfPath(t *testing.T) {
	stringStrip, err := New("service-api/:")

	assert.Nil(t, err)

	result := stringStrip.apply("/service-api/foo/bar")

	assert.Equal(t, "/foo/bar", result)
}

func TestDoNotStripAnything(t *testing.T) {
	stringStrip, err := New("")

	assert.Nil(t, err)

	result := stringStrip.apply("/service-api/foo/bar")

	assert.Equal(t, "/service-api/foo/bar", result)
}

func TestErrorForWrongFormat(t *testing.T) {
	stringStrip, err := New("no-dots")

	assert.NotNil(t, err)
	assert.Nil(t, stringStrip)
}
