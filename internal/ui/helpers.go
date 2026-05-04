package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aether/internal/config"
	"aether/internal/favorites"
	"aether/internal/history"
	"aether/internal/metadata"
	"aether/internal/player"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func metadataTick(p *player.Player, station config.Station, playID int) tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		provider := metadataProvider(p, station)
		metadata, err := provider.Fetch(context.Background(), station)
		return metadataMsg{playID: playID, station: station, metadata: metadata, err: err}
	})
}

func metadataProvider(p *player.Player, station config.Station) metadata.Provider {
	switch station.MetadataProvider() {
	case "generic", "mpv", "":
		return metadata.MPVProvider{Player: p}
	default:
		return metadata.MPVProvider{Player: p}
	}
}

func newHistoryRecorder() (*history.Recorder, error) {
	path, err := history.DefaultPath()
	if err != nil {
		return nil, err
	}
	return history.NewRecorder(path)
}

func newFavoritesStore() (*favorites.Store, error) {
	path, err := favorites.DefaultPath()
	if err != nil {
		return nil, err
	}
	return favorites.NewStore(path)
}

func loadHistoryItems(limit int) []history.Entry {
	path, err := history.DefaultPath()
	if err != nil {
		return nil
	}
	entries, err := history.ReadLast(path, limit)
	if err != nil {
		return nil
	}
	return entries
}

func loadFavoriteItems(store *favorites.Store) []favorites.Entry {
	if store == nil {
		return nil
	}
	entries, err := store.List()
	if err != nil {
		return nil
	}
	return entries
}

func (m Model) visibleHistory(limit int) []history.Entry {
	if limit <= 0 || len(m.historyItems) == 0 {
		return nil
	}
	if len(m.historyItems) <= limit {
		return m.historyItems
	}
	return m.historyItems[len(m.historyItems)-limit:]
}

func (m Model) airLogEntries(limit int) []history.Entry {
	entries := m.visibleHistory(limit)
	if !m.hasCurrent || m.mode == playbackStopped || !hasCurrentTrackMetadata(m.track) {
		return entries
	}
	current := history.Entry{
		Station:  m.current.Name,
		URL:      m.current.URL,
		Provider: m.current.MetadataProvider(),
		Artist:   m.track.Artist,
		Title:    m.track.Title,
		Album:    m.track.Album,
		Raw:      m.track.Raw,
		Source:   m.track.Source,
	}
	for _, entry := range entries {
		if m.isCurrentTrackEntry(entry) {
			return entries
		}
	}
	entries = append(append([]history.Entry{}, entries...), current)
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	return entries
}

func (m Model) isCurrentTrackEntry(entry history.Entry) bool {
	if !hasCurrentTrackMetadata(m.track) {
		return false
	}
	return normalizeTrackKey(entry.Artist, entry.Title, entry.Raw) == normalizeTrackKey(m.track.Artist, m.track.Title, m.track.Raw)
}

func hasCurrentTrackMetadata(md metadata.Metadata) bool {
	return normalizeTrackKey(md.Artist, md.Title, md.Raw) != ""
}

func normalizeTrackKey(artist, title, raw string) string {
	key := strings.ToLower(strings.TrimSpace(artist + "|" + title + "|" + raw))
	switch key {
	case "", "||", "—", "-", "unknown", "unknown|unknown|":
		return ""
	default:
		return key
	}
}

func metadataID(md metadata.Metadata) string {
	id := strings.ToLower(strings.TrimSpace(md.Artist + "|" + md.Title + "|" + md.Raw))
	switch id {
	case "", "||", "—", "-|", "unknown":
		return ""
	default:
		return id
	}
}

func labelValue(label, value string) string {
	return labelStyle.Render(label+": ") + valueStyle.Render(value)
}

func isSearchResultNumber(key string) bool {
	runes := []rune(key)
	return len(runes) == 1 && runes[0] >= '1' && runes[0] <= '8'
}

func (m Model) currentTrackText() string {
	if m.track.Artist != "" && m.track.Title != "" {
		return m.track.Artist + " — " + m.track.Title
	}
	if m.track.Title != "" {
		return m.track.Title
	}
	if m.track.Raw != "" {
		return m.track.Raw
	}
	return "—"
}

func (m Model) currentStationName() string {
	if m.hasCurrent {
		return m.current.Name
	}
	if len(m.stations) > 0 && m.cursor >= 0 && m.cursor < len(m.stations) {
		return m.stations[m.cursor].Name
	}
	return "—"
}

func trimLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func cleanEmpty(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "—" {
		return "—"
	}
	return value
}

func historyLine(entry history.Entry, width int) string {
	track := trackText(entry.Artist, entry.Title, entry.Raw)
	line := entry.Time.Format("15:04") + "  " + track
	return helpStyle.Render(truncate(line, width))
}

func trackText(artist, title, raw string) string {
	if artist != "" && title != "" {
		return artist + " — " + title
	}
	if title != "" {
		return title
	}
	if raw != "" {
		return raw
	}
	return "—"
}

func stationIndex(stations []config.Station, url string) int {
	for i, station := range stations {
		if station.URL == url {
			return i
		}
	}
	return -1
}

func signalBars(m Model) string {
	if !m.hasCurrent || m.mode == playbackStopped {
		return helpStyle.Render("▁▁▁▁")
	}
	return valueStyle.Render("▂▄▆█")
}

func stateBadge(m Model) string {
	if !m.hasCurrent || m.mode == playbackStopped {
		return lipgloss.NewStyle().Bold(true).Foreground(mutedColor).Render(" STOPPED ")
	}
	if m.mode == playbackPreview {
		return lipgloss.NewStyle().Bold(true).Foreground(yellow).Render(" PREVIEW ")
	}
	return lipgloss.NewStyle().Bold(true).Foreground(green).Render(" PLAYING ")
}

func volumeBar(volume, width int) string {
	filled := volume * width / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return valueStyle.Render(strings.Repeat("█", filled)) + helpStyle.Render(strings.Repeat("░", width-filled))
}

func thinVolumeBar(volume, width int) string {
	filled := volume * width / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return valueStyle.Render(strings.Repeat("━", filled)) + helpStyle.Render(strings.Repeat("─", width-filled))
}

func truncate(s string, width int) string {
	if width <= 1 {
		return "…"
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func nowPlayingLine(m Model) string {
	if !m.hasCurrent || m.mode == playbackStopped {
		return helpStyle.Render("Статус: stopped")
	}
	return helpStyle.Render(fmt.Sprintf("Статус: playing • volume: %d%%", m.volume))
}

func compactTags(tags string) string {
	if strings.TrimSpace(tags) == "" {
		return "no tags"
	}
	parts := strings.Split(tags, ",")
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, ", ")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
