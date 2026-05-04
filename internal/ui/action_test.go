package ui

import (
	"strings"
	"testing"

	"aether/internal/config"
	"aether/internal/player"
	"aether/internal/settings"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpIgnoresGlobalPauseAction(t *testing.T) {
	m := Model{settings: settings.Default(), view: viewHelp, hasCurrent: true, mode: playbackStation}
	if action := m.actionForKey(" "); action != "" {
		t.Fatalf("help space action = %q, want empty", action)
	}
}

func TestPauseActionRequiresActivePlayback(t *testing.T) {
	m := Model{settings: settings.Default()}
	if action := m.actionForKey(" "); action != "" {
		t.Fatalf("stopped space action = %q, want empty", action)
	}

	m.hasCurrent = true
	m.mode = playbackStation
	if action := m.actionForKey(" "); action != "Pause" {
		t.Fatalf("playing space action = %q, want Pause", action)
	}
}

func TestPlaybackStatusReflectsStreamHealth(t *testing.T) {
	m := Model{hasCurrent: true, mode: playbackStation}
	if got := m.playbackStatusText(); got != "TUNING" {
		t.Fatalf("empty health status = %q, want TUNING", got)
	}

	m.health = player.Health{PausedForCache: true}
	if got := m.playbackStatusText(); got != "BUFFERING" {
		t.Fatalf("cache status = %q, want BUFFERING", got)
	}

	m.health = player.Health{CoreIdle: true}
	if got := m.playbackStatusText(); got != "NO AUDIO" {
		t.Fatalf("idle status = %q, want NO AUDIO", got)
	}

	m.health = player.Health{AudioCodec: "aac"}
	if got := m.playbackStatusText(); got != "ON AIR" {
		t.Fatalf("healthy status = %q, want ON AIR", got)
	}
}

func TestHeaderKeepsStateAndLastActionVisible(t *testing.T) {
	m := Model{
		settings:   settings.Default(),
		stations:   []config.Station{{Name: "Station", URL: "https://example.com"}},
		hasCurrent: true,
		mode:       playbackStation,
		health:     player.Health{AudioCodec: "aac"},
		actionText: "Favorite station",
		volume:     70,
	}

	line := m.consoleHeaderLine(62)
	if got := lipgloss.Width(line); got != 62 {
		t.Fatalf("header width = %d, want 62", got)
	}
	if !strings.Contains(line, "ON AIR") {
		t.Fatalf("header does not contain playback state: %q", line)
	}
	if !strings.Contains(line, "Favorite station") {
		t.Fatalf("header does not contain last action: %q", line)
	}
}
