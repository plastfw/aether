package radiobrowser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchBuildsRequestAndDecodesStations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("name"); got != "jazz" {
			t.Fatalf("name query = %q, want jazz", got)
		}
		if got := r.URL.Query().Get("tag"); got != "jazz" {
			t.Fatalf("tag query = %q, want jazz", got)
		}
		if got := r.URL.Query().Get("limit"); got != "3" {
			t.Fatalf("limit = %q, want 3", got)
		}
		if got := r.URL.Query().Get("offset"); got != "6" {
			t.Fatalf("offset = %q, want 6", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Station{{Name: "Jazz FM", URL: "http://raw", URLResolved: "http://resolved", Bitrate: 128}})
	}))
	defer server.Close()

	oldBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = oldBaseURL }()

	stations, err := Search(" jazz ", 3, 6)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(stations) != 1 {
		t.Fatalf("expected one station, got %d", len(stations))
	}
	if stations[0].StreamURL() != "http://resolved" {
		t.Fatalf("StreamURL = %q", stations[0].StreamURL())
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	if _, err := Search("   ", 10, 0); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearchReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadGateway)
	}))
	defer server.Close()

	oldBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = oldBaseURL }()

	if _, err := Search("rock", 1, 0); err == nil {
		t.Fatal("expected HTTP status error")
	}
}
