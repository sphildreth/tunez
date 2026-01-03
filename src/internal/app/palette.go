package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// PaletteState holds the command palette state.
type PaletteState struct {
	input    string
	cursor   int
	matches  []fuzzy.Match
	selected int
	registry *CommandRegistry
}

// NewPaletteState creates a new palette state.
func NewPaletteState(registry *CommandRegistry) *PaletteState {
	return &PaletteState{
		registry: registry,
		matches:  make([]fuzzy.Match, 0),
	}
}

// Reset clears the palette state.
func (p *PaletteState) Reset() {
	p.input = ""
	p.cursor = 0
	p.matches = nil
	p.selected = 0
}

// SetInput sets the search input and updates matches.
func (p *PaletteState) SetInput(input string) {
	p.input = input
	p.cursor = len(input)
	p.updateMatches()
}

// Input returns the current input.
func (p *PaletteState) Input() string {
	return p.input
}

// InsertChar adds a character at the cursor position.
func (p *PaletteState) InsertChar(ch rune) {
	p.input = p.input[:p.cursor] + string(ch) + p.input[p.cursor:]
	p.cursor++
	p.updateMatches()
}

// Backspace removes the character before the cursor.
func (p *PaletteState) Backspace() {
	if p.cursor > 0 {
		p.input = p.input[:p.cursor-1] + p.input[p.cursor:]
		p.cursor--
		p.updateMatches()
	}
}

// Delete removes the character at the cursor.
func (p *PaletteState) Delete() {
	if p.cursor < len(p.input) {
		p.input = p.input[:p.cursor] + p.input[p.cursor+1:]
		p.updateMatches()
	}
}

// CursorLeft moves the cursor left.
func (p *PaletteState) CursorLeft() {
	if p.cursor > 0 {
		p.cursor--
	}
}

// CursorRight moves the cursor right.
func (p *PaletteState) CursorRight() {
	if p.cursor < len(p.input) {
		p.cursor++
	}
}

// SelectUp moves selection up.
func (p *PaletteState) SelectUp() {
	if p.selected > 0 {
		p.selected--
	}
}

// SelectDown moves selection down.
func (p *PaletteState) SelectDown() {
	maxIdx := len(p.matches) - 1
	if p.input == "" {
		maxIdx = len(p.registry.commands) - 1
	}
	if p.selected < maxIdx {
		p.selected++
	}
}

// SelectedCommand returns the currently selected command.
func (p *PaletteState) SelectedCommand() *Command {
	if p.input == "" {
		// Show all commands when no input
		if p.selected < len(p.registry.commands) {
			return &p.registry.commands[p.selected]
		}
		return nil
	}

	if p.selected < len(p.matches) {
		idx := p.matches[p.selected].Index
		return &p.registry.commands[idx]
	}
	return nil
}

func (p *PaletteState) updateMatches() {
	if p.input == "" {
		p.matches = nil
		p.selected = 0
		return
	}

	names := p.registry.SearchableNames()
	p.matches = fuzzy.Find(p.input, names)
	p.selected = 0
}

// Render renders the command palette overlay.
func (p *PaletteState) Render(m *Model) string {
	var b strings.Builder

	// Title
	b.WriteString(m.theme.Title.Render("  ═══ Command Palette ═══  "))
	b.WriteString("\n\n")

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(40)

	// Show cursor in input
	inputDisplay := p.input
	if p.cursor < len(p.input) {
		inputDisplay = p.input[:p.cursor] + "│" + p.input[p.cursor:]
	} else {
		inputDisplay = p.input + "│"
	}
	b.WriteString(inputStyle.Render(inputDisplay))
	b.WriteString("\n\n")

	// Results list
	var items []Command
	var matchIndices [][]int

	if p.input == "" {
		// Show all commands grouped by category
		items = p.registry.commands
	} else {
		// Show fuzzy matches
		for _, match := range p.matches {
			items = append(items, p.registry.commands[match.Index])
			matchIndices = append(matchIndices, match.MatchedIndexes)
		}
	}

	if len(items) == 0 && p.input != "" {
		b.WriteString(m.theme.Dim.Render("  No matching commands"))
		b.WriteString("\n")
	}

	// Limit display to 10 items
	maxDisplay := 10
	startIdx := 0
	if p.selected >= maxDisplay {
		startIdx = p.selected - maxDisplay + 1
	}
	endIdx := startIdx + maxDisplay
	if endIdx > len(items) {
		endIdx = len(items)
	}

	currentCategory := ""
	for i := startIdx; i < endIdx; i++ {
		cmd := items[i]

		// Show category header (only when showing all commands)
		if p.input == "" && cmd.Category != currentCategory {
			currentCategory = cmd.Category
			b.WriteString(m.theme.Accent.Render("  " + currentCategory))
			b.WriteString("\n")
		}

		// Highlight selected item
		prefix := "   "
		if i == p.selected {
			prefix = m.theme.Highlight.Render(" ▸ ")
		}

		// Command name with fuzzy match highlighting
		name := cmd.Name
		if len(matchIndices) > i-startIdx && p.input != "" {
			name = highlightMatches(cmd.Name, matchIndices[i-startIdx], m.theme.Accent)
		}

		// Keybinding hint
		keyHint := ""
		if cmd.Keybinding != "" {
			keyHint = m.theme.Dim.Render(fmt.Sprintf(" [%s]", cmd.Keybinding))
		}

		if i == p.selected {
			b.WriteString(prefix + m.theme.Text.Bold(true).Render(name) + keyHint)
		} else {
			b.WriteString(prefix + m.theme.Text.Render(name) + keyHint)
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("  ↑↓ navigate  Enter select  Esc close"))

	// Wrap in a box
	content := b.String()
	paletteBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, paletteBox)
}

// highlightMatches highlights matched characters in a string.
func highlightMatches(s string, indices []int, style lipgloss.Style) string {
	if len(indices) == 0 {
		return s
	}

	// Build a set of matched indices
	matchSet := make(map[int]bool)
	for _, idx := range indices {
		matchSet[idx] = true
	}

	var result strings.Builder
	for i, ch := range s {
		if matchSet[i] {
			result.WriteString(style.Render(string(ch)))
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
