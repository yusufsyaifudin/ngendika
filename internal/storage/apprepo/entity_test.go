package apprepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	app := NewApp("abc", "abc")
	assert.Equal(t, "abc", app.ClientID)
}
