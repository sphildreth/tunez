package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
