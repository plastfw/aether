package ui

import "aether/internal/settings"

func (m Model) actionForKey(key string) string {
	if action := m.tabActionForKey(key); action != "" {
		return action
	}
	if m.view == viewHelp {
		return m.helpActionForKey(key)
	}
	if m.search.open {
		return m.searchActionForKey(key)
	}
	if m.view == viewFavoriteTracks {
		return m.favoriteTracksActionForKey(key)
	}
	if m.renamingStation {
		return m.renameActionForKey(key)
	}
	if m.deletingStation {
		return m.deleteStationActionForKey(key)
	}
	return m.globalActionForKey(key)
}

func (m Model) tabActionForKey(key string) string {
	for _, tab := range m.settings.UI.Tabs {
		if tab.Enabled && settings.Match(key, tab.Keys) {
			return "Tab: " + tab.Label
		}
	}
	return ""
}

func (m Model) globalActionForKey(key string) string {
	keys := m.settings.Keys.Global
	switch {
	case settings.Match(key, keys.Search):
		return "Search"
	case settings.Match(key, keys.Channels):
		return "Channels"
	case settings.Match(key, keys.Play) && len(m.stations) > 0:
		return "Play"
	case settings.Match(key, keys.Pause) && m.hasCurrent && m.mode != playbackStopped:
		return "Pause"
	case settings.Match(key, keys.FavoriteTrack) && m.hasCurrent && m.mode != playbackStopped:
		return "Favorite track"
	case settings.Match(key, keys.FavoriteStation) && len(m.stations) > 0:
		return "Favorite station"
	case settings.Match(key, keys.FavoriteTracks):
		return "Tracks"
	case settings.Match(key, keys.RenameStation) && m.channelsOpen && len(m.stations) > 0:
		return "Rename station"
	case settings.Match(key, keys.DeleteStation) && m.channelsOpen && len(m.stations) > 0:
		return "Delete station"
	case settings.Match(key, keys.VolumeUp) && m.hasCurrent:
		return "Volume up"
	case settings.Match(key, keys.VolumeDown) && m.hasCurrent:
		return "Volume down"
	case settings.Match(key, keys.Up) && m.cursor > 0:
		return "Move up"
	case settings.Match(key, keys.Down) && m.cursor < len(m.stations)-1:
		return "Move down"
	case settings.Match(key, keys.Quit):
		return "Quit"
	default:
		return ""
	}
}

func (m Model) searchActionForKey(key string) string {
	keys := m.settings.Keys.Search
	switch {
	case settings.Match(key, keys.Close):
		return "Close search"
	case settings.Match(key, keys.Preview) && len(m.search.results) > 0:
		return "Preview station"
	case settings.Match(key, keys.Preview) && m.search.query != "" && !m.search.loading:
		return "Run search"
	case settings.Match(key, keys.Add) && len(m.search.results) > 0:
		return "Add station"
	case settings.Match(key, keys.Up) && m.search.cursor > 0:
		return "Move up"
	case settings.Match(key, keys.Down) && m.search.cursor < len(m.search.results)-1:
		return "Move down"
	case settings.Match(key, keys.NextPage) && m.search.query != "" && !m.search.loading:
		return "Next page"
	case settings.Match(key, keys.PrevPage) && m.search.page > 0 && !m.search.loading:
		return "Previous page"
	default:
		return ""
	}
}

func (m Model) favoriteTracksActionForKey(key string) string {
	keys := m.settings.Keys.FavoriteTracks
	if m.confirmClear {
		switch {
		case settings.Match(key, keys.Confirm):
			return "Confirm"
		case settings.Match(key, keys.Cancel):
			return "Cancel"
		default:
			return ""
		}
	}
	switch {
	case settings.Match(key, keys.Close):
		return "Close tracks"
	case settings.Match(key, keys.Up) && m.favCursor > 0:
		return "Move up"
	case settings.Match(key, keys.Down) && m.favCursor < len(m.favoriteItems)-1:
		return "Move down"
	case settings.Match(key, keys.Copy) && len(m.favoriteItems) > 0:
		return "Copy track"
	case settings.Match(key, keys.Delete) && len(m.favoriteItems) > 0:
		return "Delete track"
	case settings.Match(key, keys.Clear) && len(m.favoriteItems) > 0:
		return "Clear tracks"
	default:
		return ""
	}
}

func (m Model) helpActionForKey(key string) string {
	if settings.Match(key, m.settings.Keys.Global.Quit) {
		return "Close help"
	}
	return ""
}

func (m Model) renameActionForKey(key string) string {
	switch {
	case settings.Match(key, m.settings.Keys.Global.Play):
		return "Save rename"
	case settings.Match(key, m.settings.Keys.Global.Quit):
		return "Cancel rename"
	default:
		return ""
	}
}

func (m Model) deleteStationActionForKey(key string) string {
	switch {
	case settings.Match(key, m.settings.Keys.Global.Play):
		return "Confirm delete"
	case settings.Match(key, m.settings.Keys.Global.Quit):
		return "Cancel delete"
	default:
		return ""
	}
}
