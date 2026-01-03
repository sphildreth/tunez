package artwork

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
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
	"sync"
	"time"
)

//go:embed default_artwork.png
var defaultArtworkPNG []byte

var (
	defaultArtworkCache   = make(map[string]string)
	defaultArtworkCacheMu sync.RWMutex
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

// ConvertToANSI converts image data to ANSI art using half-block characters
// for double vertical resolution and true color for best quality.
func ConvertToANSI(ctx context.Context, data []byte, width, height int) (string, error) {
	if width <= 0 {
		width = 40
	}
	if height <= 0 {
		height = 20
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

	// Calculate dimensions preserving aspect ratio
	// Each character cell displays 2 vertical pixels using half-blocks
	aspectRatio := float64(imgWidth) / float64(imgHeight)
	charWidth := width
	charHeight := int(float64(width) / aspectRatio)
	
	// charHeight is in "half rows" (2 pixels per char), so actual terminal rows = charHeight/2
	terminalRows := (charHeight + 1) / 2
	if terminalRows > height {
		terminalRows = height
		charHeight = terminalRows * 2
		charWidth = int(float64(charHeight) * aspectRatio)
	}

	// Use half-block rendering: each character shows top and bottom pixel
	// ▀ = upper half block (fg = top pixel, bg = bottom pixel)
	var result strings.Builder
	for row := 0; row < terminalRows; row++ {
		for x := 0; x < charWidth; x++ {
			// Sample top pixel (row * 2)
			topY := (row * 2 * imgHeight) / charHeight
			// Sample bottom pixel (row * 2 + 1)
			bottomY := ((row*2 + 1) * imgHeight) / charHeight
			sampleX := (x * imgWidth) / charWidth

			if sampleX >= imgWidth {
				sampleX = imgWidth - 1
			}
			if topY >= imgHeight {
				topY = imgHeight - 1
			}
			if bottomY >= imgHeight {
				bottomY = imgHeight - 1
			}

			// Get top pixel color
			tr, tg, tb, ta := img.At(bounds.Min.X+sampleX, bounds.Min.Y+topY).RGBA()
			tr8, tg8, tb8 := uint8(tr>>8), uint8(tg>>8), uint8(tb>>8)
			ta8 := uint8(ta >> 8)

			// Get bottom pixel color  
			br, bg, bb, ba := img.At(bounds.Min.X+sampleX, bounds.Min.Y+bottomY).RGBA()
			br8, bg8, bb8 := uint8(br>>8), uint8(bg>>8), uint8(bb>>8)
			ba8 := uint8(ba >> 8)

			// Handle transparency
			topTransparent := ta8 < 128
			bottomTransparent := ba8 < 128

			if topTransparent && bottomTransparent {
				result.WriteString(" ")
			} else if topTransparent {
				// Only bottom pixel visible - use lower half block
				result.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▄\x1b[0m", br8, bg8, bb8))
			} else if bottomTransparent {
				// Only top pixel visible - use upper half block
				result.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▀\x1b[0m", tr8, tg8, tb8))
			} else {
				// Both pixels visible - use upper half block with fg=top, bg=bottom
				result.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m", 
					tr8, tg8, tb8, br8, bg8, bb8))
			}
		}
		if row < terminalRows-1 {
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

// DefaultArtwork returns the default tunez logo as ANSI art.
// Results are cached in memory by dimensions.
func DefaultArtwork(width, height int) string {
	if width <= 0 {
		width = 20
	}
	if height <= 0 {
		height = 10
	}

	key := fmt.Sprintf("%d:%d", width, height)

	// Check cache
	defaultArtworkCacheMu.RLock()
	if cached, ok := defaultArtworkCache[key]; ok {
		defaultArtworkCacheMu.RUnlock()
		return cached
	}
	defaultArtworkCacheMu.RUnlock()

	// Convert embedded PNG to ANSI
	result, err := ConvertToANSI(context.Background(), defaultArtworkPNG, width, height)
	if err != nil {
		// Fallback to text placeholder if conversion fails
		return Placeholder(width, height)
	}

	// Cache result
	defaultArtworkCacheMu.Lock()
	defaultArtworkCache[key] = result
	defaultArtworkCacheMu.Unlock()

	return result
}
