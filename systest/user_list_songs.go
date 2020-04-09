package systest

import (
	"fmt"
	"net/http"
)

func UserListSongs(username string, secret string, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/listSongs", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
