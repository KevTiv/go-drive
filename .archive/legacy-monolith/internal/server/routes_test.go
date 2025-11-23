package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler directly
	handler := http.HandlerFunc(s.HelloWorldHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body (trim whitespace/newlines for comparison)
	expected := `{"message":"Hello World"}`
	actual := strings.TrimSpace(rr.Body.String())
	if actual != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", actual, expected)
	}
}

func stringPtr(s string) *string {
	return &s
}
