package systest

import (
	"fmt"
	"net/http"
)

func UserGetSessionInfo(sessionID string) (*http.Response, error) {
	endpointUrl := fmt.Sprintf("%v/sessionInfo/%v", BackendBaseUrl, sessionID)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
