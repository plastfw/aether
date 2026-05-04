package ui

import (
	"fmt"
	"strings"

	"aether/internal/config"
	"aether/internal/metadata"
	"aether/internal/player"
	"aether/internal/radiobrowser"
	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSearch(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	switch {
	case matchSearchCommand(key, m.settings.Keys.Search.Close):
		m.search = searchState{}
		m.status = "Search закрыт"
		return m, nil
	case key == "backspace":
		if len(m.search.query) > 0 && !m.search.loading {
			runes := []rune(m.search.query)
			m.search.query = string(runes[:len(runes)-1])
			m.search.results = nil
			m.search.cursor = 0
			m.search.searched = false
		}
	case matchSearchCommand(key, m.settings.Keys.Search.Preview):
		if len(m.search.results) > 0 {
			return m.previewSelectedSearchResult()
		}
		return m.runSearch()
	case matchSearchCommand(key, m.settings.Keys.Search.Add):
		return m.addSelectedSearchResult()
	case matchSearchCommand(key, m.settings.Keys.Search.Up):
		if m.search.cursor > 0 {
			m.search.cursor--
		}
	case matchSearchCommand(key, m.settings.Keys.Search.Down):
		if m.search.cursor < len(m.search.results)-1 {
			m.search.cursor++
		}
	case matchSearchCommand(key, m.settings.Keys.Search.NextPage):
		return m.nextSearchPage()
	case matchSearchCommand(key, m.settings.Keys.Search.PrevPage):
		return m.prevSearchPage()
	default:
		if len(msg.Runes) > 0 && !m.search.loading {
			for _, r := range msg.Runes {
				m.search.query += string(r)
			}
			m.search.results = nil
			m.search.cursor = 0
			m.search.searched = false
		}
	}
	return m, nil
}

func (m Model) runSearch() (Model, tea.Cmd) {
	if strings.TrimSpace(m.search.query) == "" || m.search.loading {
		return m, nil
	}
	m.search.loading = true
	m.search.err = ""
	m.search.page = max(0, m.search.page)
	return m, m.searchCmd(m.search.query, m.search.page)
}

func (m Model) nextSearchPage() (Model, tea.Cmd) {
	if !m.search.loading && strings.TrimSpace(m.search.query) != "" {
		m.search.page++
		m.search.loading = true
		m.search.err = ""
		return m, m.searchCmd(m.search.query, m.search.page)
	}
	return m, nil
}

func (m Model) prevSearchPage() (Model, tea.Cmd) {
	if !m.search.loading && m.search.page > 0 {
		m.search.page--
		m.search.loading = true
		m.search.err = ""
		return m, m.searchCmd(m.search.query, m.search.page)
	}
	return m, nil
}

func (m Model) selectedSearchStation() (config.Station, bool) {
	if len(m.search.results) == 0 || m.search.cursor < 0 || m.search.cursor >= len(m.search.results) {
		return config.Station{}, false
	}
	result := m.search.results[m.search.cursor]
	return config.Station{Name: result.Name, URL: result.StreamURL(), Provider: "generic"}, true
}

func (m Model) previewSelectedSearchResult() (Model, tea.Cmd) {
	station, ok := m.selectedSearchStation()
	if !ok {
		return m, nil
	}
	if err := m.player.Play(station.URL); err != nil {
		m.search.err = err.Error()
		return m, nil
	}
	m.playing = stationIndex(m.stations, station.URL)
	m.mode = playbackPreview
	m.paused = false
	m.health = player.Health{}
	m.current = station
	m.hasCurrent = true
	m.track = metadata.Metadata{}
	m.lastNotifyID = ""
	m.playID++
	if volume, err := m.player.Volume(); err == nil {
		m.volume = volume
	}
	m.status = "Предпрослушивание: " + station.Name
	return m, tea.Batch(metadataTick(m.player, station, m.playID), playbackHealthTick(m.player, station, m.playID))
}

func (m Model) addSelectedSearchResult() (Model, tea.Cmd) {
	station, ok := m.selectedSearchStation()
	if !ok {
		return m, nil
	}
	added, err := config.AppendStation(m.configPath, station)
	if err != nil {
		m.search.err = err.Error()
		return m, nil
	}
	if !added {
		idx := stationIndex(m.stations, station.URL)
		if idx >= 0 {
			m.cursor = idx
			if m.hasCurrent && m.current.URL == station.URL {
				m.playing = idx
				m.mode = playbackStation
				m.current = m.stations[idx]
				m.status = "Уже добавлена и играет: " + m.current.Name
				m.search = searchState{}
				return m, nil
			}
		}
		m.status = "Станция уже есть: " + station.Name
		m.search.err = "already added"
		return m, nil
	}
	m.stations = append(m.stations, station)
	m.cursor = len(m.stations) - 1
	if m.hasCurrent && m.current.URL == station.URL {
		m.playing = m.cursor
		m.mode = playbackStation
		m.current = station
		m.status = "Добавлена и играет: " + station.Name
	} else {
		m.status = "Добавлена станция: " + station.Name
	}
	m.search = searchState{}
	return m, nil
}

func (m Model) searchCmd(query string, page int) tea.Cmd {
	pageSize := m.settings.Search.PageSize
	return func() tea.Msg {
		results, err := radiobrowser.Search(query, pageSize, page*pageSize)
		return searchResultsMsg{results: results, err: err}
	}
}

func (m Model) searchView() string {
	layout := computeLayout(m.width, m.height)
	if layout.mode == layoutTooSmall {
		return tooSmallView(m.width, m.height)
	}
	contentW := max(1, layout.width-2)
	return m.consolePageView(contentW, layout.height, func(width, height int) []string {
		content := m.searchContent(width-4, height-3)
		return consoleBox(m.tabTitle(settings.UITabSearch), width, height, content)
	})
}

func (m Model) searchContent(width, height int) []string {
	query := m.search.query + "█"
	status := consoleDimStyle.Render(fmt.Sprintf("page %d", m.search.page+1))
	if m.search.loading {
		status = consoleAmberStyle.Render("searching...")
	}

	header := []string{
		fitLine(status, width),
		fitLine(consoleDimStyle.Render("query ")+consoleTitleStyle.Render(query), width),
		"",
	}
	if m.search.err != "" {
		header = append(header, errStyle.Render(m.search.err))
	}
	if m.search.searched && len(m.search.results) == 0 && m.search.err == "" {
		header = append(header, consoleDimStyle.Render("nothing found"))
	}

	hints := m.searchHintLines(width)
	resultsHeight := max(0, height-len(header)-len(hints))
	resultLines := m.searchResultLines(width, resultsHeight)

	lines := append([]string{}, header...)
	lines = append(lines, resultLines...)
	hintStart := max(0, height-len(hints))
	lines = fitLines(lines, hintStart, width)
	for _, hint := range hints {
		lines = append(lines, centerLine(hint, width))
	}
	return fitLines(lines, height, width)
}

func (m Model) searchResultLines(width, height int) []string {
	lines := []string{}
	maxResults := max(0, height/2)
	for i, result := range m.search.results {
		if i >= maxResults {
			break
		}
		prefix := fmt.Sprintf("%d. ", i+1)
		line := prefix + truncate(result.Name, width-6)
		if i == m.search.cursor {
			lines = append(lines, selStyle.Render("▶ "+line))
		} else {
			lines = append(lines, itemStyle.Render("  "+line))
		}
		details := fmt.Sprintf("%s • %dk • votes:%d", compactTags(result.Tags), result.Bitrate, result.Votes)
		lines = append(lines, consoleDimStyle.Render("   "+truncate(details, width-4)))
	}
	return fitLines(lines, height, width)
}

func (m Model) searchHintLines(width int) []string {
	return []string{
		renderKeyHintText(
			keyHint(m.searchMoveKeys(), "moves"),
			keyHint(m.settings.Keys.Search.Preview, "search/preview"),
			keyHint(m.settings.Keys.Search.Add, "add"),
			keyHint(m.settings.Keys.Search.Close, "back"),
		),
	}
}

func (m Model) searchMoveKeys() []string {
	keys := []string{}
	keys = append(keys, m.settings.Keys.Search.Up...)
	keys = append(keys, m.settings.Keys.Search.Down...)
	keys = append(keys, m.settings.Keys.Search.PrevPage...)
	keys = append(keys, m.settings.Keys.Search.NextPage...)
	return keys
}

func (m Model) searchPanel(width int) string {
	return strings.Join(m.searchContent(width-4, max(8, m.height-6)), "\n")
}

func matchSearchCommand(key string, values []string) bool {
	if len([]rune(key)) == 1 {
		return false
	}
	return settings.Match(key, values)
}
