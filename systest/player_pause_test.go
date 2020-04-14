package systest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerPause(t *testing.T) {
	dropDB()
	setupDB()

	err := setPaused(false)
	assert.NoError(t, err)

	resp, err := PlayerPause(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// short pause because PlayerPlay is async
	time.Sleep(200 * time.Millisecond)
	p, err := getPlayer()
	assert.NoError(t, err)
	assert.WithinDuration(t, testNow.Add(90 * time.Second), testNow.Add(p.Progress()), 1 * time.Second)
}
