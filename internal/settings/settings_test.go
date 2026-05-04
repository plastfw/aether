package settings

import "testing"

func TestNormalize(t *testing.T) {
	cfg := Settings{}
	cfg.Player.Volume = 999
	cfg.Search.PageSize = -1
	cfg.Normalize()
	if cfg.Player.MaxVolume != 200 {
		t.Fatalf("max volume = %d, want 200", cfg.Player.MaxVolume)
	}
	if cfg.Player.Volume != 70 {
		t.Fatalf("volume = %d, want 70", cfg.Player.Volume)
	}
	if cfg.Player.VolumeStep != 5 {
		t.Fatalf("volume step = %d, want 5", cfg.Player.VolumeStep)
	}
	if cfg.Search.PageSize != 8 {
		t.Fatalf("page size = %d, want 8", cfg.Search.PageSize)
	}
	if cfg.History.MaxEntries != 200 {
		t.Fatalf("history max entries = %d, want 200", cfg.History.MaxEntries)
	}
	if !Match(" ", cfg.Keys.Global.Pause) {
		t.Fatalf("pause key = %v, want space", cfg.Keys.Global.Pause)
	}
}

func TestDefaultHotkeys(t *testing.T) {
	cfg := Default()

	checks := []struct {
		name   string
		key    string
		values []string
	}{
		{name: "global quit", key: "ctrl+c", values: cfg.Keys.Global.Quit},
		{name: "global search", key: "/", values: cfg.Keys.Global.Search},
		{name: "global channels", key: "tab", values: cfg.Keys.Global.Channels},
		{name: "global play", key: "enter", values: cfg.Keys.Global.Play},
		{name: "global pause", key: " ", values: cfg.Keys.Global.Pause},
		{name: "global favorite track", key: "B", values: cfg.Keys.Global.FavoriteTrack},
		{name: "global favorite station", key: "F", values: cfg.Keys.Global.FavoriteStation},
		{name: "global favorite tracks", key: "T", values: cfg.Keys.Global.FavoriteTracks},
		{name: "global rename station", key: "R", values: cfg.Keys.Global.RenameStation},
		{name: "global delete station", key: "D", values: cfg.Keys.Global.DeleteStation},
		{name: "global volume up", key: "+", values: cfg.Keys.Global.VolumeUp},
		{name: "global volume down", key: "-", values: cfg.Keys.Global.VolumeDown},
		{name: "global up", key: "up", values: cfg.Keys.Global.Up},
		{name: "global down", key: "down", values: cfg.Keys.Global.Down},
		{name: "search close", key: "ctrl+c", values: cfg.Keys.Search.Close},
		{name: "search preview", key: "enter", values: cfg.Keys.Search.Preview},
		{name: "search up", key: "up", values: cfg.Keys.Search.Up},
		{name: "search down", key: "down", values: cfg.Keys.Search.Down},
		{name: "search add", key: "tab", values: cfg.Keys.Search.Add},
		{name: "search next page", key: "right", values: cfg.Keys.Search.NextPage},
		{name: "search prev page", key: "left", values: cfg.Keys.Search.PrevPage},
		{name: "favorite tracks close", key: "ctrl+c", values: cfg.Keys.FavoriteTracks.Close},
		{name: "favorite tracks up", key: "up", values: cfg.Keys.FavoriteTracks.Up},
		{name: "favorite tracks down", key: "down", values: cfg.Keys.FavoriteTracks.Down},
		{name: "favorite tracks delete", key: "delete", values: cfg.Keys.FavoriteTracks.Delete},
		{name: "favorite tracks clear", key: "ctrl+d", values: cfg.Keys.FavoriteTracks.Clear},
		{name: "favorite tracks copy", key: "enter", values: cfg.Keys.FavoriteTracks.Copy},
		{name: "favorite tracks confirm", key: "enter", values: cfg.Keys.FavoriteTracks.Confirm},
		{name: "favorite tracks cancel", key: "ctrl+c", values: cfg.Keys.FavoriteTracks.Cancel},
	}

	for _, check := range checks {
		if !Match(check.key, check.values) {
			t.Fatalf("%s: expected %q in %v", check.name, check.key, check.values)
		}
	}
}

func TestDefaultTabs(t *testing.T) {
	cfg := Default()

	checks := []struct {
		id    string
		key   string
		label string
	}{
		{id: UITabRadio, key: "1", label: "Radio"},
		{id: UITabSearch, key: "2", label: "Search"},
		{id: UITabTracks, key: "3", label: "Tracks"},
		{id: UITabHelp, key: "4", label: "Help"},
	}

	for _, check := range checks {
		tab, ok := cfg.Tab(check.id)
		if !ok {
			t.Fatalf("tab %q not found", check.id)
		}
		if tab.Label != check.label {
			t.Fatalf("tab %q label = %q, want %q", check.id, tab.Label, check.label)
		}
		if !Match(check.key, tab.Keys) {
			t.Fatalf("tab %q expected key %q in %v", check.id, check.key, tab.Keys)
		}
	}
}
