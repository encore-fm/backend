// +build !ci

package systest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerSeek(t *testing.T) {
	dropDB()
	setupDB()
	_, err := resetPlayerController(TestSessionID)
	assert.NoError(t, err)

	position := 60 * time.Second

	time.Sleep(2 * time.Second)

	resp, err := PlayerSeek(TestAdminUsername, TestAdminSecret, TestSessionID, position)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// short pause because PlayerPlay is async
	time.Sleep(500 * time.Millisecond)
	p, err := getPlayer()
	assert.NoError(t, err)
	assert.WithinDuration(t, testNow.Add(position), testNow.Add(p.Progress()), 1000*time.Millisecond)
}
