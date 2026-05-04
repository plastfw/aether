package ui

import (
	"fmt"
	"time"

	"aether/internal/config"
	"aether/internal/favorites"
	"aether/internal/history"
	"aether/internal/metadata"
	"aether/internal/notify"
	"aether/internal/player"
	"aether/internal/radiobrowser"
	"aether/internal/settings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewMode int

const (
	viewMain viewMode = iota
	viewFavoriteTracks
	viewHelp
)

type playbackMode int

const (
	playbackStopped playbackMode = iota
	playbackStation
	playbackPreview
)

type Model struct {
	stations        []config.Station
	cursor          int
	playing         int
	mode            playbackMode
	paused          bool
	health          player.Health
	volume          int
	track           metadata.Metadata
	current         config.Station
	hasCurrent      bool
	status          string
	width           int
	height          int
	player          *player.Player
	history         *history.Recorder
	favorites       *favorites.Store
	historyItems    []history.Entry
	favoriteItems   []favorites.Entry
	lastNotifyID    string
	playID          int
	configPath      string
	settings        settings.Settings
	search          searchState
	view            viewMode
	channelsOpen    bool
	renamingStation bool
	renameInput     string
	deletingStation bool
	favCursor       int
	confirmClear    bool
	actionText      string
	actionSeq       int
	animFrame       int
}

type searchState struct {
	open     bool
	query    string
	results  []radiobrowser.Station
	cursor   int
	page     int
	loading  bool
	err      string
	searched bool
}

type metadataMsg struct {
	playID   int
	station  config.Station
	metadata metadata.Metadata
	err      error
}

type searchResultsMsg struct {
	results []radiobrowser.Station
	err     error
}

type clearActionMsg struct {
	seq int
}

type animationMsg struct{}

type playbackHealthMsg struct {
	playID  int
	station config.Station
	health  player.Health
}

var (
	accent     = lipgloss.Color("5")
	accent2    = lipgloss.Color("6")
	green      = lipgloss.Color("2")
	yellow     = lipgloss.Color("3")
	red        = lipgloss.Color("1")
	mutedColor = lipgloss.Color("8")
	textColor  = lipgloss.Color("7")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	itemStyle  = lipgloss.NewStyle().PaddingLeft(1).Foreground(textColor)
	selStyle   = lipgloss.NewStyle().Foreground(green).Bold(true)
	helpStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	errStyle   = lipgloss.NewStyle().Foreground(red)
	mutedStyle = lipgloss.NewStyle().Foreground(yellow)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent2).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 2)

	controlsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 2)

	panelTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	labelStyle      = lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle      = lipgloss.NewStyle().Foreground(green).Bold(true)
	artistStyle     = lipgloss.NewStyle().Foreground(green).Bold(true)
)

func New(stations []config.Station, configPath string, appSettings settings.Settings) Model {
	historyRecorder, _ := newHistoryRecorder()
	favoritesStore, _ := newFavoritesStore()
	appSettings.Normalize()
	historyItems := loadHistoryItems(appSettings.History.MaxEntries)
	favoriteItems := loadFavoriteItems(favoritesStore)
	return Model{
		stations:      stations,
		playing:       -1,
		volume:        appSettings.Player.Volume,
		status:        "Выбери станцию и нажми Enter",
		width:         100,
		height:        30,
		player:        player.New(appSettings.Player.Volume, appSettings.Player.MaxVolume),
		history:       historyRecorder,
		favorites:     favoritesStore,
		historyItems:  historyItems,
		favoriteItems: favoriteItems,
		configPath:    configPath,
		settings:      appSettings,
	}
}

func (m Model) Init() tea.Cmd { return animationTick() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animationMsg:
		m.animFrame++
		return m, animationTick()
	case clearActionMsg:
		if msg.seq == m.actionSeq {
			m.actionText = ""
		}
		return m, nil
	case playbackHealthMsg:
		if !m.hasCurrent || msg.playID != m.playID || msg.station.URL != m.current.URL || m.mode == playbackStopped || m.paused {
			return m, nil
		}
		m.health = msg.health
		if msg.health.AudioCodec == "" {
			if msg.health.IdleActive || msg.health.CoreIdle {
				m.status = "Поток не стартовал: " + msg.station.Name
			} else {
				m.status = "Буферизация / пока нет audio: " + msg.station.Name
			}
		}
		return m, playbackHealthTick(m.player, m.current, m.playID)
	case searchResultsMsg:
		m.search.loading = false
		m.search.searched = true
		if msg.err != nil {
			m.search.err = msg.err.Error()
			m.search.results = nil
			return m, nil
		}
		m.search.err = ""
		m.search.results = msg.results
		m.search.cursor = 0
		return m, nil
	case metadataMsg:
		if !m.hasCurrent || msg.playID != m.playID || msg.station.URL != m.current.URL {
			return m, nil
		}
		if msg.err == nil {
			m.track = msg.metadata
			_ = m.history.Record(m.current, msg.metadata)
			m.historyItems = loadHistoryItems(m.settings.History.MaxEntries)
			id := metadataID(msg.metadata)
			if m.settings.Notifications.Enabled && id != "" && id != m.lastNotifyID {
				_ = notify.Track(m.current.Name, msg.metadata)
				m.lastNotifyID = id
			}
		}
		return m, metadataTick(m.player, m.current, m.playID)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		key := msg.String()
		action := m.actionForKey(key)
		if action != "" {
			m = m.withAction(action)
		}
		finishKey := func(model Model, cmd tea.Cmd) (tea.Model, tea.Cmd) {
			if action == "" {
				return model, cmd
			}
			return model, tea.Batch(cmd, model.clearActionCmd())
		}
		if updated, ok := m.activateTabKey(key); ok {
			return finishKey(updated, nil)
		}
		if m.view == viewHelp {
			updated, cmd := m.updateHelp(msg)
			return finishKey(updated, cmd)
		}
		if m.view == viewFavoriteTracks {
			updated, cmd := m.updateFavoriteTracks(msg)
			return finishKey(updated, cmd)
		}
		if m.search.open {
			updated, cmd := m.updateSearch(msg)
			return finishKey(updated, cmd)
		}
		if m.renamingStation {
			updated, cmd := m.updateRenameStation(msg)
			return finishKey(updated, cmd)
		}
		if m.deletingStation {
			updated, cmd := m.updateDeleteStation(msg)
			return finishKey(updated, cmd)
		}
		switch {
		case settings.Match(key, m.settings.Keys.Global.Search):
			m.search = searchState{open: true}
			m.status = "Search: введи запрос и нажми Enter"
		case settings.Match(key, m.settings.Keys.Global.Quit):
			_ = m.player.Stop()
			return m, tea.Quit
		case settings.Match(key, m.settings.Keys.Global.Channels):
			m.channelsOpen = !m.channelsOpen
			if m.channelsOpen {
				m.status = "Channels открыты"
			} else {
				m.status = "Channels закрыты"
			}
		case settings.Match(key, m.settings.Keys.Global.FavoriteTracks):
			m.favoriteItems = loadFavoriteItems(m.favorites)
			m.view = viewFavoriteTracks
			m.confirmClear = false
			m.status = "Favorite tracks"
		case settings.Match(key, m.settings.Keys.Global.FavoriteStation):
			updated, cmd := m.toggleStationFavorite()
			return finishKey(updated, cmd)
		case settings.Match(key, m.settings.Keys.Global.RenameStation):
			updated, cmd := m.startRenameStation()
			return finishKey(updated, cmd)
		case settings.Match(key, m.settings.Keys.Global.DeleteStation):
			updated, cmd := m.startDeleteStation()
			return finishKey(updated, cmd)
		case settings.Match(key, m.settings.Keys.Global.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case settings.Match(key, m.settings.Keys.Global.Down):
			if m.cursor < len(m.stations)-1 {
				m.cursor++
			}
		case settings.Match(key, m.settings.Keys.Global.Play):
			if len(m.stations) == 0 {
				m.status = "Нет станций в конфиге"
				return finishKey(m, nil)
			}
			m.channelsOpen = false
			updated, cmd := m.startStation(m.cursor)
			return finishKey(updated, cmd)
		case settings.Match(key, m.settings.Keys.Global.Pause):
			updated, cmd := m.togglePause()
			return finishKey(updated, cmd)
		case settings.Match(key, m.settings.Keys.Global.VolumeUp):
			volume, err := m.player.AddVolume(m.settings.Player.VolumeStep)
			if err != nil {
				m.status = errStyle.Render("volume: " + err.Error())
			} else {
				m.volume = volume
				m.status = fmt.Sprintf("Громкость: %d%%", m.volume)
			}
		case settings.Match(key, m.settings.Keys.Global.VolumeDown):
			volume, err := m.player.AddVolume(-m.settings.Player.VolumeStep)
			if err != nil {
				m.status = errStyle.Render("volume: " + err.Error())
			} else {
				m.volume = volume
				m.status = fmt.Sprintf("Громкость: %d%%", m.volume)
			}
		case settings.Match(key, m.settings.Keys.Global.FavoriteTrack):
			if !m.hasCurrent || m.mode == playbackStopped {
				m.status = "Сначала запусти станцию"
				return finishKey(m, nil)
			}
			added, err := m.favorites.Add(m.current, m.track)
			if err != nil {
				m.status = errStyle.Render("favorite: " + err.Error())
			} else if added {
				m.favoriteItems = loadFavoriteItems(m.favorites)
				m.status = "Добавлено в favorites: " + m.track.Display()
			} else {
				m.status = "Уже в favorites или нет metadata"
			}
		}
		return finishKey(m, nil)
	}
	return m, nil
}

func (m Model) withAction(action string) Model {
	m.actionText = action
	m.actionSeq++
	return m
}

func (m Model) clearActionCmd() tea.Cmd {
	seq := m.actionSeq
	return tea.Tick(1250*time.Millisecond, func(time.Time) tea.Msg {
		return clearActionMsg{seq: seq}
	})
}

func animationTick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg {
		return animationMsg{}
	})
}
