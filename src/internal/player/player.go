package player

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Event describes playback state updates emitted by mpv.
type Event struct {
	TimePos   *float64
	Duration  *float64
	Paused    *bool
	Volume    *float64
	Muted     *bool
	Ended     bool   // true when track ended naturally (eof)
	EndReason string // "eof", "stop", "quit", "error", "redirect"
	Err       error
}

// Options configures the Controller.
type Options struct {
	MPVPath        string
	IPCPath        string
	Logger         *slog.Logger
	DisableProcess bool
	Dial           func(ctx context.Context, network, addr string) (net.Conn, error)
	ExtraArgs      []string
}

// Controller manages the mpv process and IPC connection.
type Controller struct {
	opts   Options
	cmd    *exec.Cmd
	conn   net.Conn
	mu     sync.Mutex
	events chan Event
	done   chan struct{}
}

func New(opts Options) *Controller {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Controller{
		opts:   opts,
		events: make(chan Event, 32),
		done:   make(chan struct{}),
	}
}

func defaultIPCPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\tunez-mpv`
	}
	return filepath.Join(os.TempDir(), "tunez-mpv.sock")
}

// Start launches mpv (unless disabled) and connects to the IPC socket.
func (c *Controller) Start(ctx context.Context) error {
	c.opts.Logger.Debug("starting player controller", slog.String("ipc_path", c.opts.IPCPath), slog.Bool("disable_process", c.opts.DisableProcess))
	c.mu.Lock()
	// Reinitialize done channel if previously closed (for restarts)
	select {
	case <-c.done:
		c.done = make(chan struct{})
	default:
	}
	c.mu.Unlock()

	if c.opts.IPCPath == "" {
		c.opts.IPCPath = defaultIPCPath()
		c.opts.Logger.Debug("using default ipc path", slog.String("ipc_path", c.opts.IPCPath))
	}
	if !c.opts.DisableProcess {
		if err := c.spawnMPV(ctx); err != nil {
			c.opts.Logger.Error("failed to spawn mpv", slog.Any("err", err))
			return err
		}
		c.opts.Logger.Debug("mpv process spawned")
	}
	if err := c.connect(ctx); err != nil {
		c.opts.Logger.Error("failed to connect to mpv ipc", slog.Any("err", err))
		return err
	}
	c.opts.Logger.Debug("connected to mpv ipc")
	if err := c.observeProperties(); err != nil {
		c.opts.Logger.Error("failed to observe mpv properties", slog.Any("err", err))
		return err
	}
	c.opts.Logger.Debug("started observing mpv properties")
	go c.readLoop()
	c.opts.Logger.Debug("player controller started successfully")
	return nil
}

func (c *Controller) spawnMPV(ctx context.Context) error {
	args := []string{
		"--idle=yes",
		"--force-window=no",
		"--no-terminal",
		"--no-video",
		"--input-ipc-server=" + c.opts.IPCPath,
	}
	args = append(args, c.opts.ExtraArgs...)
	c.opts.Logger.Debug("spawning mpv process", slog.String("mpv_path", c.opts.MPVPath), slog.Any("args", args))
	c.cmd = exec.CommandContext(ctx, c.opts.MPVPath, args...)
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}
	c.opts.Logger.Debug("mpv process started", slog.Int("pid", c.cmd.Process.Pid))
	return nil
}

func (c *Controller) connect(ctx context.Context) error {
	c.opts.Logger.Debug("connecting to mpv ipc", slog.String("ipc_path", c.opts.IPCPath))
	dial := c.opts.Dial
	if dial == nil {
		dial = (&net.Dialer{Timeout: 5 * time.Second}).DialContext
	}
	var conn net.Conn
	var err error
	baseDelay := 50 * time.Millisecond
	maxDelay := 500 * time.Millisecond
	maxRetries := 10
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < maxRetries; i++ {
		conn, err = dial(ctx, networkForPath(c.opts.IPCPath), c.opts.IPCPath)
		if err == nil {
			c.conn = conn
			c.opts.Logger.Debug("connected to mpv ipc on attempt", slog.Int("attempt", i+1))
			return nil
		}

		select {
		case <-ctx.Done():
			c.opts.Logger.Error("mpv ipc connection cancelled", slog.Any("err", ctx.Err()))
			return fmt.Errorf("connect mpv ipc: %w", ctx.Err())
		default:
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i))
			if delay > maxDelay {
				delay = maxDelay
			}
			jitter := time.Duration(float64(delay) * 0.2 * rng.Float64())
			c.opts.Logger.Debug("mpv ipc connection failed, retrying", slog.Int("attempt", i+1), slog.Any("err", err), slog.Duration("delay", delay+jitter))
			time.Sleep(delay + jitter)
		}
	}
	c.opts.Logger.Error("failed to connect to mpv ipc after retries", slog.Any("err", err))
	return fmt.Errorf("connect mpv ipc: %w", err)
}

func networkForPath(path string) string {
	return "unix"
}

func (c *Controller) observeProperties() error {
	props := []string{"time-pos", "duration", "pause", "volume", "mute"}
	for i, p := range props {
		if err := c.send(map[string]any{
			"command": []any{"observe_property", i + 1, p},
		}); err != nil {
			return err
		}
	}
	return nil
}

// Events returns the event channel.
func (c *Controller) Events() <-chan Event { return c.events }

func (c *Controller) send(cmd map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("mpv not connected")
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(append(b, '\n'))
	return err
}

// Play loads a URL into mpv.
func (c *Controller) Play(url string, headers map[string]string) error {
	c.opts.Logger.Debug("playing track", slog.String("url", url), slog.Int("header_count", len(headers)))
	if len(headers) > 0 {
		var headerLines []string
		for k, v := range headers {
			headerLines = append(headerLines, fmt.Sprintf("%s: %s", k, v))
		}
		_ = c.send(map[string]any{"command": []any{"set_property", "http-header-fields", strings.Join(headerLines, "\n")}})
	}
	err := c.send(map[string]any{
		"command": []any{"loadfile", url, "replace"},
	})
	if err != nil {
		c.opts.Logger.Error("failed to send play command", slog.Any("err", err))
	} else {
		c.opts.Logger.Debug("play command sent successfully")
	}
	return err
}

func (c *Controller) TogglePause(paused bool) error {
	c.opts.Logger.Debug("toggling pause", slog.Bool("paused", paused))
	err := c.send(map[string]any{"command": []any{"set_property", "pause", paused}})
	if err != nil {
		c.opts.Logger.Error("failed to send pause command", slog.Any("err", err))
	}
	return err
}

func (c *Controller) Seek(deltaSeconds float64) error {
	c.opts.Logger.Debug("seeking", slog.Float64("delta_seconds", deltaSeconds))
	err := c.send(map[string]any{"command": []any{"seek", deltaSeconds, "relative"}})
	if err != nil {
		c.opts.Logger.Error("failed to send seek command", slog.Any("err", err))
	}
	return err
}

func (c *Controller) SetVolume(vol float64) error {
	if vol < 0 {
		vol = 0
	}
	if vol > 100 {
		vol = 100
	}
	c.opts.Logger.Debug("setting volume", slog.Float64("volume", vol))
	err := c.send(map[string]any{"command": []any{"set_property", "volume", vol}})
	if err != nil {
		c.opts.Logger.Error("failed to send volume command", slog.Any("err", err))
	}
	return err
}

func (c *Controller) SetMute(mute bool) error {
	c.opts.Logger.Debug("setting mute", slog.Bool("mute", mute))
	err := c.send(map[string]any{"command": []any{"set_property", "mute", mute}})
	if err != nil {
		c.opts.Logger.Error("failed to send mute command", slog.Any("err", err))
	}
	return err
}

func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close done channel only once
	select {
	case <-c.done:
		// already closed
	default:
		close(c.done)
	}

	if c.conn != nil {
		// Send quit command (best effort, ignore errors)
		b, _ := json.Marshal(map[string]any{"command": []any{"quit"}})
		_, _ = c.conn.Write(append(b, '\n'))
		_ = c.conn.Close()
		c.conn = nil
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait() // Reap zombie process
		c.cmd = nil
	}
	return nil
}

func (c *Controller) readLoop() {
	defer close(c.events)
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg ipcMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			c.events <- Event{Err: fmt.Errorf("decode: %w", err)}
			continue
		}
		switch msg.Event {
		case "property-change":
			c.handlePropertyChange(msg)
		case "end-file":
			// Only set Ended=true for natural end (eof), not for stop/quit/error
			// "stop" happens when we load a new file, "quit" when mpv exits
			c.events <- Event{
				Ended:     msg.Reason == "eof",
				EndReason: msg.Reason,
			}
		}
	}
	if err := scanner.Err(); err != nil {
		c.events <- Event{Err: err}
	}
}

type ipcMessage struct {
	Event  string      `json:"event"`
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	Reason string      `json:"reason"` // for end-file event: "eof", "stop", "quit", "error", "redirect"
}

func (c *Controller) handlePropertyChange(msg ipcMessage) {
	switch msg.Name {
	case "time-pos":
		if v, ok := toFloat(msg.Data); ok {
			c.events <- Event{TimePos: &v}
		}
	case "duration":
		if v, ok := toFloat(msg.Data); ok {
			c.events <- Event{Duration: &v}
		}
	case "pause":
		if b, ok := msg.Data.(bool); ok {
			c.events <- Event{Paused: &b}
		}
	case "volume":
		if v, ok := toFloat(msg.Data); ok {
			c.events <- Event{Volume: &v}
		}
	case "mute":
		if b, ok := msg.Data.(bool); ok {
			c.events <- Event{Muted: &b}
		}
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
