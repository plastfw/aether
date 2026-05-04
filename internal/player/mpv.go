package player

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Health struct {
	AudioCodec         string
	AudioBitrate       int
	IdleActive         bool
	CoreIdle           bool
	PausedForCache     bool
	CacheBuffering     int
	DemuxerCacheSecond float64
}

type Player struct {
	cmd       *exec.Cmd
	socket    string
	volume    int
	maxVolume int
	paused    bool
	mu        sync.Mutex
}

func New(defaultVolume int, maxVolume int) *Player {
	if maxVolume < 100 {
		maxVolume = 100
	}
	if defaultVolume < 0 {
		defaultVolume = 0
	}
	if defaultVolume > maxVolume {
		defaultVolume = maxVolume
	}
	return &Player{volume: defaultVolume, maxVolume: maxVolume}
}

func (p *Player) Play(url string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensureStartedLocked(); err != nil {
		return err
	}
	if err := p.commandLocked("set_property", "volume", p.volume); err != nil {
		return fmt.Errorf("sync volume: %w", err)
	}
	if err := p.commandLocked("loadfile", url, "replace"); err != nil {
		return fmt.Errorf("load stream: %w", err)
	}
	if err := p.commandLocked("set_property", "pause", false); err != nil {
		return fmt.Errorf("resume stream: %w", err)
	}
	p.paused = false
	return nil
}

func (p *Player) AddVolume(delta int) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.commandLocked("add", "volume", delta); err != nil {
		return p.volume, err
	}
	vol, err := p.volumeLocked()
	if err != nil {
		return p.volume, err
	}
	p.volume = vol
	return vol, nil
}

func (p *Player) SetVolume(volume int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if volume < 0 {
		volume = 0
	}
	if volume > p.maxVolume {
		volume = p.maxVolume
	}
	if err := p.ensureStartedLocked(); err != nil {
		return err
	}
	if err := p.commandLocked("set_property", "volume", volume); err != nil {
		return err
	}
	p.volume = volume
	return nil
}

func (p *Player) Volume() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	vol, err := p.volumeLocked()
	if err != nil {
		return p.volume, err
	}
	p.volume = vol
	return vol, nil
}

func (p *Player) TogglePause() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.commandLocked("cycle", "pause"); err != nil {
		return p.paused, err
	}
	paused := p.boolPropertyLocked("pause")
	p.paused = paused
	return paused, nil
}

func (p *Player) Paused() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	paused, err := p.pausedLocked()
	if err != nil {
		return p.paused, err
	}
	p.paused = paused
	return paused, nil
}

func (p *Player) Health() Health {
	p.mu.Lock()
	defer p.mu.Unlock()

	return Health{
		AudioCodec:         p.stringPropertyLocked("audio-codec"),
		AudioBitrate:       p.intPropertyLocked("audio-bitrate"),
		IdleActive:         p.boolPropertyLocked("idle-active"),
		CoreIdle:           p.boolPropertyLocked("core-idle"),
		PausedForCache:     p.boolPropertyLocked("paused-for-cache"),
		CacheBuffering:     p.intPropertyLocked("cache-buffering-state"),
		DemuxerCacheSecond: p.floatPropertyLocked("demuxer-cache-duration"),
	}
}

func (p *Player) RawMetadata() (map[string]any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := p.queryLocked("get_property", "metadata")
	if err != nil {
		return nil, err
	}
	obj, ok := data.(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}
	return obj, nil
}

func (p *Player) Stop() error { return p.Quit() }

func (p *Player) Quit() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	_ = p.commandLocked("quit")
	done := make(chan struct{})
	go func(cmd *exec.Cmd) {
		_, _ = cmd.Process.Wait()
		close(done)
	}(p.cmd)

	select {
	case <-done:
	case <-time.After(700 * time.Millisecond):
		_ = p.cmd.Process.Kill()
		<-done
	}

	p.cmd = nil
	if p.socket != "" {
		_ = os.Remove(p.socket)
		p.socket = ""
	}
	return nil
}

func (p *Player) ensureStartedLocked() error {
	if p.cmd != nil && p.cmd.Process != nil && p.socket != "" {
		if conn, err := net.DialTimeout("unix", p.socket, 100*time.Millisecond); err == nil {
			_ = conn.Close()
			return nil
		}
	}

	if _, err := exec.LookPath("mpv"); err != nil {
		return errors.New("mpv not found: install it with `sudo pacman -S mpv`")
	}

	p.socket = filepath.Join(os.TempDir(), fmt.Sprintf("aether-mpv-%d.sock", os.Getpid()))
	_ = os.Remove(p.socket)

	cmd := exec.Command(
		"mpv",
		"--no-video",
		"--really-quiet",
		"--force-window=no",
		"--idle=yes",
		fmt.Sprintf("--volume=%d", p.volume),
		fmt.Sprintf("--volume-max=%d", p.maxVolume),
		"--input-ipc-server="+p.socket,
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}
	p.cmd = cmd
	return p.waitIPCLocked()
}

func (p *Player) waitIPCLocked() error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", p.socket, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("mpv IPC socket did not become ready")
}

func (p *Player) commandLocked(args ...any) error {
	_, err := p.queryLocked(args...)
	return err
}

func (p *Player) boolPropertyLocked(name string) bool {
	value, _ := p.boolPropertyWithErrorLocked(name)
	return value
}

func (p *Player) pausedLocked() (bool, error) {
	return p.boolPropertyWithErrorLocked("pause")
}

func (p *Player) boolPropertyWithErrorLocked(name string) (bool, error) {
	data, err := p.queryLocked("get_property", name)
	if err != nil {
		return false, err
	}
	value, ok := data.(bool)
	return ok && value, nil
}

func (p *Player) stringPropertyLocked(name string) string {
	data, err := p.queryLocked("get_property", name)
	if err != nil {
		return ""
	}
	value, ok := data.(string)
	if !ok {
		return ""
	}
	return value
}

func (p *Player) intPropertyLocked(name string) int {
	value := p.floatPropertyLocked(name)
	if value <= 0 {
		return 0
	}
	return int(math.Round(value))
}

func (p *Player) floatPropertyLocked(name string) float64 {
	data, err := p.queryLocked("get_property", name)
	if err != nil {
		return 0
	}
	value, ok := data.(float64)
	if !ok {
		return 0
	}
	return value
}

func (p *Player) volumeLocked() (int, error) {
	data, err := p.queryLocked("get_property", "volume")
	if err != nil {
		return p.volume, err
	}
	value, ok := data.(float64)
	if !ok {
		return p.volume, nil
	}
	return int(math.Round(value)), nil
}

func (p *Player) queryLocked(args ...any) (any, error) {
	if p.socket == "" {
		return nil, errors.New("mpv is not running")
	}

	conn, err := net.DialTimeout("unix", p.socket, 500*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	payload, err := json.Marshal(map[string]any{"command": args})
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(700 * time.Millisecond))
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data  any    `json:"data"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" && resp.Error != "success" {
		return nil, errors.New(resp.Error)
	}
	return resp.Data, nil
}
