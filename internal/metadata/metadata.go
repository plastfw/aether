package metadata

import (
	"context"
	"strings"

	"aether/internal/config"
)

type Metadata struct {
	Artist string
	Title  string
	Album  string
	ArtURL string
	Raw    string
	Source string
}

type Provider interface {
	Fetch(ctx context.Context, station config.Station) (Metadata, error)
}

func (m Metadata) Display() string {
	if m.Artist != "" && m.Title != "" {
		return m.Artist + " — " + m.Title
	}
	if m.Title != "" {
		return m.Title
	}
	if m.Raw != "" {
		return m.Raw
	}
	return "—"
}

func FirstString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := obj[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func SplitICYTitle(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	for _, sep := range []string{" - ", " – ", " — "} {
		parts := strings.SplitN(raw, sep, 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", raw
}
