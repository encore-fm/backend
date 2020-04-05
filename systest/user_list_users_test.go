// +build !ci

package systest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_UserList(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID

	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	resp, err := UserList(username, secret, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	// expect http status OK
	var response []*user.ListElement

	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, int(count), len(response))
}
