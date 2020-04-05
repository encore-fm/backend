package systest

import (
	"fmt"
	"net/http"
)

func AdminRemoveSong(username string, secret string, sessionID string, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/removeSong/%v", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
