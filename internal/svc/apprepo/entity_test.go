package apprepo

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	app := App{
		ClientID:  strings.ToLower("abc"),
		Name:      "abc",
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}
	assert.Equal(t, "abc", app.ClientID)
}
