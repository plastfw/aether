package ui

import (
	"strings"

	"aether/internal/settings"
)

func (m Model) activateTabKey(key string) (Model, bool) {
	for _, tab := range m.settings.UI.Tabs {
		if !tab.Enabled || !settings.Match(key, tab.Keys) {
			continue
		}
		return m.activateTab(tab.ID), true
	}
	return m, false
}

func (m Model) activateTab(id string) Model {
	switch id {
	case settings.UITabRadio:
		m.view = viewMain
		m.search = searchState{}
		m.confirmClear = false
		m.status = m.tabStatus(id)
	case settings.UITabSearch:
		m.view = viewMain
		m.search = searchState{open: true}
		m.confirmClear = false
		m.status = m.tabStatus(id)
	case settings.UITabTracks:
		m.search = searchState{}
		m.favoriteItems = loadFavoriteItems(m.favorites)
		m.view = viewFavoriteTracks
		m.confirmClear = false
		m.status = m.tabStatus(id)
	case settings.UITabHelp:
		m.search = searchState{}
		m.view = viewHelp
		m.confirmClear = false
		m.status = m.tabStatus(id)
	}
	return m
}

func (m Model) tabStatus(id string) string {
	if tab, ok := m.settings.Tab(id); ok {
		return tab.Label
	}
	return id
}

func (m Model) tabTitle(id string) string {
	return strings.ToUpper(m.tabStatus(id))
}

func (m Model) isActiveTab(id string) bool {
	switch id {
	case settings.UITabRadio:
		return m.view == viewMain && !m.search.open
	case settings.UITabSearch:
		return m.search.open
	case settings.UITabTracks:
		return m.view == viewFavoriteTracks
	case settings.UITabHelp:
		return m.view == viewHelp
	default:
		return false
	}
}
