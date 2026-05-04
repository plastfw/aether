package favorites

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
	Station  string    `json:"station,omitempty"`
	URL      string    `json:"url,omitempty"`
	Provider string    `json:"provider,omitempty"`
	Artist   string    `json:"artist,omitempty"`
	Title    string    `json:"title,omitempty"`
	Album    string    `json:"album,omitempty"`
	Raw      string    `json:"raw,omitempty"`
	Source   string    `json:"source,omitempty"`
}

func (e Entry) Display() string {
	if e.Artist != "" && e.Title != "" {
		return e.Artist + " — " + e.Title
	}
	if e.Title != "" {
		return e.Title
	}
	if e.Raw != "" {
		return e.Raw
	}
	return "—"
}

type Store struct {
	path string
	seen map[string]struct{}
	mu   sync.Mutex
}

func DefaultPath() (string, error) { return TrackPath() }

func TrackPath() (string, error) {
	if dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); dataHome != "" {
		return filepath.Join(dataHome, "aether", "favorite_tracks.jsonl"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "aether", "favorite_tracks.jsonl"), nil
}

func NewStore(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("favorites path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: path, seen: map[string]struct{}{}}
	_ = s.loadSeen()
	return s, nil
}

func (s *Store) Add(station config.Station, md metadata.Metadata) (bool, error) {
	if s == nil || !hasUsefulMetadata(md) {
		return false, nil
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
	return s.addEntry(entry)
}

func (s *Store) List() ([]Entry, error) {
	return ReadAll(s.path)
}

func (s *Store) Remove(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := ReadAll(s.path)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(entries) {
		return nil
	}
	entries = append(entries[:index], entries[index+1:]...)
	return s.writeAllLocked(entries)
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(s.path, nil, 0o644); err != nil {
		return err
	}
	s.seen = map[string]struct{}{}
	return nil
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) addEntry(entry Entry) (bool, error) {
	key := entry.dedupeKey()
	if key == "" {
		return false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return false, nil
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return false, err
	}
	defer file.Close()
	line, err := json.Marshal(entry)
	if err != nil {
		return false, err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		return false, err
	}
	s.seen[key] = struct{}{}
	return true, nil
}

func (s *Store) loadSeen() error {
	entries, err := ReadAll(s.path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if key := entry.dedupeKey(); key != "" {
			s.seen[key] = struct{}{}
		}
	}
	return nil
}

func (s *Store) writeAllLocked(entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	s.seen = map[string]struct{}{}
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(line, '\n')); err != nil {
			return err
		}
		if key := entry.dedupeKey(); key != "" {
			s.seen[key] = struct{}{}
		}
	}
	return nil
}

func ReadAll(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	entries := []Entry{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			entries = append(entries, entry)
		}
	}
	return entries, scanner.Err()
}

func (e Entry) dedupeKey() string {
	parts := []string{e.Artist, e.Title, e.Raw}
	return strings.ToLower(strings.TrimSpace(strings.Join(parts, "|")))
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
