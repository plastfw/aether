package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendStationAndToggleFavorite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stations.toml")
	if err := EnsureDefault(path); err != nil {
		t.Fatalf("EnsureDefault: %v", err)
	}

	station := Station{Name: "Test Radio", URL: "https://example.com/stream", Provider: "generic"}
	added, err := AppendStation(path, station)
	if err != nil {
		t.Fatalf("AppendStation: %v", err)
	}
	if !added {
		t.Fatal("expected station to be added")
	}

	added, err = AppendStation(path, station)
	if err != nil {
		t.Fatalf("AppendStation duplicate: %v", err)
	}
	if added {
		t.Fatal("expected duplicate station not to be added")
	}

	cfg, updated, favorite, err := ToggleFavorite(path, station.URL)
	if err != nil {
		t.Fatalf("ToggleFavorite: %v", err)
	}
	if !favorite || !updated.Favorite {
		t.Fatalf("expected favorite=true, got favorite=%v updated=%+v", favorite, updated)
	}
	if len(cfg.Stations) == 0 || cfg.Stations[0].URL != station.URL {
		t.Fatalf("favorite station should be sorted first, got %+v", cfg.Stations)
	}
}

func TestSaveCreatesBackupAndPreservesFavoriteOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stations.toml")
	cfg := Config{Stations: []Station{
		{Name: "Regular", URL: "https://example.com/regular"},
		{Name: "Favorite", URL: "https://example.com/favorite", Favorite: true},
	}}

	for i := 0; i < 6; i++ {
		if err := Save(path, cfg); err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}

	backups, err := filepath.Glob(filepath.Join(dir, "backups", "stations.toml.*.bak"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(backups) != 3 {
		t.Fatalf("backup count = %d, want 3", len(backups))
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Stations[0].Name; got != "Favorite" {
		t.Fatalf("favorite station should be first, got %q", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
}

func TestDeleteStation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stations.toml")
	keep := Station{Name: "Keep", URL: "https://example.com/keep", Provider: "generic"}
	remove := Station{Name: "Remove", URL: "https://example.com/remove", Provider: "generic"}

	if err := Save(path, Config{Stations: []Station{keep, remove}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg, removed, err := DeleteStation(path, remove.URL)
	if err != nil {
		t.Fatalf("DeleteStation: %v", err)
	}
	if removed.URL != remove.URL {
		t.Fatalf("removed URL = %q, want %q", removed.URL, remove.URL)
	}
	if len(cfg.Stations) != 1 || cfg.Stations[0].URL != keep.URL {
		t.Fatalf("unexpected stations after delete: %+v", cfg.Stations)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Stations) != 1 || loaded.Stations[0].URL != keep.URL {
		t.Fatalf("unexpected persisted stations after delete: %+v", loaded.Stations)
	}

	if _, _, err := DeleteStation(path, "https://example.com/missing"); err == nil {
		t.Fatal("expected error for missing station")
	}
}

func TestRenameStation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stations.toml")

	original := Station{
		Name:     "Old Name",
		URL:      "https://example.com/stream",
		Provider: "generic",
	}

	if err := Save(path, Config{Stations: []Station{original}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg, err := RenameStation(path, original.URL, "New Name")
	if err != nil {
		t.Fatalf("RenameStation: %v", err)
	}

	if len(cfg.Stations) != 1 {
		t.Fatalf("expected 1 station, got %d", len(cfg.Stations))
	}

	if cfg.Stations[0].Name != "New Name" {
		t.Fatalf("station name = %q, want %q", cfg.Stations[0].Name, "New Name")
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Stations[0].Name != "New Name" {
		t.Fatalf("persisted station name = %q, want %q", loaded.Stations[0].Name, "New Name")
	}

	if _, err := RenameStation(path, original.URL, "   "); err == nil {
		t.Fatal("expected error for empty station name")
	}

	if _, err := RenameStation(path, "https://example.com/missing", "Name"); err == nil {
		t.Fatal("expected error for missing station")
	}
}
