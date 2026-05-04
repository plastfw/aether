package notify

import (
	"os/exec"
	"strings"

	"aether/internal/metadata"
)

func Available() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}

func Track(station string, md metadata.Metadata) error {
	if !Available() || !hasUsefulMetadata(md) {
		return nil
	}

	title := md.Display()
	args := []string{"--app-name=Aether", "Aether", title}
	if strings.TrimSpace(station) != "" {
		args = []string{"--app-name=Aether", "Aether", title + "\n" + station}
	}
	return exec.Command("notify-send", args...).Run()
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
