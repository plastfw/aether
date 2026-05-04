package ui

import (
	"aether/internal/config"
	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startDeleteStation() (Model, tea.Cmd) {
	if !m.channelsOpen {
		m.status = "Открой список станций через Tab"
		return m, nil
	}
	if len(m.stations) == 0 || m.cursor < 0 || m.cursor >= len(m.stations) {
		m.status = "Нет станции для удаления"
		return m, nil
	}
	m.deletingStation = true
	m.status = "Удалить станцию? " + m.stations[m.cursor].Name
	return m, nil
}

func (m Model) updateDeleteStation(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	switch {
	case settings.Match(key, m.settings.Keys.Global.Quit):
		m.deletingStation = false
		m.status = "Удаление станции отменено"
	case settings.Match(key, m.settings.Keys.Global.Play):
		return m.confirmDeleteStation()
	}
	return m, nil
}

func (m Model) confirmDeleteStation() (Model, tea.Cmd) {
	if len(m.stations) == 0 || m.cursor < 0 || m.cursor >= len(m.stations) {
		m.deletingStation = false
		m.status = "Нет станции для удаления"
		return m, nil
	}

	station := m.stations[m.cursor]
	cfg, removed, err := config.DeleteStation(m.configPath, station.URL)
	if err != nil {
		m.status = errStyle.Render("delete station: " + err.Error())
		return m, nil
	}

	wasCurrent := m.hasCurrent && m.current.URL == removed.URL
	m.stations = cfg.Stations
	if m.cursor >= len(m.stations) {
		m.cursor = max(0, len(m.stations)-1)
	}
	m.deletingStation = false

	if wasCurrent {
		_ = m.player.Stop()
		m.playing = -1
		m.mode = playbackStopped
		m.current = config.Station{}
		m.hasCurrent = false
		m.status = "Удалена и остановлена: " + removed.Name
		return m, nil
	}
	if m.hasCurrent {
		m.playing = stationIndex(m.stations, m.current.URL)
	}
	m.status = "Удалена станция: " + removed.Name
	return m, nil
}
