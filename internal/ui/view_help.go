package ui

import (
	"strings"

	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateHelp(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	switch {
	case settings.Match(key, m.settings.Keys.Global.Quit):
		m.view = viewMain
		m.status = "Главный экран"
	}
	return m, nil
}

func (m Model) helpView() string {
	layout := computeLayout(m.width, m.height)
	if layout.mode == layoutTooSmall {
		return tooSmallView(m.width, m.height)
	}
	contentW := max(1, layout.width-2)

	return m.consolePageView(contentW, layout.height, func(width, height int) []string {
		moveKeys := append([]string{}, m.settings.Keys.Global.Up...)
		moveKeys = append(moveKeys, m.settings.Keys.Global.Down...)
		volumeKeys := append([]string{}, m.settings.Keys.Global.VolumeDown...)
		volumeKeys = append(volumeKeys, m.settings.Keys.Global.VolumeUp...)
		content := centeredHelpContent(width, height-3, []string{
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.Channels, "open / close channels dropdown")),
			m.renderKeyHelpLine(width, keyHint(moveKeys, "move selection")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.Play, "play selected channel")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.Pause, "pause / resume")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.FavoriteTrack, "favorite current track")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.FavoriteStation, "favorite station")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.RenameStation, "rename selected station")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.DeleteStation, "delete selected station")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.FavoriteTracks, "favorite tracks")),
			m.renderKeyHelpLine(width, keyHint(volumeKeys, "volume control")),
			m.renderKeyHelpLine(width, keyHint(m.settings.Keys.Global.Quit, "back / cancel / quit")),
		})
		return consoleBox(m.tabTitle(settings.UITabHelp), width, height, content)
	})
}

func centeredHelpContent(width, height int, lines []string) []string {
	contentWidth := 0
	for _, line := range lines {
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	leftPad := max(0, (width-contentWidth)/2)
	verticalPad := max(0, (height-len(lines))/2)
	out := make([]string, 0, len(lines)+verticalPad)
	for i := 0; i < verticalPad; i++ {
		out = append(out, "")
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, strings.Repeat(" ", leftPad)+line)
	}
	return out
}
