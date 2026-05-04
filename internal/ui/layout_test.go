package ui

import (
	"strings"
	"testing"

	"aether/internal/config"
	"aether/internal/settings"

	"github.com/charmbracelet/lipgloss"
)

func TestMinimumLayoutRendersInsideBounds(t *testing.T) {
	m := New([]config.Station{{Name: "Station", URL: "https://example.com", Provider: "generic"}}, "", settings.Default())
	m.width = minTerminalWidth
	m.height = minTerminalHeight

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) != minTerminalHeight {
		t.Fatalf("line count = %d, want %d", len(lines), minTerminalHeight)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got > minTerminalWidth {
			t.Fatalf("line %d width = %d, want <= %d: %q", i, got, minTerminalWidth, line)
		}
	}
}

func TestTooSmallLayoutWarning(t *testing.T) {
	if computeLayout(minTerminalWidth-1, minTerminalHeight).mode != layoutTooSmall {
		t.Fatal("expected too small layout below minimum width")
	}
	if computeLayout(minTerminalWidth, minTerminalHeight-1).mode != layoutTooSmall {
		t.Fatal("expected too small layout below minimum height")
	}
	if computeLayout(minTerminalWidth, minTerminalHeight).mode == layoutTooSmall {
		t.Fatal("expected minimum size to be renderable")
	}
}
