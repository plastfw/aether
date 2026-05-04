package doctor

import (
	"fmt"
	"os"
	"os/exec"

	"aether/internal/clipboard"
	"aether/internal/favorites"
	"path/filepath"
	"strings"

	"aether/internal/config"
	"aether/internal/history"
	"aether/internal/settings"
)

type Check struct {
	Name    string
	OK      bool
	Message string
}

func Run() []Check {
	checks := []Check{
		checkBinary("mpv", true),
		checkBinary("notify-send", false),
		checkClipboard(),
	}
	checks = append(checks, checkConfig()...)
	checks = append(checks, checkSettings()...)
	checks = append(checks, checkHistory()...)
	checks = append(checks, checkFavorites()...)
	return checks
}

func Print(checks []Check) int {
	exitCode := 0
	fmt.Println("aether doctor")
	fmt.Println()
	for _, check := range checks {
		mark := "✓"
		if !check.OK {
			mark = "✗"
			exitCode = 1
		}
		fmt.Printf("%s %s — %s\n", mark, check.Name, check.Message)
	}
	return exitCode
}

func checkClipboard() Check {
	if clipboard.Available() {
		return Check{Name: "clipboard", OK: true, Message: "wl-copy/xclip/xsel"}
	}
	return Check{Name: "clipboard", OK: true, Message: "not found (optional, install wl-clipboard)"}
}

func checkBinary(name string, required bool) Check {
	path, err := exec.LookPath(name)
	if err == nil {
		return Check{Name: name, OK: true, Message: path}
	}
	msg := "not found"
	if required {
		msg += " (required)"
	} else {
		msg += " (optional)"
	}
	return Check{Name: name, OK: !required, Message: msg}
}

func checkConfig() []Check {
	path, err := config.DefaultPath()
	if err != nil {
		return []Check{{Name: "config path", OK: false, Message: err.Error()}}
	}

	checks := []Check{{Name: "config path", OK: true, Message: path}}
	if err := config.EnsureDefault(path); err != nil {
		return append(checks, Check{Name: "config create", OK: false, Message: err.Error()})
	}

	cfg, err := config.Load(path)
	if err != nil {
		return append(checks, Check{Name: "config load", OK: false, Message: err.Error()})
	}
	checks = append(checks, Check{Name: "stations", OK: len(cfg.Stations) > 0, Message: fmt.Sprintf("%d loaded", len(cfg.Stations))})
	return checks
}

func checkSettings() []Check {
	path, err := settings.DefaultPath()
	if err != nil {
		return []Check{{Name: "app config path", OK: false, Message: err.Error()}}
	}
	if err := settings.EnsureDefault(path); err != nil {
		return []Check{{Name: "app config create", OK: false, Message: err.Error()}}
	}
	cfg, err := settings.Load(path)
	if err != nil {
		return []Check{{Name: "app config load", OK: false, Message: err.Error()}}
	}
	return []Check{{Name: "app config", OK: true, Message: fmt.Sprintf("%s (volume:%d search:%d)", path, cfg.Player.Volume, cfg.Search.PageSize)}}
}

func checkHistory() []Check {
	path, err := history.DefaultPath()
	if err != nil {
		return []Check{{Name: "history path", OK: false, Message: err.Error()}}
	}
	return checkWritablePath("history path", path)
}

func checkFavorites() []Check {
	path, err := favorites.DefaultPath()
	if err != nil {
		return []Check{{Name: "favorites path", OK: false, Message: err.Error()}}
	}
	return checkWritablePath("favorites path", path)
}

func checkWritablePath(name, path string) []Check {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return []Check{{Name: name, OK: false, Message: err.Error()}}
	}

	probe := filepath.Join(dir, ".write-test")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return []Check{{Name: name, OK: false, Message: err.Error()}}
	}
	_ = os.Remove(probe)

	message := path
	if strings.Contains(path, ".cache") {
		message += " (warning: cache path)"
	}
	return []Check{{Name: name, OK: true, Message: message}}
}
