package systest

import (
	"fmt"
	"net/http"
)

func UserGetAuthToken(username, secret, sessionID string) (*http.Response, error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/authToken", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
