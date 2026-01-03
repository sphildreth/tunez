package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhowden/tag"
	"github.com/tunez/tunez/internal/provider"
)

func TestFilesystemGetStream(t *testing.T) {
	dir := t.TempDir()
	song := filepath.Join(dir, "track.mp3")
	if err := os.WriteFile(song, []byte("fake audio"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	p := New()
	settings := map[string]any{
		"roots":         []any{dir},
		"index_db":      filepath.Join(dir, "index.sqlite"),
		"scan_on_start": true,
	}
	if err := p.Initialize(context.Background(), settings); err != nil {
		t.Fatalf("init: %v", err)
	}
	page, err := p.ListTracks(context.Background(), "", "", "", provider.ListReq{PageSize: 10})
	if err != nil {
		t.Fatalf("list tracks: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 track got %d", len(page.Items))
	}
	stream, err := p.GetStream(context.Background(), page.Items[0].ID)
	if err != nil {
		t.Fatalf("get stream: %v", err)
	}
	if !strings.HasPrefix(stream.URL, "file://") {
		t.Fatalf("expected file url got %s", stream.URL)
	}
}

// TestExtractYear tests the year extraction logic with various tag scenarios
func TestExtractYear(t *testing.T) {
	tests := []struct {
		name     string
		rawTags  map[string]any
		expected int
	}{
		{
			name:     "TORY only",
			rawTags:  map[string]any{"TORY": 1995},
			expected: 1995,
		},
		{
			name:     "TDOR only",
			rawTags:  map[string]any{"TDOR": 2001},
			expected: 2001,
		},
		{
			name:     "TDRL only",
			rawTags:  map[string]any{"TDRL": 2010},
			expected: 2010,
		},
		{
			name:     "YEAR only",
			rawTags:  map[string]any{"YEAR": 2020},
			expected: 2020,
		},
		{
			name:     "TORY takes priority over TDOR",
			rawTags:  map[string]any{"TORY": 1995, "TDOR": 2001},
			expected: 1995,
		},
		{
			name:     "TORY takes priority over TDRL",
			rawTags:  map[string]any{"TORY": 1995, "TDRL": 2010},
			expected: 1995,
		},
		{
			name:     "TORY takes priority over YEAR",
			rawTags:  map[string]any{"TORY": 1995, "YEAR": 2020},
			expected: 1995,
		},
		{
			name:     "TDOR takes priority over TDRL",
			rawTags:  map[string]any{"TDOR": 2001, "TDRL": 2010},
			expected: 2001,
		},
		{
			name:     "TDOR takes priority over YEAR",
			rawTags:  map[string]any{"TDOR": 2001, "YEAR": 2020},
			expected: 2001,
		},
		{
			name:     "TDRL takes priority over YEAR",
			rawTags:  map[string]any{"TDRL": 2010, "YEAR": 2020},
			expected: 2010,
		},
		{
			name:     "String year value",
			rawTags:  map[string]any{"TORY": "1998"},
			expected: 1998,
		},
		{
			name:     "Date string with year",
			rawTags:  map[string]any{"TORY": "1998-05-20"},
			expected: 1998,
		},
		{
			name:     "No year tags returns 0",
			rawTags:  map[string]any{},
			expected: 0,
		},
		{
			name:     "Invalid year (too old) returns 0",
			rawTags:  map[string]any{"TORY": 1800},
			expected: 0,
		},
		{
			name:     "Invalid year (too new) returns 0",
			rawTags:  map[string]any{"TORY": 2150},
			expected: 0,
		},
		{
			name:     "Current year (2026) should be valid if in tag",
			rawTags:  map[string]any{"TORY": 2026},
			expected: 2026,
		},
		{
			name:     "DATE tag should NOT be used",
			rawTags:  map[string]any{"DATE": 2025, "YEAR": 2020},
			expected: 2020,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock metadata object
			meta := &mockMetadata{raw: tt.rawTags}
			result := extractYear(meta)
			if result != tt.expected {
				t.Errorf("extractYear() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestParseYearValue tests the parseYearValue function directly
func TestParseYearValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{"int year", 1995, 1995},
		{"int64 year", int64(2001), 2001},
		{"string year", "2010", 2010},
		{"string date", "2020-05-15", 2020},
		{"string datetime", "2020-05-15T10:30:00Z", 2020},
		{"invalid string", "not a year", 0},
		{"empty string", "", 0},
		{"too old", 1800, 0},
		{"too new", 2150, 0},
		{"current year", 2026, 2026},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseYearValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseYearValue(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// mockMetadata implements tag.Metadata for testing
type mockMetadata struct {
	raw map[string]any
}

func (m *mockMetadata) Format() tag.Format {
	return tag.UnknownFormat
}

func (m *mockMetadata) FileType() tag.FileType {
	return tag.UnknownFileType
}

func (m *mockMetadata) Title() string {
	return "Test Title"
}

func (m *mockMetadata) Album() string {
	return "Test Album"
}

func (m *mockMetadata) Artist() string {
	return "Test Artist"
}

func (m *mockMetadata) AlbumArtist() string {
	return ""
}

func (m *mockMetadata) Composer() string {
	return ""
}

func (m *mockMetadata) Year() int {
	// This should NOT be used by our extractYear function
	return 2026 // Current year - if this is returned, it's a bug
}

func (m *mockMetadata) Genre() string {
	return ""
}

func (m *mockMetadata) Track() (int, int) {
	return 1, 10
}

func (m *mockMetadata) Disc() (int, int) {
	return 1, 1
}

func (m *mockMetadata) Picture() *tag.Picture {
	return nil
}

func (m *mockMetadata) Lyrics() string {
	return ""
}

func (m *mockMetadata) Comment() string {
	return ""
}

func (m *mockMetadata) Raw() map[string]any {
	return m.raw
}
