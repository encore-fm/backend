package systest

import (
	"fmt"
	"net/http"
)

func UserJoin(username string, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/join/%v", BackendBaseUrl, username, sessionID)

	return http.Post(endpointUrl, "", nil)
}
