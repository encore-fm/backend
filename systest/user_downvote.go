package systest

import (
	"fmt"
	"net/http"
)

func UserDownvote(username, secret, sessionID, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/vote/%v/down", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

