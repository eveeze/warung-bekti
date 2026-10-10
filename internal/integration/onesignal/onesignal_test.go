package onesignal

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestOneSignalClient(t *testing.T) {
	c := NewClient("app-123", "key-123")
	if c.AppID != "app-123" || c.APIKey != "key-123" {
		t.Errorf("unexpected client config")
	}

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Basic error-key" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "notif-123"}`))
	}))
	defer server.Close()

	// Intercept requests and route to mock server
	c.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Rewrite URL to test server
			testReq, _ := http.NewRequest(req.Method, server.URL, req.Body)
			for k, v := range req.Header {
				testReq.Header[k] = v
			}
			return http.DefaultTransport.RoundTrip(testReq)
		}),
	}

	// Test 1: Send to segments ("All")
	err := c.SendNotification("Title", "Message", nil, map[string]string{"foo": "bar"})
	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}

	// Test 2: Send to playerIDs
	err = c.SendNotification("Title", "Message", []string{"player-1", "player-2"}, nil)
	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}

	// Test 3: API Error >= 400
	c.APIKey = "error-key"
	err = c.SendNotification("Title", "Message", nil, nil)
	if err == nil {
		t.Errorf("expected error for status 400")
	}
}
