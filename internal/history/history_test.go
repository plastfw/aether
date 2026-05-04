package history

import (
	"os"
	"path/filepath"
	"testing"

	"aether/internal/config"
	"aether/internal/metadata"
)

func TestRecorderDedupesConsecutiveTracks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	station := config.Station{Name: "Station", URL: "https://example.com"}
	md := metadata.Metadata{Artist: "Artist", Title: "Title"}

	recorder, err := NewRecorder(path)
	if err != nil {
		t.Fatalf("NewRecorder: %v", err)
	}
	if err := recorder.Record(station, md); err != nil {
		t.Fatalf("Record first: %v", err)
	}
	if err := recorder.Record(station, md); err != nil {
		t.Fatalf("Record duplicate: %v", err)
	}

	restartedRecorder, err := NewRecorder(path)
	if err != nil {
		t.Fatalf("NewRecorder restarted: %v", err)
	}
	if err := restartedRecorder.Record(station, md); err != nil {
		t.Fatalf("Record duplicate after restart: %v", err)
	}

	entries, err := ReadLast(path, 10)
	if err != nil {
		t.Fatalf("ReadLast: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 deduped entry, got %d", len(entries))
	}
	if entries[0].Artist != "Artist" || entries[0].Title != "Title" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}

func TestTrimKeepsLastEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	data := []byte(
		`{"artist":"A1","title":"T1","time":"2026-01-01T00:00:00Z"}` + "\n" +
			`{"artist":"A2","title":"T2","time":"2026-01-01T00:01:00Z"}` + "\n" +
			`{"artist":"A3","title":"T3","time":"2026-01-01T00:02:00Z"}` + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := Trim(path, 2); err != nil {
		t.Fatalf("Trim: %v", err)
	}

	entries, err := ReadLast(path, 10)
	if err != nil {
		t.Fatalf("ReadLast: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries len = %d, want 2", len(entries))
	}
	if entries[0].Artist != "A2" || entries[1].Artist != "A3" {
		t.Fatalf("unexpected entries after trim: %+v", entries)
	}
}

func TestReadLastSkipsInvalidLinesAndLimits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	data := []byte("not-json\n" +
		`{"artist":"A1","title":"T1","time":"2026-01-01T00:00:00Z"}` + "\n" +
		`{"artist":"A2","title":"T2","time":"2026-01-01T00:01:00Z"}` + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	entries, err := ReadLast(path, 1)
	if err != nil {
		t.Fatalf("ReadLast: %v", err)
	}
	if len(entries) != 1 || entries[0].Artist != "A2" {
		t.Fatalf("expected last valid entry, got %+v", entries)
	}
}
