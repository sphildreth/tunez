package filesystem

import (
	"context"
	"fmt"
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

// TestProvider_Integration performs a full integration test:
// 1. Setup a temp directory with audio files
// 2. Initialize provider (triggering scan)
// 3. Verify data via List* and Search methods
func TestProvider_Integration(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	tmpDir := t.TempDir()
	filesDir := filepath.Join(tmpDir, "music")
	dbPath := filepath.Join(tmpDir, "index.sqlite")

	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Helper to create a dummy MP3 file
	createTrack := func(artist, album, title string, trackNo int) {
		dir := filepath.Join(filesDir, artist, album)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		fname := fmt.Sprintf("%02d - %s.mp3", trackNo, title)
		path := filepath.Join(dir, fname)
		// We actually don't need real MP3 content because our provider uses
		// tag.ReadFrom which handles errors gracefully or we can mock it.
		// However, the current implementation *requires* tag parsing to work well
		// for metadata, OR it falls back to filename parsing.
		// Let's rely on filename fallback logic or write minimal ID3 headers if needed.
		// For now, empty files will trigger fallback logic in processFile:
		// Artist=Unknown (or from dir?), Album=Dir, Title=Filename
		// Wait! The provider logic:
		// if ti.ArtistName == "" { ti.ArtistName = "Unknown Artist" }
		// if ti.AlbumTitle == "" { ti.AlbumTitle = dir name }
		// So we can test the fallback logic easily.
		if err := os.WriteFile(path, []byte("fake mp3 content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create structure:
	// Artist A
	//   Album 1
	//     01 - Track A.mp3
	//     02 - Track B.mp3
	// Artist B
	//   Album 2
	//     01 - Track C.mp3
	createTrack("Artist A", "Album 1", "Track A", 1)
	createTrack("Artist A", "Album 1", "Track B", 2)
	createTrack("Artist B", "Album 2", "Track C", 1)

	// 2. Initialize
	p := New()
	settings := map[string]any{
		"roots":         []any{filesDir},
		"index_db":      dbPath,
		"scan_on_start": true,
		"page_size":     int64(10),
	}

	// Use a channel to verify progress callback
	progressCh := make(chan string, 10)
	settings["scan_progress"] = func(n int, path string) {
		progressCh <- path
	}

	if err := p.Initialize(ctx, settings); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Give async scan a moment to finish?
	// Initialize waits for p.scan() if scan_on_start is true.
	// So we should be good.

	// 3. Verify Data

	// List Artists
	artists, err := p.ListArtists(ctx, provider.ListReq{})
	if err != nil {
		t.Fatalf("ListArtists failed: %v", err)
	}
	// With fallback logic:
	// Artist A directory -> ArtistName will be "Unknown Artist" because tag parsing fails on empty files?
	// Let's check the code.
	// processFile: meta, err := tag.ReadFrom(f) -> will fail
	// Fallbacks:
	// ArtistName = "Unknown Artist"
	// AlbumTitle = parent dir name ("Album 1")
	// TrackTitle = filename ("01 - Track A")
	//
	// Wait, if all artists are "Unknown Artist", they merge.
	// So we should expect 1 artist "Unknown Artist" with 2 albums "Album 1" and "Album 2".
	//
	// NOTE: This highlights that for better tests we might want a way to mock tag reading
	// or write real ID3 tags. Since writing real tags is complex without a lib,
	// let's adjust expectations to the fallback behavior, which IS worth testing.

	if len(artists.Items) == 0 {
		t.Fatal("Expected artists, got none")
	}
	if artists.Items[0].Name != "Unknown Artist" {
		// If by some miracle it parsed something else
		t.Logf("Got artist: %s", artists.Items[0].Name)
	}

	// Let's verify we have 3 tracks total
	// ListAlbums (global validation via Search or iterating)
	// Let's allow for the fact that they might be grouped under Unknown Artist

	// List Albums for Unknown Artist
	unknownId := artists.Items[0].ID
	albums, err := p.ListAlbums(ctx, unknownId, provider.ListReq{})
	if err != nil {
		t.Fatalf("ListAlbums failed: %v", err)
	}

	// We expect "Album 1" and "Album 2"
	foundAlbums := make(map[string]bool)
	for _, a := range albums.Items {
		foundAlbums[a.Title] = true
	}
	if !foundAlbums["Album 1"] {
		t.Error("Expected Album 1")
	}
	if !foundAlbums["Album 2"] {
		t.Error("Expected Album 2")
	}

	// List Tracks for Album 1
	var album1ID string
	for _, a := range albums.Items {
		if a.Title == "Album 1" {
			album1ID = a.ID
			break
		}
	}
	tracks, err := p.ListTracks(ctx, album1ID, "", "", provider.ListReq{})
	if err != nil {
		t.Fatalf("ListTracks failed: %v", err)
	}
	if len(tracks.Items) != 2 {
		t.Errorf("Expected 2 tracks for Album 1, got %d", len(tracks.Items))
	}

	// Search
	res, err := p.Search(ctx, "Track A", provider.ListReq{})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res.Tracks.Items) == 0 {
		t.Error("Search for 'Track A' returned no results")
	}
	if !strings.Contains(res.Tracks.Items[0].Title, "Track A") {
		t.Errorf("Search result title mismatch, got %s", res.Tracks.Items[0].Title)
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
