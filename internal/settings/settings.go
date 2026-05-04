package settings

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Settings struct {
	Player        Player        `toml:"player"`
	Search        Search        `toml:"search"`
	History       History       `toml:"history"`
	Notifications Notifications `toml:"notifications"`
	UI            UI            `toml:"ui"`
	Keys          Keys          `toml:"keys"`
}

type Player struct {
	Volume     int `toml:"volume"`
	MaxVolume  int `toml:"max_volume"`
	VolumeStep int `toml:"volume_step"`
}

type Search struct {
	PageSize int `toml:"page_size"`
}

type History struct {
	MaxEntries int `toml:"max_entries"`
}

type Notifications struct {
	Enabled bool `toml:"enabled"`
}

type UI struct {
	Tabs       []UITab    `toml:"tabs"`
	KeyAliases []KeyAlias `toml:"key_aliases"`
}

// KeyAlias lets the UI render several equivalent physical keys as one compact label.
type KeyAlias struct {
	Keys  []string `toml:"keys"`
	Label string   `toml:"label"`
}

type UITab struct {
	ID      string   `toml:"id"`
	Label   string   `toml:"label"`
	Keys    []string `toml:"keys"`
	Enabled bool     `toml:"enabled"`
}

const (
	UITabRadio  = "radio"
	UITabSearch = "search"
	UITabTracks = "tracks"
	UITabHelp   = "help"
)

type Keys struct {
	Global         GlobalKeys         `toml:"global"`
	Search         SearchKeys         `toml:"search"`
	FavoriteTracks FavoriteTracksKeys `toml:"favorite_tracks"`
}

type GlobalKeys struct {
	Quit            []string `toml:"quit"`
	Search          []string `toml:"search"`
	Channels        []string `toml:"channels"`
	Play            []string `toml:"play"`
	Pause           []string `toml:"pause"`
	FavoriteTrack   []string `toml:"favorite_track"`
	FavoriteStation []string `toml:"favorite_station"`
	FavoriteTracks  []string `toml:"favorite_tracks"`
	RenameStation   []string `toml:"rename_station"`
	DeleteStation   []string `toml:"delete_station"`
	VolumeUp        []string `toml:"volume_up"`
	VolumeDown      []string `toml:"volume_down"`
	Up              []string `toml:"up"`
	Down            []string `toml:"down"`
}

type SearchKeys struct {
	Close    []string `toml:"close"`
	Preview  []string `toml:"preview"`
	Up       []string `toml:"up"`
	Down     []string `toml:"down"`
	Add      []string `toml:"add"`
	NextPage []string `toml:"next_page"`
	PrevPage []string `toml:"prev_page"`
}

type FavoriteTracksKeys struct {
	Close   []string `toml:"close"`
	Up      []string `toml:"up"`
	Down    []string `toml:"down"`
	Delete  []string `toml:"delete"`
	Clear   []string `toml:"clear"`
	Copy    []string `toml:"copy"`
	Confirm []string `toml:"confirm"`
	Cancel  []string `toml:"cancel"`
}

func Default() Settings {
	return Settings{
		Player:        Player{Volume: 70, MaxVolume: 200, VolumeStep: 5},
		Search:        Search{PageSize: 8},
		History:       History{MaxEntries: 200},
		Notifications: Notifications{Enabled: true},
		UI: UI{
			KeyAliases: []KeyAlias{
				{Keys: []string{"-", "_"}, Label: "-"},
				{Keys: []string{"+", "="}, Label: "+"},
			},
			Tabs: []UITab{
				{ID: UITabRadio, Label: "Radio", Keys: []string{"1"}, Enabled: true},
				{ID: UITabSearch, Label: "Search", Keys: []string{"2"}, Enabled: true},
				{ID: UITabTracks, Label: "Tracks", Keys: []string{"3"}, Enabled: true},
				{ID: UITabHelp, Label: "Help", Keys: []string{"4"}, Enabled: true},
			},
		},
		Keys: Keys{
			Global: GlobalKeys{
				Quit: []string{"ctrl+c"}, Search: []string{"/"}, Channels: []string{"tab"}, Play: []string{"enter"}, Pause: []string{" "},
				FavoriteTrack: []string{"B"}, FavoriteStation: []string{"F"}, FavoriteTracks: []string{"T"}, RenameStation: []string{"R"}, DeleteStation: []string{"D"},
				VolumeUp: []string{"+", "="}, VolumeDown: []string{"-", "_"}, Up: []string{"up"}, Down: []string{"down"},
			},
			Search: SearchKeys{
				Close: []string{"ctrl+c"}, Preview: []string{"enter"}, Up: []string{"up"}, Down: []string{"down"}, Add: []string{"tab"},
				NextPage: []string{"right"}, PrevPage: []string{"left"},
			},
			FavoriteTracks: FavoriteTracksKeys{
				Close: []string{"ctrl+c"}, Up: []string{"up"}, Down: []string{"down"}, Delete: []string{"delete"}, Clear: []string{"ctrl+d"},
				Copy: []string{"enter"}, Confirm: []string{"enter"}, Cancel: []string{"ctrl+c"},
			},
		},
	}
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "aether", "config.toml"), nil
}

func EnsureDefault(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultConfig), 0o644)
}

func Load(path string) (Settings, error) {
	cfg := Default()
	_, err := toml.DecodeFile(path, &cfg)
	cfg.Normalize()
	return cfg, err
}

func (s *Settings) Normalize() {
	if s.Player.MaxVolume <= 0 {
		s.Player.MaxVolume = 200
	}
	if s.Player.MaxVolume < 100 {
		s.Player.MaxVolume = 100
	}
	if s.Player.MaxVolume > 1000 {
		s.Player.MaxVolume = 1000
	}
	if s.Player.Volume < 0 || s.Player.Volume > s.Player.MaxVolume {
		s.Player.Volume = min(70, s.Player.MaxVolume)
	}
	if s.Player.VolumeStep <= 0 {
		s.Player.VolumeStep = 5
	}
	if s.Search.PageSize <= 0 || s.Search.PageSize > 50 {
		s.Search.PageSize = 8
	}
	if s.History.MaxEntries <= 0 {
		s.History.MaxEntries = 200
	}
	if s.History.MaxEntries > 100000 {
		s.History.MaxEntries = 100000
	}
	s.normalizeUITabs()
	s.normalizeKeyAliases()
	s.normalizeKeys()
}

func (s *Settings) normalizeUITabs() {
	defaults := Default().UI.Tabs
	if len(s.UI.Tabs) == 0 {
		s.UI.Tabs = defaults
		return
	}
	seen := map[string]bool{}
	for i := range s.UI.Tabs {
		if s.UI.Tabs[i].ID == "" {
			continue
		}
		seen[s.UI.Tabs[i].ID] = true
		for _, def := range defaults {
			if s.UI.Tabs[i].ID != def.ID {
				continue
			}
			if s.UI.Tabs[i].Label == "" {
				s.UI.Tabs[i].Label = def.Label
			}
			if len(s.UI.Tabs[i].Keys) == 0 {
				s.UI.Tabs[i].Keys = def.Keys
			}
			break
		}
	}
	for _, def := range defaults {
		if !seen[def.ID] {
			s.UI.Tabs = append(s.UI.Tabs, def)
		}
	}
}

func (s *Settings) normalizeKeyAliases() {
	if len(s.UI.KeyAliases) == 0 {
		s.UI.KeyAliases = Default().UI.KeyAliases
	}
}

func (s *Settings) normalizeKeys() {
	defaults := Default().Keys.Global
	if len(s.Keys.Global.Pause) == 0 {
		s.Keys.Global.Pause = defaults.Pause
	}
}

func (s Settings) Tab(id string) (UITab, bool) {
	for _, tab := range s.UI.Tabs {
		if tab.ID == id && tab.Enabled {
			return tab, true
		}
	}
	return UITab{}, false
}

func (t UITab) DisplayLabel() string {
	if len(t.Keys) == 0 {
		return t.Label
	}
	return t.Keys[0] + " " + t.Label
}

func Match(key string, values []string) bool {
	for _, value := range values {
		if key == value {
			return true
		}
	}
	return false
}

const defaultConfig = `[player]
volume = 70
max_volume = 200
volume_step = 5

[search]
page_size = 8

[history]
max_entries = 200

[notifications]
enabled = true

[ui]

[[ui.key_aliases]]
keys = ["-", "_"]
label = "-"

[[ui.key_aliases]]
keys = ["+", "="]
label = "+"

[keys.global]
quit = ["ctrl+c"]
search = ["/"]
channels = ["tab"]
play = ["enter"]
pause = [" "]
favorite_track = ["B"]
favorite_station = ["F"]
favorite_tracks = ["T"]
rename_station = ["R"]
delete_station = ["D"]
volume_up = ["+", "="]
volume_down = ["-", "_"]
up = ["up"]
down = ["down"]

[[ui.tabs]]
id = "radio"
label = "Radio"
keys = ["1"]
enabled = true

[[ui.tabs]]
id = "search"
label = "Search"
keys = ["2"]
enabled = true

[[ui.tabs]]
id = "tracks"
label = "Tracks"
keys = ["3"]
enabled = true

[[ui.tabs]]
id = "help"
label = "Help"
keys = ["4"]
enabled = true

[keys.search]
close = ["ctrl+c"]
preview = ["enter"]
up = ["up"]
down = ["down"]
add = ["tab"]
next_page = ["right"]
prev_page = ["left"]

[keys.favorite_tracks]
close = ["ctrl+c"]
up = ["up"]
down = ["down"]
delete = ["delete"]
clear = ["ctrl+d"]
copy = ["enter"]
confirm = ["enter"]
cancel = ["ctrl+c"]
`
