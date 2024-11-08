package tests

import (
	"kasmlink/pkg/api"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test functions
func TestCreateHTTPClient(t *testing.T) {
	client := api.CreateHTTPClient(true)
	if client == nil {
		t.Errorf("Expected non-nil HTTP client")
	}

	if client.Transport == nil {
		t.Errorf("Expected non-nil Transport in HTTP client")
	}
}

func TestHandleResponse(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request to test server: %v", err)
	}

	body, err := api.HandleResponse(resp, http.StatusOK)
	if err != nil {
		t.Errorf("HandleResponse returned an error: %v", err)
	}

	expected := "Test response"
	if string(body) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(body))
	}
}

func TestMakeGetRequest(t *testing.T) {
	api := &api.KasmAPI{APIKey: "dummy-key", SkipTLSVerification: true}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer dummy-key" {
			t.Errorf("Expected Authorization header to be 'Bearer dummy-key'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	}))
	defer ts.Close()

	body, err := api.MakeGetRequest(ts.URL)
	if err != nil {
		t.Errorf("MakeGetRequest returned an error: %v", err)
	}

	expected := "Test response"
	if string(body) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(body))
	}
}

func TestMakePostRequest(t *testing.T) {
	api := &api.KasmAPI{APIKey: "dummy-key", SkipTLSVerification: true}
	payload := map[string]string{"key": "value"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer dummy-key" {
			t.Errorf("Expected Authorization header to be 'Bearer dummy-key'")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	}))
	defer ts.Close()

	body, err := api.MakePostRequest(ts.URL, payload)
	if err != nil {
		t.Errorf("MakePostRequest returned an error: %v", err)
	}

	expected := "Test response"
	if string(body) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(body))
	}
}
