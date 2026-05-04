package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"aether/internal/config"
	"aether/internal/doctor"
	"aether/internal/favorites"
	"aether/internal/history"
	"aether/internal/radiobrowser"
	"aether/internal/settings"
	"aether/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.0.0"

func main() {
	if len(os.Args) > 1 {
		runCommand(os.Args[1:])
		return
	}
	runTUI()
}

func runCommand(args []string) {
	switch args[0] {
	case "doctor":
		os.Exit(doctor.Print(doctor.Run()))
	case "config":
		if len(args) == 2 && args[1] == "path" {
			printPath(config.DefaultPath())
			return
		}
		if len(args) == 2 && args[1] == "app-path" {
			printPath(settings.DefaultPath())
			return
		}
		usageAndExit()
	case "history":
		runHistoryCommand(args[1:])
	case "favorites":
		runFavoritesCommand(args[1:])
	case "search":
		runSearchCommand(args[1:])
	case "version", "--version", "-v":
		fmt.Println("aether " + version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		usageAndExit()
	}
}

func runHistoryCommand(args []string) {
	if len(args) == 1 && args[0] == "path" {
		printPath(history.DefaultPath())
		return
	}
	if len(args) == 0 || args[0] == "list" || args[0] == "tail" {
		limit := 20
		if len(args) > 1 {
			parsed, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "invalid history limit:", args[1])
				os.Exit(2)
			}
			limit = parsed
		}
		path, err := history.DefaultPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		entries, err := history.ReadLast(path, limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, entry := range entries {
			fmt.Printf("%s  %s\n", entry.Time.Format("15:04"), trackLine(entry.Artist, entry.Title, entry.Raw))
		}
		return
	}
	usageAndExit()
}

func runFavoritesCommand(args []string) {
	if len(args) == 1 && args[0] == "path" {
		printPath(favorites.DefaultPath())
		return
	}
	if len(args) == 0 || args[0] == "list" {
		path, err := favorites.DefaultPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		entries, err := favorites.ReadAll(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for i, entry := range entries {
			fmt.Printf("%3d. %s\n", i+1, entry.Display())
		}
		return
	}
	usageAndExit()
}

func runSearchCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "search query is required")
		usageAndExit()
	}
	query := strings.Join(args, " ")
	stations, err := radiobrowser.Search(query, 12, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "search:", err)
		os.Exit(1)
	}
	for i, station := range stations {
		fmt.Printf("%2d. %s\n", i+1, station.Name)
		fmt.Printf("    %s • %dk • votes:%d • clicks:%d\n", compactTags(station.Tags), station.Bitrate, station.Votes, station.ClickCount)
		fmt.Printf("    %s\n", station.StreamURL())
	}
}

func runTUI() {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config path:", err)
		os.Exit(1)
	}

	if err := config.EnsureDefault(path); err != nil {
		fmt.Fprintln(os.Stderr, "create default config:", err)
		os.Exit(1)
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}

	settingsPath, err := settings.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "settings path:", err)
		os.Exit(1)
	}
	if err := settings.EnsureDefault(settingsPath); err != nil {
		fmt.Fprintln(os.Stderr, "create default settings:", err)
		os.Exit(1)
	}
	appSettings, err := settings.Load(settingsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load settings:", err)
		os.Exit(1)
	}

	historyPath, err := history.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "history path:", err)
		os.Exit(1)
	}
	if err := history.Trim(historyPath, appSettings.History.MaxEntries); err != nil {
		fmt.Fprintln(os.Stderr, "trim history:", err)
		os.Exit(1)
	}

	program := tea.NewProgram(ui.New(cfg.Stations, path, appSettings), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "run:", err)
		os.Exit(1)
	}
}

func printPath(path string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(path)
}

func usageAndExit() {
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Printf(`aether %s — terminal radio player

Usage:
  aether                    Start TUI
  aether doctor             Run diagnostics
  aether config path        Print stations config path
  aether config app-path    Print app config path
  aether history path       Print history JSONL path
  aether history list [n]   Print recent history
  aether favorites path     Print favorite tracks JSONL path
  aether favorites list     Print favorite tracks
  aether search <query>     Search stations via Radio Browser
  aether version            Print version
  aether help               Show help
`, version)
}

func trackLine(artist, title, raw string) string {
	if artist != "" && title != "" {
		return artist + " — " + title
	}
	if title != "" {
		return title
	}
	if raw != "" {
		return raw
	}
	return "—"
}

func truncate(s string, width int) string {
	if len([]rune(s)) <= width {
		return s
	}
	runes := []rune(s)
	return string(runes[:width-1]) + "…"
}

func compactTags(tags string) string {
	if strings.TrimSpace(tags) == "" {
		return "no tags"
	}
	parts := strings.Split(tags, ",")
	if len(parts) > 4 {
		parts = parts[:4]
	}
	return strings.Join(parts, ", ")
}
