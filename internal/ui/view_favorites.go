package ui

import (
	"fmt"

	"aether/internal/clipboard"
	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateFavoriteTracks(msg tea.KeyMsg) (Model, tea.Cmd) {
	entries := m.favoriteItems
	key := msg.String()
	keys := m.settings.Keys.FavoriteTracks
	if m.confirmClear {
		switch {
		case settings.Match(key, keys.Confirm):
			if err := m.favorites.Clear(); err != nil {
				m.status = errStyle.Render("favorites clear: " + err.Error())
			} else {
				m.favoriteItems = nil
				m.favCursor = 0
				m.status = "Все favorite tracks удалены"
			}
			m.confirmClear = false
		case settings.Match(key, keys.Cancel):
			m.confirmClear = false
			m.status = "Удаление отменено"
		}
		return m, nil
	}
	switch {
	case settings.Match(key, keys.Close):
		m.view = viewMain
		m.status = "Главный экран"
	case settings.Match(key, keys.Up):
		if m.favCursor > 0 {
			m.favCursor--
		}
	case settings.Match(key, keys.Down):
		if m.favCursor < len(entries)-1 {
			m.favCursor++
		}
	case settings.Match(key, keys.Delete):
		if len(entries) > 0 {
			removed := entries[m.favCursor].Display()
			if err := m.favorites.Remove(m.favCursor); err != nil {
				m.status = errStyle.Render("favorite remove: " + err.Error())
			} else {
				m.favoriteItems = loadFavoriteItems(m.favorites)
				if m.favCursor > 0 && m.favCursor >= len(entries)-1 {
					m.favCursor--
				}
				m.status = "Удалено: " + removed
			}
		}
	case settings.Match(key, keys.Clear):
		if len(entries) > 0 {
			m.confirmClear = true
			m.status = "Удалить все favorite tracks?"
		}
	case settings.Match(key, keys.Copy):
		if len(entries) > 0 {
			if err := clipboard.Copy(entries[m.favCursor].Display()); err != nil {
				m.status = errStyle.Render("clipboard: " + err.Error())
			} else {
				m.status = "Скопировано: " + entries[m.favCursor].Display()
			}
		}
	}
	return m, nil
}

func (m Model) favoriteTracksView() string {
	layout := computeLayout(m.width, m.height)
	if layout.mode == layoutTooSmall {
		return tooSmallView(m.width, m.height)
	}
	contentW := max(1, layout.width-2)
	return m.consolePageView(contentW, layout.height, func(width, height int) []string {
		content := m.favoriteTracksContent(width-4, height-3)
		return consoleBox(m.tabTitle(settings.UITabTracks), width, height, content)
	})
}

func (m Model) favoriteTracksContent(width, height int) []string {
	entries := m.favoriteItems
	header := []string{}
	if m.confirmClear {
		header = append(header, errStyle.Render("Confirm clear all favorite tracks"))
	}
	hints := m.favoriteTracksHintLines(width)
	listHeight := max(0, height-len(header)-len(hints))
	list := m.favoriteTrackLines(width, listHeight)

	lines := append([]string{}, header...)
	lines = append(lines, list...)
	if len(entries) == 0 && len(header) == 0 {
		lines = []string{consoleDimStyle.Render("favorite tracks is empty")}
	}
	hintStart := max(0, height-len(hints))
	lines = fitLines(lines, hintStart, width)
	for _, hint := range hints {
		lines = append(lines, centerLine(hint, width))
	}
	return fitLines(lines, height, width)
}

func (m Model) favoriteTrackLines(width, height int) []string {
	entries := m.favoriteItems
	if len(entries) == 0 {
		return nil
	}
	visible := max(1, height)
	start := 0
	if m.favCursor >= visible {
		start = m.favCursor - visible + 1
	}
	end := min(len(entries), start+visible)
	lines := []string{}
	for i := start; i < end; i++ {
		prefix := "  "
		if i == m.favCursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%3d. %s", prefix, i+1, entries[i].Display())
		if i == m.favCursor {
			lines = append(lines, selStyle.Render(truncate(line, width)))
		} else {
			lines = append(lines, itemStyle.Render(truncate(line, width)))
		}
	}
	return fitLines(lines, height, width)
}

func (m Model) favoriteTracksHintLines(width int) []string {
	keys := m.settings.Keys.FavoriteTracks
	if m.confirmClear {
		return []string{renderKeyHintText(
			keyHint(keys.Confirm, "confirm"),
			keyHint(keys.Cancel, "cancel"),
		)}
	}
	return []string{renderKeyHintText(
		keyHint(append(keys.Up, keys.Down...), "moves"),
		keyHint(keys.Copy, "copy"),
		keyHint(keys.Delete, "delete"),
		keyHint(keys.Clear, "clear all"),
		keyHint(keys.Close, "back"),
	)}
}
