package artwork

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"sync"
)

// Protocol represents the terminal graphics protocol to use
type Protocol string

const (
	ProtocolANSI  Protocol = "ansi"  // Half-block Unicode characters (universal fallback)
	ProtocolSixel Protocol = "sixel" // Sixel graphics (iTerm2, WezTerm, foot, mlterm, xterm)
	ProtocolKitty Protocol = "kitty" // Kitty graphics protocol
)

var (
	detectedProtocol     Protocol
	detectedProtocolOnce sync.Once
)

// DetectProtocol detects the best available graphics protocol for the current terminal.
// Results are cached after first detection.
func DetectProtocol() Protocol {
	detectedProtocolOnce.Do(func() {
		detectedProtocol = detectProtocolImpl()
	})
	return detectedProtocol
}

// ResetProtocolDetection resets the cached protocol detection (useful for testing)
func ResetProtocolDetection() {
	detectedProtocolOnce = sync.Once{}
	detectedProtocol = ""
}

// ForceProtocol allows overriding the detected protocol (for config/testing)
func ForceProtocol(p Protocol) {
	detectedProtocolOnce.Do(func() {})
	detectedProtocol = p
}

func detectProtocolImpl() Protocol {
	// Check for Kitty terminal first (best quality)
	if isKittyTerminal() {
		return ProtocolKitty
	}

	// Check for Sixel support
	if isSixelSupported() {
		return ProtocolSixel
	}

	// Fallback to ANSI half-blocks
	return ProtocolANSI
}

func isKittyTerminal() bool {
	// Kitty sets TERM=xterm-kitty or KITTY_WINDOW_ID
	term := os.Getenv("TERM")
	if strings.Contains(term, "kitty") {
		return true
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	return false
}

func isSixelSupported() bool {
	// Check common terminals that support Sixel
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	wtSession := os.Getenv("WT_SESSION")       // Windows Terminal
	lcTerminal := os.Getenv("LC_TERMINAL")     // iTerm2 sets this
	iterm := os.Getenv("ITERM_SESSION_ID")     // iTerm2
	wezterm := os.Getenv("WEZTERM_EXECUTABLE") // WezTerm

	// iTerm2
	if termProgram == "iTerm.app" || lcTerminal == "iTerm2" || iterm != "" {
		return true
	}

	// WezTerm
	if wezterm != "" || strings.Contains(strings.ToLower(termProgram), "wezterm") {
		return true
	}

	// foot terminal
	if term == "foot" || term == "foot-extra" {
		return true
	}

	// mlterm
	if strings.HasPrefix(term, "mlterm") {
		return true
	}

	// Windows Terminal (recent versions support sixel)
	if wtSession != "" {
		return true
	}

	// xterm with sixel (check XTERM_VERSION and assume sixel if modern)
	if strings.HasPrefix(term, "xterm") && os.Getenv("XTERM_VERSION") != "" {
		return true
	}

	// Konsole (KDE) - recent versions support sixel
	if termProgram == "konsole" {
		return true
	}

	// Contour terminal
	if os.Getenv("CONTOUR_SESSION_ID") != "" {
		return true
	}

	return false
}

// Render converts image data to the appropriate terminal format based on detected protocol.
// This is the main entry point that auto-selects the best rendering method.
func Render(ctx context.Context, data []byte, width, height int, quality QualityLevel, scaleMode ScaleMode) (string, error) {
	protocol := DetectProtocol()
	return RenderWithProtocol(ctx, data, width, height, quality, scaleMode, protocol)
}

// RenderWithProtocol converts image data using a specific protocol.
func RenderWithProtocol(ctx context.Context, data []byte, width, height int, quality QualityLevel, scaleMode ScaleMode, protocol Protocol) (string, error) {
	switch protocol {
	case ProtocolKitty:
		return ConvertToKitty(ctx, data, width, height)
	case ProtocolSixel:
		return ConvertToSixel(ctx, data, width, height)
	default:
		return ConvertToANSI(ctx, data, width, height, quality, scaleMode)
	}
}

// ConvertToKitty converts image data to Kitty graphics protocol format.
// The Kitty protocol transmits PNG data base64-encoded with escape sequences.
// Width/height are in terminal cells (characters).
func ConvertToKitty(ctx context.Context, data []byte, widthCells, heightCells int) (string, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	// Scale image to target pixel size
	// Assume ~10 pixels per cell width, ~20 pixels per cell height (typical terminal)
	targetWidth := widthCells * 10
	targetHeight := heightCells * 20

	// Scale maintaining aspect ratio
	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	scale := min(float64(targetWidth)/float64(imgW), float64(targetHeight)/float64(imgH))
	newW := int(float64(imgW) * scale)
	newH := int(float64(imgH) * scale)

	// Simple nearest-neighbor scaling
	scaled := scaleImage(img, newW, newH)

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("encode png: %w", err)
	}

	// Kitty graphics protocol format:
	// \x1b_Ga=T,f=100,s=<width>,v=<height>,c=<cols>,r=<rows>;<base64 data>\x1b\\
	// a=T: transmit and display
	// f=100: PNG format
	// c=cols, r=rows: size in terminal cells
	b64Data := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Kitty requires chunked transmission for large images (max 4096 bytes per chunk)
	var result strings.Builder
	chunkSize := 4096

	for i := 0; i < len(b64Data); i += chunkSize {
		end := i + chunkSize
		if end > len(b64Data) {
			end = len(b64Data)
		}
		chunk := b64Data[i:end]

		// m=0 for last chunk, m=1 for more chunks coming
		more := 0
		if end < len(b64Data) {
			more = 1
		}

		if i == 0 {
			// First chunk includes full header
			result.WriteString(fmt.Sprintf("\x1b_Ga=T,f=100,c=%d,r=%d,m=%d;%s\x1b\\",
				widthCells, heightCells, more, chunk))
		} else {
			// Continuation chunks
			result.WriteString(fmt.Sprintf("\x1b_Gm=%d;%s\x1b\\", more, chunk))
		}
	}

	return result.String(), nil
}

// ConvertToSixel converts image data to Sixel format.
// Sixel is a bitmap graphics format that works in many terminals.
// Width/height are in terminal cells (characters).
func ConvertToSixel(ctx context.Context, data []byte, widthCells, heightCells int) (string, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	// Scale image to target pixel size
	// Sixel: each cell is roughly 10x20 pixels, but sixel rows are 6 pixels high
	targetWidth := widthCells * 10
	targetHeight := heightCells * 20

	// Scale maintaining aspect ratio
	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	scale := min(float64(targetWidth)/float64(imgW), float64(targetHeight)/float64(imgH))
	newW := int(float64(imgW) * scale)
	newH := int(float64(imgH) * scale)

	// Ensure height is multiple of 6 (sixel row height)
	newH = ((newH + 5) / 6) * 6

	// Scale the image
	scaled := scaleImage(img, newW, newH)

	// Convert to sixel
	return encodeToSixel(scaled)
}

// scaleImage scales an image using nearest-neighbor interpolation
func scaleImage(img image.Image, newWidth, newHeight int) image.Image {
	bounds := img.Bounds()
	oldWidth := bounds.Dx()
	oldHeight := bounds.Dy()

	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := bounds.Min.X + x*oldWidth/newWidth
			srcY := bounds.Min.Y + y*oldHeight/newHeight
			result.Set(x, y, img.At(srcX, srcY))
		}
	}

	return result
}

// encodeToSixel encodes an image to sixel format
func encodeToSixel(img image.Image) (string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Build color palette (sixel supports up to 256 colors)
	// Use a simple quantization approach
	palette, colorMap := buildPalette(img, 256)

	var result strings.Builder

	// Sixel header: DCS P1 ; P2 ; P3 q
	// P1=0: pixel aspect ratio 2:1 (default)
	// P2=0: no background
	// P3=0: horizontal grid size
	result.WriteString("\x1bPq")

	// Define color palette
	// # Pc ; Pu ; Px ; Py ; Pz
	// Pc = color number, Pu = color system (2=RGB), Px/Py/Pz = R/G/B (0-100)
	for i, c := range palette {
		r := int(c.R) * 100 / 255
		g := int(c.G) * 100 / 255
		b := int(c.B) * 100 / 255
		result.WriteString(fmt.Sprintf("#%d;2;%d;%d;%d", i, r, g, b))
	}

	// Encode pixels in sixel format
	// Sixel encodes 6 vertical pixels per character
	for y := 0; y < height; y += 6 {
		// For each color, emit the sixel data for this row
		for colorIdx := range palette {
			var lineData strings.Builder
			hasData := false

			for x := 0; x < width; x++ {
				var sixelBits byte = 0

				// Collect 6 vertical pixels
				for bit := 0; bit < 6; bit++ {
					py := y + bit
					if py >= height {
						continue
					}

					px := bounds.Min.X + x
					ppy := bounds.Min.Y + py
					c := img.At(px, ppy)

					if colorMap[colorKey(c)] == colorIdx {
						sixelBits |= 1 << bit
					}
				}

				if sixelBits > 0 {
					hasData = true
				}

				// Sixel character = bits + 63 (0x3F)
				lineData.WriteByte(byte(sixelBits) + 0x3F)
			}

			if hasData {
				// Select color and output line
				result.WriteString(fmt.Sprintf("#%d%s", colorIdx, lineData.String()))
			}
		}

		// Move to next sixel row ($ = carriage return, - = newline)
		if y+6 < height {
			result.WriteString("-")
		}
	}

	// Sixel terminator
	result.WriteString("\x1b\\")

	return result.String(), nil
}

// RGBA represents a color
type RGBA struct {
	R, G, B, A uint8
}

func colorKey(c interface{}) uint32 {
	r, g, b, _ := c.(interface {
		RGBA() (uint32, uint32, uint32, uint32)
	}).RGBA()
	// Quantize to 5 bits per channel for palette matching
	return ((r >> 11) << 10) | ((g >> 11) << 5) | (b >> 11)
}

// buildPalette creates a color palette from an image using median cut quantization
// colorFreq holds a quantized color key and its frequency count
type colorFreq struct {
	key   uint32
	count int
}

func buildPalette(img image.Image, maxColors int) ([]RGBA, map[uint32]int) {
	bounds := img.Bounds()

	// Collect unique colors (quantized)
	colorCounts := make(map[uint32]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			key := colorKey(img.At(x, y))
			colorCounts[key]++
		}
	}

	// Simple approach: take most frequent colors
	var colors []colorFreq
	for k, v := range colorCounts {
		colors = append(colors, colorFreq{k, v})
	}

	// Sort by frequency
	for i := 0; i < len(colors)-1; i++ {
		for j := i + 1; j < len(colors); j++ {
			if colors[j].count > colors[i].count {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}

	// Take top colors
	if len(colors) > maxColors {
		colors = colors[:maxColors]
	}

	// Build palette and color map
	palette := make([]RGBA, len(colors))
	colorMap := make(map[uint32]int)

	for i, cf := range colors {
		// Expand back from quantized color
		r := uint8(((cf.key >> 10) & 0x1F) << 3)
		g := uint8(((cf.key >> 5) & 0x1F) << 3)
		b := uint8((cf.key & 0x1F) << 3)
		palette[i] = RGBA{r, g, b, 255}
		colorMap[cf.key] = i
	}

	// Map remaining colors to nearest palette entry
	for key := range colorCounts {
		if _, ok := colorMap[key]; !ok {
			colorMap[key] = findNearestColor(key, colors)
		}
	}

	return palette, colorMap
}

func findNearestColor(key uint32, palette []colorFreq) int {
	r := int((key >> 10) & 0x1F)
	g := int((key >> 5) & 0x1F)
	b := int(key & 0x1F)

	bestIdx := 0
	bestDist := 1 << 30

	for i, cf := range palette {
		pr := int((cf.key >> 10) & 0x1F)
		pg := int((cf.key >> 5) & 0x1F)
		pb := int(cf.key & 0x1F)

		dist := (r-pr)*(r-pr) + (g-pg)*(g-pg) + (b-pb)*(b-pb)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}

	return bestIdx
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
