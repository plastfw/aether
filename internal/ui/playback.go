package ui

import (
	"time"

	"aether/internal/config"
	"aether/internal/metadata"
	"aether/internal/player"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) toggleStationFavorite() (Model, tea.Cmd) {
	if len(m.stations) == 0 || m.cursor < 0 || m.cursor >= len(m.stations) {
		return m, nil
	}
	url := m.stations[m.cursor].URL
	cfg, station, favorite, err := config.ToggleFavorite(m.configPath, url)
	if err != nil {
		m.status = errStyle.Render("station favorite: " + err.Error())
		return m, nil
	}
	m.stations = cfg.Stations
	m.cursor = stationIndex(m.stations, station.URL)
	if m.hasCurrent {
		m.playing = stationIndex(m.stations, m.current.URL)
	}
	if favorite {
		m.status = "★ Станция в избранном: " + station.Name
	} else {
		m.status = "☆ Станция убрана из избранного: " + station.Name
	}
	return m, nil
}

func (m Model) startStation(index int) (Model, tea.Cmd) {
	if index < 0 || index >= len(m.stations) {
		return m, nil
	}
	st := m.stations[index]
	if err := m.player.Play(st.URL); err != nil {
		m.status = errStyle.Render(err.Error())
		m.playing = -1
		return m, nil
	}
	m.playing = index
	m.mode = playbackStation
	m.paused = false
	m.health = player.Health{}
	m.current = st
	m.hasCurrent = true
	m.track = metadata.Metadata{}
	m.lastNotifyID = ""
	m.playID++
	if volume, err := m.player.Volume(); err == nil {
		m.volume = volume
	}
	m.status = "Играет: " + st.Name
	return m, tea.Batch(metadataTick(m.player, st, m.playID), playbackHealthTick(m.player, st, m.playID))
}

func (m Model) togglePause() (Model, tea.Cmd) {
	if !m.hasCurrent || m.mode == playbackStopped {
		m.status = "Сначала запусти станцию"
		return m, nil
	}
	paused, err := m.player.TogglePause()
	if err != nil {
		m.status = errStyle.Render("pause: " + err.Error())
		return m, nil
	}
	m.paused = paused
	if paused {
		m.status = "Пауза: " + m.current.Name
		return m, nil
	}
	m.status = "Продолжает играть: " + m.current.Name
	return m, playbackHealthTick(m.player, m.current, m.playID)
}

func playbackHealthTick(player playbackHealthReader, station config.Station, playID int) tea.Cmd {
	return tea.Tick(1800*time.Millisecond, func(time.Time) tea.Msg {
		return playbackHealthMsg{playID: playID, station: station, health: player.Health()}
	})
}

type playbackHealthReader interface {
	Health() player.Health
}
