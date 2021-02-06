package mobile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCompare(t *testing.T) {
	assert.Equal(t, 0, VersionCompare("5.2.0", "5.2.0", "."))
	assert.Equal(t, 1, VersionCompare("5.2.1", "5.2.0", "."))
	assert.Equal(t, 1, VersionCompare("5.11.1", "5.2.0", "."))
	assert.Equal(t, 1, VersionCompare("6.2.0", "5.2.0", "."))
	assert.Equal(t, 1, VersionCompare("6.11.0", "5.2.0", "."))
	assert.Equal(t, 1, VersionCompare("6.2", "5.2.0", "."))
	assert.Equal(t, -1, VersionCompare("5.2", "5.2.0", "."))
	assert.Equal(t, -1, VersionCompare("5.2", "5.2.1", "."))
	assert.Equal(t, -1, VersionCompare("5.2.0", "5.2.1", "."))
	assert.Equal(t, -1, VersionCompare("5.2.0", "6.2", "."))
}
