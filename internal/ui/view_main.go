package ui

import (
	"fmt"
	"strings"

	"aether/internal/history"
	"aether/internal/settings"

	"github.com/charmbracelet/lipgloss"
)

var (
	consoleFrameStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(accent2)
	consoleBoxStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8"))
	consoleRailStyle  = lipgloss.NewStyle().Foreground(accent2)
	consoleDimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	consoleHotStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	consoleRedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	consoleAmberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	consoleTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
)

func (m Model) View() string {
	layout := computeLayout(m.width, m.height)
	if layout.mode == layoutTooSmall {
		return tooSmallView(m.width, m.height)
	}
	if m.view == viewFavoriteTracks {
		return m.favoriteTracksView()
	}
	if m.view == viewHelp {
		return m.helpView()
	}
	if m.search.open {
		return m.searchView()
	}
	return m.consoleView(layout)
}

func (m Model) consoleView(layout layoutSpec) string {
	innerW := max(1, layout.width)
	innerH := max(1, layout.height)
	contentW := max(1, innerW-2)
	return m.consolePageView(contentW, innerH, m.consoleBody)
}

func (m Model) consolePageView(width, height int, body func(int, int) []string) string {
	contentH := max(1, height)
	bodyH := max(1, contentH-2)
	lines := make([]string, 0, contentH)
	lines = append(lines, m.consoleHeaderLine(width))
	lines = append(lines, consoleDimStyle.Render(strings.Repeat("─", width)))
	lines = append(lines, body(width, bodyH)...)
	lines = fitLines(lines, min(contentH, len(lines)), width)
	return renderRail(lines, width)
}

func (m Model) consoleHeaderLine(width int) string {
	left := consoleTitleStyle.Render("AETHER") + "  " + m.tabsLine()
	rightParts := []string{}
	if m.actionText != "" {
		rightParts = append(rightParts, consoleTitleStyle.Render(m.actionText))
	}
	rightParts = append(rightParts,
		m.playbackStatusStyle().Render(m.playbackStatusLabel()),
		consoleDimStyle.Render("VOL ")+consoleTitleStyle.Render(fmt.Sprintf("%d%%", m.volume)),
	)
	right := strings.Join(rightParts, consoleDimStyle.Render("  •  "))
	return alignHeader(left, right, width)
}

func (m Model) consoleBody(width, height int) []string {
	if height <= 0 {
		return nil
	}
	if m.channelsOpen {
		actionH := 0
		if m.renamingStation || m.deletingStation {
			actionH = 5
		}
		maxChannelsH := max(5, height-actionH-4)
		channelsH := min(maxChannelsH, min(12, len(m.stations)+3))
		logH := max(4, height-channelsH-actionH)
		lines := m.consoleChannelsBox(width, channelsH)
		if m.renamingStation {
			lines = append(lines, m.renameStationBox(width, actionH)...)
		}
		if m.deletingStation {
			lines = append(lines, m.deleteStationBox(width, actionH)...)
		}
		lines = append(lines, m.consoleLogBox(width, logH)...)
		return fitLines(lines, min(height, len(lines)), width)
	}

	lines := []string{m.consoleChannelSummary(width)}
	logH := max(4, height-1)
	lines = append(lines, m.consoleLogBox(width, logH)...)
	return fitLines(lines, min(height, len(lines)), width)
}

func (m Model) consoleChannelsBox(width, height int) []string {
	visible := max(1, height-3)
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(len(m.stations), start+visible)
	content := []string{}
	if len(m.stations) == 0 {
		content = append(content, errStyle.Render("no stations configured"))
	} else {
		for i := start; i < end; i++ {
			st := m.stations[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "> "
			}
			marker := consoleDimStyle.Render("■")
			if i == m.playing && m.mode == playbackStation {
				marker = valueStyle.Render("■")
			} else if i == m.playing && m.mode == playbackPreview {
				marker = consoleAmberStyle.Render("◇")
			}
			star := " "
			if st.Favorite {
				star = "★"
			}
			line := fmt.Sprintf("%s%02d %s %s %s", prefix, i+1, marker, star, st.Name)
			if i == m.cursor {
				content = append(content, selStyle.Render(truncate(line, width-4)))
			} else {
				content = append(content, itemStyle.Render(truncate(line, width-4)))
			}
		}
	}
	return consoleBox("CHANNELS", width, height, content)
}

func (m Model) renameStationBox(width, height int) []string {
	content := []string{
		fitLine(consoleDimStyle.Render("name ")+consoleTitleStyle.Render(m.renameInput+"█"), width-4),
		centerLine(renderKeyHintText(keyHint(m.settings.Keys.Global.Play, "save"), keyHint(m.settings.Keys.Global.Quit, "cancel")), width-4),
	}
	return consoleBox("RENAME STATION", width, height, content)
}

func (m Model) deleteStationBox(width, height int) []string {
	station := ""
	if len(m.stations) > 0 && m.cursor >= 0 && m.cursor < len(m.stations) {
		station = m.stations[m.cursor].Name
	}
	content := []string{
		fitLine(consoleRedStyle.Render("delete ")+consoleTitleStyle.Render(station), width-4),
		centerLine(renderKeyHintText(keyHint(m.settings.Keys.Global.Play, "confirm"), keyHint(m.settings.Keys.Global.Quit, "cancel")), width-4),
	}
	return consoleBox("DELETE STATION", width, height, content)
}

func (m Model) consoleLogBox(width, height int) []string {
	entries := m.airLogEntries(max(1, height-3))
	currentIndex := -1
	for i, entry := range entries {
		if m.isCurrentTrackEntry(entry) {
			currentIndex = i
		}
	}
	content := []string{}
	if len(entries) == 0 {
		content = append(content, consoleDimStyle.Render("air log is empty"))
	} else {
		for i, entry := range entries {
			line := historyLine(entry, width-4)
			if i == currentIndex {
				line = currentTrackLine(entry, width-4)
			}
			content = append(content, line)
		}
	}
	return consoleBox("AIR LOG", width, height, content)
}

func currentTrackLine(entry history.Entry, width int) string {
	prefix := consoleAmberStyle.Render("NOW PLAYING ") + consoleDimStyle.Render("— ")
	artist := strings.TrimSpace(entry.Artist)
	title := strings.TrimSpace(entry.Title)
	if artist != "" && title != "" {
		return fitLine(prefix+artistStyle.Render(artist)+consoleDimStyle.Render(" — ")+consoleTitleStyle.Render(title), width)
	}
	return fitLine(prefix+consoleTitleStyle.Render(trackText(entry.Artist, entry.Title, entry.Raw)), width)
}

func (m Model) playbackStatusText() string {
	if !m.hasCurrent || m.mode == playbackStopped {
		return "STANDBY"
	}
	if m.paused {
		return "PAUSED"
	}
	if m.health.IdleActive || m.health.CoreIdle {
		return "NO AUDIO"
	}
	if m.health.PausedForCache || (m.health.CacheBuffering > 0 && m.health.CacheBuffering < 100) {
		return "BUFFERING"
	}
	if m.health.AudioCodec == "" {
		return "TUNING"
	}
	if m.mode == playbackPreview {
		return "PREVIEW"
	}
	return "ON AIR"
}

func (m Model) playbackStatusLabel() string {
	status := m.playbackStatusText()
	switch status {
	case "ON AIR", "PREVIEW":
		frames := []string{"◉", "◎"}
		return frames[m.animFrame%len(frames)] + " " + status
	case "BUFFERING", "TUNING":
		frames := []string{"◐", "◓", "◑", "◒"}
		return frames[m.animFrame%len(frames)] + " " + status
	default:
		return status
	}
}

func (m Model) playbackStatusStyle() lipgloss.Style {
	switch m.playbackStatusText() {
	case "ON AIR", "PREVIEW":
		return valueStyle
	case "NO AUDIO":
		return consoleRedStyle
	case "BUFFERING", "TUNING", "PAUSED":
		return consoleAmberStyle
	default:
		return consoleDimStyle
	}
}

func (m Model) currentSource() string {
	if m.hasCurrent {
		return m.current.MetadataProvider()
	}
	if len(m.stations) > 0 && m.cursor >= 0 && m.cursor < len(m.stations) {
		return m.stations[m.cursor].MetadataProvider()
	}
	return "generic"
}

func tooSmallView(width, height int) string {
	message := strings.Join([]string{
		consoleTitleStyle.Render("AETHER"),
		consoleDimStyle.Render("signal window too small"),
		consoleDimStyle.Render("resize to tune in"),
	}, "\n")
	boxW := max(1, min(34, width))
	boxH := max(1, min(5, height))
	boxInnerW := max(1, boxW-2)
	boxInnerH := max(1, boxH-2)
	box := consoleFrameStyle.Width(boxInnerW).Height(boxInnerH).Render(lipgloss.Place(boxInnerW, boxInnerH, lipgloss.Center, lipgloss.Center, message))
	return lipgloss.Place(max(1, width), max(1, height), lipgloss.Center, lipgloss.Center, box)
}

func consoleBox(title string, width, height int, content []string) []string {
	innerW := max(1, width-2)
	innerH := max(1, height-2)
	titleLine := " " + consoleTitleStyle.Render(title) + " " + consoleDimStyle.Render(strings.Repeat("─", max(0, innerW-len(title)-3)))
	body := []string{fitLine(titleLine, innerW)}
	body = append(body, content...)
	body = fitLines(body, innerH, innerW)
	rendered := consoleBoxStyle.Width(innerW).Height(innerH).Render(strings.Join(body, "\n"))
	return strings.Split(rendered, "\n")
}

func bracketMetric(label, value string) string {
	return consoleDimStyle.Render("[") + " " + consoleDimStyle.Render(label+":") + " " + consoleTitleStyle.Render(value) + " " + consoleDimStyle.Render("]")
}

func labelValuePlain(label, value string, width int) string {
	prefix := consoleDimStyle.Render(label + "  ")
	return fitLine(prefix+consoleTitleStyle.Render(value), width)
}

func (m Model) consoleChannelSummary(width int) string {
	channel := "no station"
	if len(m.stations) > 0 && m.cursor >= 0 && m.cursor < len(m.stations) {
		channel = m.stations[m.cursor].Name
	}
	return fitLine(consoleTitleStyle.Render("▶ "+channel), width)
}

func renderRail(lines []string, width int) string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, consoleRailStyle.Render("▌ ")+fitLine(line, width))
	}
	return strings.Join(out, "\n")
}

func (m Model) tabsLine() string {
	labels := []string{}
	for _, tab := range m.settings.UI.Tabs {
		if !tab.Enabled {
			continue
		}
		labels = append(labels, m.tabLabel(tab.DisplayLabel(), m.isActiveTab(tab.ID)))
	}
	return strings.Join(labels, "")
}

func (m Model) tabLabel(label string, active bool) string {
	if active {
		return consoleTitleStyle.Render("╭ " + label + " ╮")
	}
	return consoleDimStyle.Render("┌ " + label + " ┐")
}

func fitLines(lines []string, height, width int) []string {
	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		out = append(out, fitLine(line, width))
	}
	return out
}

func fitLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = truncate(line, width)
	pad := width - lipgloss.Width(line)
	if pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	return line
}

func centerLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = truncate(line, width)
	pad := width - lipgloss.Width(line)
	if pad <= 0 {
		return line
	}
	left := pad / 2
	return strings.Repeat(" ", left) + line + strings.Repeat(" ", pad-left)
}

func alignHeader(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	rw := lipgloss.Width(right)
	if rw >= width {
		return fitLine(right, width)
	}
	leftW := max(1, width-rw-1)
	return fitLine(left, leftW) + " " + right
}

func (m Model) adaptiveView() string  { return m.consoleView(computeLayout(m.width, m.height)) }
func (m Model) wideView() string      { return m.consoleView(computeLayout(m.width, m.height)) }
func (m Model) mediumView() string    { return m.consoleView(computeLayout(m.width, m.height)) }
func (m Model) shortWideView() string { return m.consoleView(computeLayout(m.width, m.height)) }
func (m Model) narrowView() string    { return m.consoleView(computeLayout(m.width, m.height)) }
func (m Model) tinyView() string      { return tooSmallView(m.width, m.height) }
func (m Model) compactView() string   { return m.consoleView(computeLayout(m.width, m.height)) }

func (m Model) header(width int) string { return m.consoleHeaderLine(width) }
func (m Model) oneLineHelp(width int) string {
	return helpStyle.MaxWidth(width).Render(m.tabStatus(settings.UITabHelp))
}
func (m Model) footer(width int) string {
	return fitLine(consoleDimStyle.Render(m.tabStatus(settings.UITabHelp)), width)
}
func (m Model) stationsPanel(width, height int) string {
	return strings.Join(m.consoleChannelsBox(width, height), "\n")
}
func (m Model) receiverPanel(width, height int) string {
	return m.consoleHeaderLine(width)
}
func (m Model) historyPanel(width, height int) string {
	return strings.Join(m.consoleLogBox(width, height), "\n")
}
