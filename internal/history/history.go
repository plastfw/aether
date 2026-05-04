package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"aether/internal/config"
	"aether/internal/metadata"
)

type Entry struct {
	Time     time.Time `json:"time"`
	Station  string    `json:"station"`
	URL      string    `json:"url"`
	Provider string    `json:"provider"`
	Artist   string    `json:"artist,omitempty"`
	Title    string    `json:"title,omitempty"`
	Album    string    `json:"album,omitempty"`
	Raw      string    `json:"raw,omitempty"`
	Source   string    `json:"source,omitempty"`
}

type Recorder struct {
	path    string
	lastKey string
	mu      sync.Mutex
}

func DefaultPath() (string, error) {
	if dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); dataHome != "" {
		return filepath.Join(dataHome, "aether", "history.jsonl"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "aether", "history.jsonl"), nil
}

func NewRecorder(path string) (*Recorder, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("history path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	recorder := &Recorder{path: path}
	entries, err := ReadLast(path, 1)
	if err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		recorder.lastKey = entries[len(entries)-1].dedupeKey()
	}
	return recorder, nil
}

func (r *Recorder) Record(station config.Station, md metadata.Metadata) error {
	if r == nil || !hasUsefulMetadata(md) {
		return nil
	}

	entry := Entry{
		Time:     time.Now(),
		Station:  station.Name,
		URL:      station.URL,
		Provider: station.MetadataProvider(),
		Artist:   md.Artist,
		Title:    md.Title,
		Album:    md.Album,
		Raw:      md.Raw,
		Source:   md.Source,
	}

	key := entry.dedupeKey()
	if key == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if key == r.lastKey {
		return nil
	}

	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		return err
	}

	r.lastKey = key
	return nil
}

func (r *Recorder) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func Trim(path string, maxEntries int) error {
	if maxEntries <= 0 {
		return nil
	}
	entries, err := ReadLast(path, maxEntries)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func ReadLast(path string, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 20
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	entries := make([]Entry, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		if len(entries) == limit {
			copy(entries, entries[1:])
			entries[len(entries)-1] = entry
		} else {
			entries = append(entries, entry)
		}
	}
	return entries, scanner.Err()
}

func (e Entry) dedupeKey() string {
	parts := []string{e.Station, e.Artist, e.Title, e.Raw}
	key := strings.ToLower(strings.TrimSpace(strings.Join(parts, "|")))
	return key
}

func hasUsefulMetadata(md metadata.Metadata) bool {
	return useful(md.Artist) || useful(md.Title) || useful(md.Raw)
}

func useful(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "—", "-", "unknown", "n/a", "null", "none":
		return false
	default:
		return true
	}
}
