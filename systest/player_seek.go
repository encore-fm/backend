package systest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func PlayerSeek(username, secret, sessionID string, position time.Duration) (*http.Response, error) {
	endpointUrl := fmt.Sprintf(
		"%v/users/%v/player/seek/%v",
		BackendBaseUrl,
		username,
		strconv.Itoa(int(position.Milliseconds())),
	)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
