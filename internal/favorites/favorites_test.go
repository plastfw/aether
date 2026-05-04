package favorites

import (
	"path/filepath"
	"testing"

	"aether/internal/config"
	"aether/internal/metadata"
)

func TestStoreAddDedupesRemoveAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorite_tracks.jsonl")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	station := config.Station{Name: "Station", URL: "https://example.com"}
	md := metadata.Metadata{Artist: "Artist", Title: "Title"}

	added, err := store.Add(station, md)
	if err != nil {
		t.Fatalf("Add first: %v", err)
	}
	if !added {
		t.Fatal("expected first add to succeed")
	}
	added, err = store.Add(station, md)
	if err != nil {
		t.Fatalf("Add duplicate: %v", err)
	}
	if added {
		t.Fatal("expected duplicate add to be ignored")
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 || entries[0].Display() != "Artist — Title" {
		t.Fatalf("unexpected entries: %+v", entries)
	}

	if err := store.Remove(0); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	entries, err = store.List()
	if err != nil {
		t.Fatalf("List after remove: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty after remove, got %+v", entries)
	}

	added, err = store.Add(station, md)
	if err != nil || !added {
		t.Fatalf("Add after remove: added=%v err=%v", added, err)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	entries, err = store.List()
	if err != nil {
		t.Fatalf("List after clear: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty after clear, got %+v", entries)
	}
}

func TestStoreIgnoresEmptyMetadata(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "favorite_tracks.jsonl"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	added, err := store.Add(config.Station{Name: "Station"}, metadata.Metadata{})
	if err != nil {
		t.Fatalf("Add empty metadata: %v", err)
	}
	if added {
		t.Fatal("empty metadata should not be added")
	}
}
