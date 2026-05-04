package ui

import (
	"strings"

	"aether/internal/config"
	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startRenameStation() (Model, tea.Cmd) {
	if !m.channelsOpen {
		m.status = "Открой список станций через Tab"
		return m, nil
	}
	if len(m.stations) == 0 || m.cursor < 0 || m.cursor >= len(m.stations) {
		m.status = "Нет станции для переименования"
		return m, nil
	}
	m.renamingStation = true
	m.renameInput = m.stations[m.cursor].Name
	m.status = "Rename station"
	return m, nil
}

func (m Model) updateRenameStation(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	switch {
	case settings.Match(key, m.settings.Keys.Global.Quit):
		m.renamingStation = false
		m.renameInput = ""
		m.status = "Переименование отменено"
		return m, nil
	case settings.Match(key, m.settings.Keys.Global.Play):
		return m.saveRenamedStation()
	case key == "backspace":
		if len(m.renameInput) > 0 {
			runes := []rune(m.renameInput)
			m.renameInput = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				m.renameInput += string(r)
			}
		}
	}
	return m, nil
}

func (m Model) saveRenamedStation() (Model, tea.Cmd) {
	if len(m.stations) == 0 || m.cursor < 0 || m.cursor >= len(m.stations) {
		m.renamingStation = false
		m.renameInput = ""
		m.status = "Нет станции для переименования"
		return m, nil
	}
	station := m.stations[m.cursor]
	newName := strings.TrimSpace(m.renameInput)
	cfg, err := config.RenameStation(m.configPath, station.URL, newName)
	if err != nil {
		m.status = errStyle.Render("rename: " + err.Error())
		return m, nil
	}
	m.stations = cfg.Stations
	m.cursor = stationIndex(m.stations, station.URL)
	if m.hasCurrent && m.current.URL == station.URL {
		m.current = m.stations[m.cursor]
	}
	if m.hasCurrent {
		m.playing = stationIndex(m.stations, m.current.URL)
	}
	m.renamingStation = false
	m.renameInput = ""
	m.status = "Станция переименована: " + newName
	return m, nil
}
