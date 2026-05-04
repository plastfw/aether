package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Station struct {
	Name     string `toml:"name"`
	URL      string `toml:"url"`
	Provider string `toml:"provider"`
	Favorite bool   `toml:"favorite,omitempty"`
}

func (s Station) MetadataProvider() string {
	if s.Provider == "" {
		return "generic"
	}
	return s.Provider
}

type Config struct {
	Stations []Station `toml:"station"`
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "aether", "stations.toml"), nil
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

func Load(path string) (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	SortStations(cfg.Stations)
	return cfg, err
}

func Save(path string, cfg Config) error {
	SortStations(cfg.Stations)
	if err := backup(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, station := range cfg.Stations {
		fmt.Fprintf(&b, "[[station]]\nname = %q\nurl = %q\nprovider = %q\n", station.Name, station.URL, station.MetadataProvider())
		if station.Favorite {
			b.WriteString("favorite = true\n")
		}
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func SortStations(stations []Station) {
	// Stable bubble-free insertion keeps user order inside favorite/non-favorite groups.
	favorites := make([]Station, 0, len(stations))
	regular := make([]Station, 0, len(stations))
	for _, station := range stations {
		if station.Favorite {
			favorites = append(favorites, station)
		} else {
			regular = append(regular, station)
		}
	}
	copy(stations, append(favorites, regular...))
}

func ToggleFavorite(path string, url string) (Config, Station, bool, error) {
	cfg, err := Load(path)
	if err != nil {
		return cfg, Station{}, false, err
	}
	var updated Station
	found := false
	for i := range cfg.Stations {
		if strings.TrimSpace(cfg.Stations[i].URL) == strings.TrimSpace(url) {
			cfg.Stations[i].Favorite = !cfg.Stations[i].Favorite
			updated = cfg.Stations[i]
			found = true
			break
		}
	}
	if !found {
		return cfg, Station{}, false, fmt.Errorf("station not found")
	}
	if err := Save(path, cfg); err != nil {
		return cfg, updated, updated.Favorite, err
	}
	cfg, err = Load(path)
	return cfg, updated, updated.Favorite, err
}

func AppendStation(path string, station Station) (bool, error) {
	cfg, err := Load(path)
	if err != nil {
		return false, err
	}
	for _, existing := range cfg.Stations {
		if strings.TrimSpace(existing.URL) == strings.TrimSpace(station.URL) {
			return false, nil
		}
	}

	if err := backup(path); err != nil {
		return false, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "\n[[station]]\nname = %q\nurl = %q\nprovider = %q\n", station.Name, station.URL, station.MetadataProvider())
	if err == nil && station.Favorite {
		_, err = fmt.Fprintln(file, "favorite = true")
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func backup(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	backupDir := filepath.Join(filepath.Dir(path), "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", filepath.Base(path), time.Now().Format("20060102-150405.000000000")))
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return err
	}
	return cleanupBackups(path, 3)
}

func cleanupBackups(path string, keep int) error {
	if keep <= 0 {
		keep = 3
	}
	pattern := filepath.Join(filepath.Dir(path), "backups", filepath.Base(path)+".*.bak")
	backups, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(backups) <= keep {
		return nil
	}
	for _, backupPath := range backups[:len(backups)-keep] {
		if err := os.Remove(backupPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func DeleteStation(path string, url string) (Config, Station, error) {
	cfg, err := Load(path)
	if err != nil {
		return cfg, Station{}, err
	}

	url = strings.TrimSpace(url)
	for i := range cfg.Stations {
		if strings.TrimSpace(cfg.Stations[i].URL) != url {
			continue
		}
		removed := cfg.Stations[i]
		cfg.Stations = append(cfg.Stations[:i], cfg.Stations[i+1:]...)
		if err := Save(path, cfg); err != nil {
			return cfg, removed, err
		}
		loaded, err := Load(path)
		return loaded, removed, err
	}
	return cfg, Station{}, fmt.Errorf("station not found")
}

func RenameStation(path string, url string, newName string) (Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return cfg, err
	}

	url = strings.TrimSpace(url)
	newName = strings.TrimSpace(newName)

	if newName == "" {
		return cfg, fmt.Errorf("station name is empty")
	}

	for i := range cfg.Stations {
		if strings.TrimSpace(cfg.Stations[i].URL) == url {
			cfg.Stations[i].Name = newName

			if err := Save(path, cfg); err != nil {
				return cfg, err
			}

			return Load(path)
		}
	}
	return cfg, fmt.Errorf("station not found")
}

const defaultConfig = `[[station]]
name = "BadRadio"
url = "https://s2.radio.co/s2b2b68744/listen"
provider = "generic"

[[station]]
name = "Lofi Radio"
url = "https://play.streamafrica.net/lofiradio"
provider = "generic"

[[station]]
name = "Radio Paradise Main Mix FLAC"
url = "https://stream.radioparadise.com/flac"
provider = "generic"
`
