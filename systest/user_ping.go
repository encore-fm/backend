package systest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func UserPing(username, secret, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/ping", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)
	return client.Do(req)
}