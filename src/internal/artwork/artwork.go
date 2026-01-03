package artwork

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("artwork not found")
	ErrInvalid  = errors.New("invalid artwork data")
)

// Cache provides caching for converted ANSI artwork.
type Cache struct {
	baseDir   string
	cacheDays int
}

// NewCache creates a new artwork cache.
func NewCache(baseDir string, cacheDays int) (*Cache, error) {
	if baseDir == "" {
		var err error
		baseDir, err = defaultCacheDir()
		if err != nil {
			return nil, fmt.Errorf("resolve cache dir: %w", err)
		}
	}
	if cacheDays <= 0 {
		cacheDays = 30
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	return &Cache{
		baseDir:   baseDir,
		cacheDays: cacheDays,
	}, nil
}

func defaultCacheDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Caches", "tunez", "artwork")
	case "windows":
		dir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "Tunez", "artwork")
	default:
		dir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez", "artwork")
	}
	return base, nil
}

// cacheKey generates a cache key from artwork reference.
func cacheKey(ref string, width int) string {
	h := sha256.New()
	h.Write([]byte(ref))
	h.Write([]byte(fmt.Sprintf(":%d", width)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Get retrieves cached ANSI artwork if available and not expired.
func (c *Cache) Get(ref string, width int) (string, bool) {
	key := cacheKey(ref, width)
	path := filepath.Join(c.baseDir, key+".ansi")

	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}

	// Check expiration
	if time.Since(info.ModTime()) > time.Duration(c.cacheDays)*24*time.Hour {
		os.Remove(path)
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	return string(data), true
}

// Set stores ANSI artwork in the cache.
func (c *Cache) Set(ref string, width int, ansi string) error {
	key := cacheKey(ref, width)
	path := filepath.Join(c.baseDir, key+".ansi")
	return os.WriteFile(path, []byte(ansi), 0o644)
}

// Clear removes all cached artwork.
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ansi") {
			os.Remove(filepath.Join(c.baseDir, e.Name()))
		}
	}
	return nil
}

// Size returns the total size of cached artwork in bytes.
func (c *Cache) Size() (int64, error) {
	var total int64
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ansi") {
			info, err := e.Info()
			if err == nil {
				total += info.Size()
			}
		}
	}
	return total, nil
}

// ConvertToANSI converts image data to ANSI art.
func ConvertToANSI(ctx context.Context, data []byte, width, height int) (string, error) {
	if width <= 0 {
		width = 20
	}
	if height <= 0 {
		height = 10
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	// Get image bounds
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	if imgWidth == 0 || imgHeight == 0 {
		return "", ErrInvalid
	}

	// Calculate aspect ratio correction (terminal chars are ~2x tall as wide)
	aspectRatio := float64(imgWidth) / float64(imgHeight)
	charHeight := int(float64(width) / aspectRatio / 2)
	if charHeight > height {
		charHeight = height
		width = int(float64(charHeight) * 2 * aspectRatio)
	}

	// Sample and convert to ANSI
	var result strings.Builder
	for y := 0; y < charHeight; y++ {
		for x := 0; x < width; x++ {
			// Sample pixel
			sampleX := (x * imgWidth) / width
			sampleY := (y * imgHeight) / charHeight
			if sampleX >= imgWidth {
				sampleX = imgWidth - 1
			}
			if sampleY >= imgHeight {
				sampleY = imgHeight - 1
			}

			r, g, b, a := img.At(bounds.Min.X+sampleX, bounds.Min.Y+sampleY).RGBA()
			// Convert from 16-bit to 8-bit
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			a8 := uint8(a >> 8)

			// Handle transparency
			if a8 < 128 {
				result.WriteString(" ")
				continue
			}

			// Use 256-color mode for better compatibility
			colorCode := rgbTo256(r8, g8, b8)
			result.WriteString(fmt.Sprintf("\x1b[48;5;%dm \x1b[0m", colorCode))
		}
		if y < charHeight-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// rgbTo256 converts RGB to the closest 256-color palette index.
func rgbTo256(r, g, b uint8) int {
	// For grayscale
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return int((r-8)/10) + 232
	}

	// For colors, use the 6x6x6 color cube (indices 16-231)
	ri := int(r) * 5 / 255
	gi := int(g) * 5 / 255
	bi := int(b) * 5 / 255
	return 16 + 36*ri + 6*gi + bi
}

// Placeholder returns a placeholder ANSI art when artwork is unavailable.
func Placeholder(width, height int) string {
	if width <= 0 {
		width = 20
	}
	if height <= 0 {
		height = 10
	}

	var result strings.Builder
	// Simple bordered box
	result.WriteString("┌" + strings.Repeat("─", width-2) + "┐\n")
	for y := 1; y < height-1; y++ {
		result.WriteString("│")
		if y == height/2 {
			text := "♪"
			padding := (width - 2 - len(text)) / 2
			result.WriteString(strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-2-padding-len(text)))
		} else {
			result.WriteString(strings.Repeat(" ", width-2))
		}
		result.WriteString("│\n")
	}
	result.WriteString("└" + strings.Repeat("─", width-2) + "┘")
	return result.String()
}
