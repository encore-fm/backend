package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Ping(t *testing.T) {
	handler := &handler{}
	serverHandler := ServerHandler(handler)

	req, err := http.NewRequest(
		"GET",
		"/ping",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// call handler func
	serverHandler.Ping(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
