package systest

import (
	"fmt"
	"net/http"
)

func PlayerSkip(username, secret, sessionID string) (*http.Response, error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/player/skip", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
