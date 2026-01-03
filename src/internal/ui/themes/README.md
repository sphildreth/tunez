# Tunez Themes

This directory contains color themes for the Tunez terminal music player. Each theme is defined in its own file and self-registers on import.

## Available Themes

| Theme | Description | Style |
|-------|-------------|-------|
| `rainbow` | Default colorful theme with vibrant pinks, blues, and purples | Multi-color |
| `blue` | Monochrome blue theme | Monochrome |
| `coffee` | Warm brown coffee/mocha inspired | Multi-color |
| `cyan` | Cool cyan/teal theme | Monochrome |
| `dracula` | Based on the popular Dracula color scheme | Multi-color |
| `forest` | Earthy green and brown nature-inspired | Multi-color |
| `green` | Classic green-on-black terminal aesthetic | Monochrome |
| `gruvbox` | Retro groove colors from the Gruvbox palette | Multi-color |
| `matrix` | Green "Matrix" style hacker theme | Monochrome |
| `mono` | Grayscale theme using white, gray, and dark gray | Monochrome |
| `neon` | Electric, high-contrast neon signs | Multi-color |
| `nocolor` | High-contrast theme for NO_COLOR environments (no ANSI colors) | Accessible |
| `nord` | Arctic, north-bluish colors from the Nord palette | Multi-color |
| `ocean` | Deep blue ocean-inspired colors | Multi-color |
| `orange` | Warm orange/amber theme | Monochrome |
| `pink` | Soft pink/magenta theme | Monochrome |
| `purple` | Monochrome purple/violet theme | Monochrome |
| `red` | Monochrome red theme | Monochrome |
| `solarized` | Based on the Solarized dark color scheme | Multi-color |
| `sunset` | Warm orange-to-purple gradient inspired | Multi-color |
| `synthwave` | 80s retro synthwave/outrun with neon colors | Multi-color |

## Theme Structure

Each theme must define a `Theme` struct with the following style elements:

```go
type Theme struct {
    Name      string           // Unique identifier (e.g., "rainbow", "dracula")
    Accent    lipgloss.Style   // Primary accent color for emphasis
    Dim       lipgloss.Style   // Muted/secondary text
    Text      lipgloss.Style   // Default body text
    Title     lipgloss.Style   // Section headers and titles
    Error     lipgloss.Style   // Error messages
    Success   lipgloss.Style   // Success messages
    Warning   lipgloss.Style   // Warning messages
    Border    lipgloss.Style   // UI borders and separators
    Highlight lipgloss.Style   // Selected/highlighted items
}
```

### Style Element Usage

| Element | Used For |
|---------|----------|
| `Accent` | Important UI elements, active items, player controls |
| `Dim` | Secondary information, timestamps, hints, disabled items |
| `Text` | Primary content text, track names, artist names |
| `Title` | Screen titles, section headers ("Now Playing", "Queue", etc.) |
| `Error` | Error messages, failed operations |
| `Success` | Success confirmations, completed actions |
| `Warning` | Warnings, important notices |
| `Border` | Box borders, separators, frames |
| `Highlight` | Currently selected item, focused element, progress bars |

## Creating a New Theme

### Step 1: Create a New File

Create a new file in this directory named after your theme (e.g., `mytheme.go`):

```go
package themes

import "github.com/charmbracelet/lipgloss"

func init() {
    Register("mytheme", MyTheme)
}

// MyTheme is a description of your theme.
func MyTheme(noColor bool) Theme {
    // Always handle noColor mode for accessibility
    if noColor {
        return NoColor(noColor)
    }
    
    // Define your colors
    primary := lipgloss.Color("#FF6600")
    secondary := lipgloss.Color("#CC5500")
    // ... more colors
    
    return Theme{
        Name:      "mytheme",
        Accent:    lipgloss.NewStyle().Foreground(primary).Bold(true),
        Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
        Text:      lipgloss.NewStyle().Foreground(secondary),
        Title:     lipgloss.NewStyle().Foreground(primary).Bold(true),
        Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true),
        Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true),
        Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true),
        Border:    lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
        Highlight: lipgloss.NewStyle().Foreground(primary).Bold(true).Underline(true),
    }
}
```

### Step 2: Key Requirements

1. **Package declaration**: Must be `package themes`

2. **Import lipgloss**: `import "github.com/charmbracelet/lipgloss"`

3. **Self-registration**: Use `init()` to register your theme:
   ```go
   func init() {
       Register("mytheme", MyTheme)
   }
   ```

4. **Handle noColor**: Always check the `noColor` parameter and return `NoColor(noColor)` if true:
   ```go
   if noColor {
       return NoColor(noColor)
   }
   ```

5. **Set the Name field**: Must match the registered name exactly

### Step 3: Color Guidelines

#### Hex Colors
Use hex color codes for consistency:
```go
lipgloss.Color("#FF6600")  // Orange
lipgloss.Color("#00FF00")  // Green
```

#### ANSI 256 Colors
For broader terminal compatibility, you can use ANSI 256 color codes:
```go
lipgloss.Color("202")  // Orange (ANSI 256)
lipgloss.Color("46")   // Green (ANSI 256)
```

#### Style Methods
Common lipgloss style methods:
```go
lipgloss.NewStyle().
    Foreground(color).      // Text color
    Background(color).      // Background color
    Bold(true).             // Bold text
    Italic(true).           // Italic text
    Underline(true).        // Underlined text
    Reverse(true).          // Swap fg/bg colors
    Blink(true).            // Blinking text (use sparingly!)
    Faint(true).            // Dimmed text
```

### Step 4: Testing Your Theme

After creating your theme file, the tests will automatically pick it up:

```bash
cd src
go test ./internal/ui/themes/... -v
```

The `TestThemesNotPanic` test will verify your theme can be constructed.

### Step 5: Manual Testing

Test your theme in the actual application:

```bash
# Edit your config to use your theme
# ~/.config/tunez/config.toml
# [ui]
# theme = "mytheme"

# Or temporarily test without changing config
cd src
go run ./cmd/tunez
```

## Design Tips

### Monochrome Themes
For single-color themes, use different shades of the same hue:
```go
bright := lipgloss.Color("#FF4444")  // Lightest
medium := lipgloss.Color("#CC0000")  // Medium
dark := lipgloss.Color("#880000")    // Dark
dim := lipgloss.Color("#550000")     // Dimmest
```

### Multi-Color Themes
For multi-color themes, ensure good contrast between:
- Text and background (readability)
- Accent and Text (visual hierarchy)
- Error/Success/Warning (distinguishable states)

### Accessibility
- Ensure sufficient contrast ratios
- Don't rely solely on color to convey information
- Test with `nocolor` mode
- Consider colorblind users (avoid red/green as only differentiators)

### Popular Color Scheme Sources
- [Dracula Theme](https://draculatheme.com/)
- [Nord Theme](https://www.nordtheme.com/)
- [Solarized](https://ethanschoonover.com/solarized/)
- [Gruvbox](https://github.com/morhetz/gruvbox)
- [Catppuccin](https://github.com/catppuccin/catppuccin)
- [Tokyo Night](https://github.com/enkia/tokyo-night-vscode-theme)

## File Structure

```
themes/
├── coffee.go          # Coffee browns
├── blue.go            # Blue monochrome
├── cyan.go            # Cyan monochrome
├── dracula.go         # Dracula scheme
├── forest.go          # Forest greens
├── green.go           # Green terminal theme
├── gruvbox.go         # Gruvbox scheme
├── matrix.go          # Matrix hacker
├── mono.go            # Grayscale theme
├── neon.go            # Neon electric
├── nocolor.go         # NO_COLOR accessible theme
├── nord.go            # Nord scheme
├── ocean.go           # Ocean blues
├── orange.go          # Orange monochrome
├── pink.go            # Pink monochrome
├── purple.go          # Purple monochrome
├── rainbow.go         # Default theme
├── README.md          # This file
├── red.go             # Red monochrome
├── solarized.go       # Solarized scheme
├── sunset.go          # Sunset gradient
├── synthwave.go       # 80s retro
├── themes_test.go     # Registry tests
└── themes.go          # Theme struct, registry, Get/Valid/Names functions
```

## API Reference

### Registration

```go
// Register adds a theme to the registry (call from init())
Register(name string, fn ThemeFunc)
```

### Retrieval

```go
// Get returns a theme by name, falls back to "rainbow" if not found
theme := themes.Get("dracula", false)

// Valid checks if a theme name exists
if themes.Valid("mytheme") { ... }

// Names returns all registered theme names
names := themes.Names()  // []string{"rainbow", "dracula", ...}
```

### ThemeFunc Signature

```go
type ThemeFunc func(noColor bool) Theme
```

The `noColor` parameter is `true` when:
- The `NO_COLOR` environment variable is set
- The user has explicitly requested no colors in config

## Contributing

When contributing a new theme:

1. Follow the naming convention: lowercase, single word (e.g., `dracula`, `nord`)
2. Add a descriptive comment above your theme function
3. Ensure all 9 style elements are defined
4. Handle `noColor` parameter properly
5. Test your theme with various terminal emulators
6. Update this README's "Available Themes" table

## License

Themes are part of the Tunez project and are licensed under the same terms as the main project.
