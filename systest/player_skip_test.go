// +build !ci

package systest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerSkip(t *testing.T) {
	dropDB()
	setupDB()

	_, err := resetPlayerController(TestSessionID)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	p, err := getPlayer()
	assert.NoError(t, err)
	assert.Equal(t, testSession.SongList[0].ID, p.CurrentSong.ID)

	resp, err := PlayerSkip(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// short pause because PlayerPlay is async
	time.Sleep(200 * time.Millisecond)
	p, err = getPlayer()
	assert.NoError(t, err)
	assert.Equal(t, testSession.SongList[1].ID, p.CurrentSong.ID)
}
