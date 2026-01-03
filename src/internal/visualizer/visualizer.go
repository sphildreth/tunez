// Package visualizer provides real-time audio spectrum visualization using CAVA.
package visualizer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Visualizer manages CAVA subprocess for audio spectrum analysis.
type Visualizer struct {
	mu       sync.RWMutex
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	bars     []int
	barCount int
	maxValue int
	running  bool
	err      error
}

// Config holds visualizer configuration.
type Config struct {
	BarCount int // Number of frequency bars (default: 16)
	MaxValue int // Maximum bar value for scaling (default: 1000)
}

// New creates a new Visualizer instance.
func New(cfg Config) *Visualizer {
	if cfg.BarCount <= 0 {
		cfg.BarCount = 16
	}
	if cfg.MaxValue <= 0 {
		cfg.MaxValue = 1000
	}
	return &Visualizer{
		barCount: cfg.BarCount,
		maxValue: cfg.MaxValue,
		bars:     make([]int, cfg.BarCount),
	}
}

// Available checks if CAVA is installed on the system.
func Available() bool {
	_, err := exec.LookPath("cava")
	return err == nil
}

// Start begins the CAVA subprocess and starts reading spectrum data.
func (v *Visualizer) Start(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.running {
		return nil
	}

	if !Available() {
		v.err = fmt.Errorf("cava not installed")
		return v.err
	}

	// Create temporary config file for CAVA
	configPath, err := v.writeConfig()
	if err != nil {
		v.err = err
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	v.cmd = exec.CommandContext(ctx, "cava", "-p", configPath)
	v.cmd.Stderr = nil // Suppress stderr

	stdout, err := v.cmd.StdoutPipe()
	if err != nil {
		v.err = err
		return err
	}

	if err := v.cmd.Start(); err != nil {
		v.err = err
		return err
	}

	v.running = true
	v.err = nil

	// Read CAVA output in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			v.parseLine(line)
		}
		v.mu.Lock()
		v.running = false
		v.mu.Unlock()
	}()

	// Cleanup config file when done
	go func() {
		<-ctx.Done()
		os.Remove(configPath)
	}()

	return nil
}

// Stop terminates the CAVA subprocess.
func (v *Visualizer) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	if v.cmd != nil && v.cmd.Process != nil {
		v.cmd.Process.Kill()
		v.cmd.Wait()
		v.cmd = nil
	}
	v.running = false
}

// Bars returns a copy of the current bar values (0 to maxValue).
func (v *Visualizer) Bars() []int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]int, len(v.bars))
	copy(result, v.bars)
	return result
}

// BarsNormalized returns bar values normalized to 0-8 range for display.
func (v *Visualizer) BarsNormalized() []int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]int, len(v.bars))
	for i, val := range v.bars {
		// Normalize to 0-8 range
		normalized := (val * 8) / v.maxValue
		if normalized > 8 {
			normalized = 8
		}
		result[i] = normalized
	}
	return result
}

// Running returns true if the visualizer is active.
func (v *Visualizer) Running() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.running
}

// Error returns any error that occurred.
func (v *Visualizer) Error() error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.err
}

// parseLine parses a line of CAVA raw output.
func (v *Visualizer) parseLine(line string) {
	parts := strings.Split(line, ";")
	v.mu.Lock()
	defer v.mu.Unlock()

	for i := 0; i < len(v.bars) && i < len(parts); i++ {
		if val, err := strconv.Atoi(strings.TrimSpace(parts[i])); err == nil {
			v.bars[i] = val
		}
	}
}

// writeConfig creates a temporary CAVA config file.
func (v *Visualizer) writeConfig() (string, error) {
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("tunez-cava-%d.conf", os.Getpid()))

	config := fmt.Sprintf(`[general]
bars = %d
framerate = 30
sensitivity = 100
autosens = 1
overshoot = 20

[input]
method = pulse
; Falls back to other methods automatically

[output]
method = raw
data_format = ascii
ascii_max_range = %d
bar_delimiter = 59
frame_delimiter = 10

[smoothing]
integral = 77
monstercat = 0
waves = 0
gravity = 100

[eq]
; Equal weight across bands
`, v.barCount, v.maxValue)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("write cava config: %w", err)
	}

	return configPath, nil
}

// Render returns an ANSI string representation of the current spectrum.
// Uses Unicode block characters for smooth visualization.
func (v *Visualizer) Render() string {
	bars := v.BarsNormalized()
	if len(bars) == 0 {
		return ""
	}

	// Unicode block characters for different heights
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	sb.WriteString("║")
	for _, val := range bars {
		if val < 0 {
			val = 0
		}
		if val > 8 {
			val = 8
		}
		sb.WriteRune(blocks[val])
	}
	sb.WriteString("║")

	return sb.String()
}

// RenderWithColor returns a colored ANSI string representation.
func (v *Visualizer) RenderWithColor(colorCode string) string {
	bars := v.BarsNormalized()
	if len(bars) == 0 {
		return ""
	}

	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	sb.WriteString("║")
	for _, val := range bars {
		if val < 0 {
			val = 0
		}
		if val > 8 {
			val = 8
		}
		sb.WriteRune(blocks[val])
	}
	sb.WriteString("║")

	return sb.String()
}
