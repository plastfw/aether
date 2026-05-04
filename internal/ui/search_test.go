package ui

import (
	"path/filepath"
	"testing"

	"aether/internal/config"
	"aether/internal/radiobrowser"
	"aether/internal/settings"
)

func TestAddSelectedSearchResultPersistsStation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stations.toml")
	initial := config.Station{Name: "Existing", URL: "https://example.com/existing", Provider: "generic"}
	if err := config.Save(path, config.Config{Stations: []config.Station{initial}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := Model{
		stations:   []config.Station{initial},
		configPath: path,
		settings:   settings.Default(),
		search: searchState{
			open: true,
			results: []radiobrowser.Station{{
				Name:        "Search Radio",
				URL:         "https://example.com/raw",
				URLResolved: "https://example.com/resolved",
			}},
		},
	}

	updated, cmd := m.addSelectedSearchResult()
	if cmd != nil {
		t.Fatal("expected no command")
	}
	if updated.search.open {
		t.Fatal("expected search to close after add")
	}
	if len(updated.stations) != 2 {
		t.Fatalf("stations len = %d, want 2", len(updated.stations))
	}
	if updated.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", updated.cursor)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Stations) != 2 {
		t.Fatalf("persisted stations len = %d, want 2", len(loaded.Stations))
	}
	if got := loaded.Stations[1].URL; got != "https://example.com/resolved" {
		t.Fatalf("persisted URL = %q, want resolved URL", got)
	}
}
