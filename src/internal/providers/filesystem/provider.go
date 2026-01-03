package filesystem

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/dhowden/tag"
	"github.com/tunez/tunez/internal/logging"
	"github.com/tunez/tunez/internal/provider"
	_ "modernc.org/sqlite"
)

var allowedExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".ogg":  true,
	".wav":  true,
	".opus": true,
}

type Config struct {
	Roots        []string
	IndexDB      string
	ScanOnInit   bool
	PageSize     int
	ScanProgress func(scanned int, current string) // optional callback for scan progress
}

type Provider struct {
	cfg Config
	db  *sql.DB
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string   { return "filesystem" }
func (p *Provider) Name() string { return "Filesystem" }

func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		provider.CapLyrics:  true,
		provider.CapArtwork: true,
	}
}

func (p *Provider) Initialize(ctx context.Context, profileCfg any) error {
	mapCfg, ok := profileCfg.(map[string]any)
	if !ok {
		return provider.ErrInvalidConfig
	}
	cfg, err := parseConfig(mapCfg)
	if err != nil {
		return err
	}
	p.cfg = cfg

	db, err := sql.Open("sqlite", cfg.IndexDB)
	if err != nil {
		return fmt.Errorf("open index db: %w", err)
	}
	p.db = db

	// Performance optimizations
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		// Non-fatal, but good to know
		slog.Warn("Failed to set WAL mode", "err", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		slog.Warn("Failed to set synchronous=NORMAL", "err", err)
	}
	if _, err := db.Exec("PRAGMA cache_size=-64000"); err != nil {
		slog.Warn("Failed to set cache_size", "err", err)
	}
	if _, err := db.Exec("PRAGMA temp_store=MEMORY"); err != nil {
		slog.Warn("Failed to set temp_store", "err", err)
	}
	if _, err := db.Exec("PRAGMA mmap_size=268435456"); err != nil {
		slog.Warn("Failed to set mmap_size", "err", err)
	}

	if err := p.ensureSchema(ctx); err != nil {
		return err
	}
	shouldScan := cfg.ScanOnInit
	if !shouldScan {
		// Check if DB is empty
		var count int
		if err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks").Scan(&count); err != nil {
			// If error, assume empty or broken, safe to try scan
			shouldScan = true
		} else if count == 0 {
			shouldScan = true
		}
	}

	if shouldScan {
		if err := p.scan(ctx); err != nil {
			return err
		}
	}
	return nil
}

func parseConfig(raw map[string]any) (Config, error) {
	cfg := Config{PageSize: 100, ScanOnInit: false}
	if v, ok := raw["roots"].([]any); ok {
		for _, r := range v {
			if s, ok := r.(string); ok {
				cfg.Roots = append(cfg.Roots, s)
			}
		}
	}
	if v, ok := raw["index_db"].(string); ok && v != "" {
		cfg.IndexDB = v
	}
	if v, ok := raw["scan_on_start"].(bool); ok {
		cfg.ScanOnInit = v
	}
	if v, ok := raw["scan_on_init"].(bool); ok {
		cfg.ScanOnInit = v
	}
	if v, ok := raw["page_size"].(int64); ok && v > 0 {
		cfg.PageSize = int(v)
	}
	if cb, ok := raw["scan_progress"].(func(int, string)); ok {
		cfg.ScanProgress = cb
	}
	if cfg.IndexDB == "" {
		stateDir, err := logging.StateDir()
		if err != nil {
			stateDir = os.TempDir()
		}
		// Ensure state directory exists
		if err := os.MkdirAll(stateDir, 0o755); err != nil {
			return Config{}, fmt.Errorf("create state dir: %w", err)
		}
		cfg.IndexDB = filepath.Join(stateDir, "filesystem.sqlite")
	}
	for i, r := range cfg.Roots {
		abs, err := filepath.Abs(r)
		if err != nil {
			return Config{}, err
		}
		cfg.Roots[i] = abs
	}
	return cfg, nil
}

func (p *Provider) ensureSchema(ctx context.Context) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS artists (id TEXT PRIMARY KEY, name TEXT NOT NULL, sort_name TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS albums (id TEXT PRIMARY KEY, artist_id TEXT NOT NULL, title TEXT NOT NULL, year INTEGER, artwork_path TEXT, FOREIGN KEY(artist_id) REFERENCES artists(id));`,
		`CREATE TABLE IF NOT EXISTS tracks (id TEXT PRIMARY KEY, album_id TEXT NOT NULL, artist_id TEXT NOT NULL, title TEXT NOT NULL, album_title TEXT NOT NULL, artist_name TEXT NOT NULL, year INTEGER, track_number INTEGER, disc_number INTEGER, duration_ms INTEGER, file_path TEXT NOT NULL UNIQUE, file_size INTEGER, file_mtime INTEGER, codec TEXT, bitrate INTEGER, FOREIGN KEY(album_id) REFERENCES albums(id), FOREIGN KEY(artist_id) REFERENCES artists(id));`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_album ON tracks(album_id, disc_number, track_number);`,
		`CREATE INDEX IF NOT EXISTS idx_albums_artist ON albums(artist_id, year, title);`,
		`CREATE INDEX IF NOT EXISTS idx_artists_sort ON artists(sort_name);`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_title ON tracks(title);`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_artist_name ON tracks(artist_name);`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_album_title ON tracks(album_title);`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_file_path ON tracks(file_path);`,
	}
	for _, stmt := range schema {
		if _, err := p.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate schema: %w", err)
		}
	}
	return nil
}

func hash(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// extractYear extracts year from ID3 tags with priority: TORY > TDOR > TDRL > YEAR
// TORY = Original Release Year (ID3v2.3)
// TDOR = Original Release Date (ID3v2.4)
// TDRL = Release Date (ID3v2.4)
// YEAR = Year from standard tag interface
func extractYear(meta tag.Metadata) int {
	raw := meta.Raw()
	if raw != nil {
		// Priority 1: TORY (Original Release Year - ID3v2.3)
		if v, ok := raw["TORY"]; ok {
			if year := parseYearValue(v); year > 0 {
				return year
			}
		}
		// Priority 2: TDOR (Original Release Date - ID3v2.4)
		if v, ok := raw["TDOR"]; ok {
			if year := parseYearValue(v); year > 0 {
				return year
			}
		}
		// Priority 3: TDRL (Release Date - ID3v2.4)
		if v, ok := raw["TDRL"]; ok {
			if year := parseYearValue(v); year > 0 {
				return year
			}
		}
		// Priority 4: YEAR (standard tag)
		if v, ok := raw["YEAR"]; ok {
			if year := parseYearValue(v); year > 0 {
				return year
			}
		}
	}
	// No fallback - return 0 if none of the specified tags are found
	return 0
}

// parseYearValue extracts a 4-digit year from various tag value formats
func parseYearValue(v any) int {
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case int:
		return val
	case int64:
		return int(val)
	default:
		s = fmt.Sprintf("%v", val)
	}
	// Try to extract 4-digit year from start of string (handles "2024-01-15" format)
	if len(s) >= 4 {
		year := 0
		for i := 0; i < 4 && i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				break
			}
			year = year*10 + int(c-'0')
		}
		if year >= 1900 && year <= 2100 {
			return year
		}
	}
	return 0
}

// trackInfo holds extracted metadata for a track
type trackInfo struct {
	Path        string
	Size        int64
	Mtime       int64
	ArtistName  string
	AlbumTitle  string
	TrackTitle  string
	TrackNo     int
	DiscNo      int
	Year        int
	DurationMs  int
	BitrateKbps int
	Codec       string
}

func (p *Provider) scan(ctx context.Context) error {
	// 1. Load existing tracks for incremental scan
	existing := make(map[string]struct {
		mtime int64
		size  int64
	})
	rows, err := p.db.QueryContext(ctx, "SELECT file_path, file_mtime, file_size FROM tracks")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var path string
			var mtime, size int64
			if err := rows.Scan(&path, &mtime, &size); err == nil {
				existing[path] = struct{ mtime, size int64 }{mtime, size}
			}
		}
	}

	// 2. Setup worker pool
	jobs := make(chan string, 100)
	results := make(chan *trackInfo, 100)
	var wg sync.WaitGroup

	// Start workers
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				if ctx.Err() != nil {
					return
				}

				info, err := os.Stat(path)
				if err != nil {
					continue
				}

				// Check if unchanged
				if e, ok := existing[path]; ok {
					if e.mtime == info.ModTime().Unix() && e.size == info.Size() {
						// Signal unchanged by sending nil info but with path?
						// Or just send the path to mark as seen.
						// We'll use a special struct or just re-emit the existing data?
						// Re-emitting existing data is safest but requires querying it.
						// Since we only have mtime/size in memory, we can't re-emit.
						// We should just signal "unchanged" so the collector knows to keep it.
						results <- &trackInfo{Path: path, Mtime: -1} // Mtime -1 indicates unchanged
						continue
					}
				}

				// Process new/changed file
				ti, err := p.processFile(path, info)
				if err != nil {
					continue
				}
				results <- ti
			}
		}()
	}

	// 3. Start collector (database writer)
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)

		// Cache known IDs to avoid redundant DB executions
		knownArtists := make(map[string]bool)
		knownAlbums := make(map[string]bool)

		// Set PRAGMAs before starting transaction
		if _, err := p.db.ExecContext(ctx, "PRAGMA synchronous=OFF"); err != nil {
			slog.Warn("Failed to set synchronous=OFF", "err", err)
		}

		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			errChan <- err
			return
		}
		defer tx.Rollback()

		insertArtist, _ := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO artists(id,name,sort_name) VALUES(?,?,?)`)
		insertAlbum, _ := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO albums(id,artist_id,title,year,artwork_path) VALUES(?,?,?,?,?)`)
		insertTrack, _ := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO tracks(id,album_id,artist_id,title,album_title,artist_name,year,track_number,disc_number,duration_ms,file_path,file_size,file_mtime,codec,bitrate) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

		seenPaths := make(map[string]bool)
		batchSize := 100
		count := 0
		scanned := 0

		for ti := range results {
			seenPaths[ti.Path] = true
			scanned++
			if p.cfg.ScanProgress != nil && scanned%10 == 0 {
				p.cfg.ScanProgress(scanned, ti.Path)
			}

			if ti.Mtime == -1 {
				// Unchanged, nothing to update in DB
				continue
			}

			// Insert/Update logic
			artistID := hash(strings.ToLower(ti.ArtistName))
			albumID := hash(artistID, strings.ToLower(ti.AlbumTitle))
			trackID := hash(ti.Path)

			if !knownArtists[artistID] {
				if _, err := insertArtist.ExecContext(ctx, artistID, ti.ArtistName, strings.ToLower(ti.ArtistName)); err != nil {
					continue
				}
				knownArtists[artistID] = true
			}

			if !knownAlbums[albumID] {
				if _, err := insertAlbum.ExecContext(ctx, albumID, artistID, ti.AlbumTitle, ti.Year, ""); err != nil {
					continue
				}
				knownAlbums[albumID] = true
			}

			if _, err := insertTrack.ExecContext(ctx, trackID, albumID, artistID, ti.TrackTitle, ti.AlbumTitle, ti.ArtistName, ti.Year, ti.TrackNo, ti.DiscNo, ti.DurationMs, ti.Path, ti.Size, ti.Mtime, ti.Codec, ti.BitrateKbps); err != nil {
				continue
			}

			count++
			if count >= batchSize {
				if err := tx.Commit(); err != nil {
					errChan <- err
					return
				}
				// Start new transaction
				tx, err = p.db.BeginTx(ctx, nil)
				if err != nil {
					errChan <- err
					return
				}

				insertArtist, _ = tx.PrepareContext(ctx, `INSERT OR IGNORE INTO artists(id,name,sort_name) VALUES(?,?,?)`)
				insertAlbum, _ = tx.PrepareContext(ctx, `INSERT OR IGNORE INTO albums(id,artist_id,title,year,artwork_path) VALUES(?,?,?,?,?)`)
				insertTrack, _ = tx.PrepareContext(ctx, `INSERT OR REPLACE INTO tracks(id,album_id,artist_id,title,album_title,artist_name,year,track_number,disc_number,duration_ms,file_path,file_size,file_mtime,codec,bitrate) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
				count = 0
			}
		}

		// Cleanup deleted files
		for path := range existing {
			if !seenPaths[path] {
				// File no longer exists or wasn't scanned
				_, _ = tx.ExecContext(ctx, "DELETE FROM tracks WHERE file_path = ?", path)
			}
		}

		if err := tx.Commit(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// 4. Walk directories and feed jobs
	for _, root := range p.cfg.Roots {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !allowedExtensions[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
			select {
			case jobs <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}
	close(jobs)
	wg.Wait()
	close(results)

	// Wait for collector
	if err := <-errChan; err != nil {
		return err
	}

	// Optimize DB after scan
	if _, err := p.db.Exec("PRAGMA optimize"); err != nil {
		slog.Warn("Failed to run PRAGMA optimize", "err", err)
	}

	return nil
}

func (p *Provider) processFile(path string, info os.FileInfo) (*trackInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ti := &trackInfo{
		Path:  path,
		Size:  info.Size(),
		Mtime: info.ModTime().Unix(),
	}

	meta, err := tag.ReadFrom(f)
	if err == nil {
		ti.ArtistName = meta.Artist()
		ti.AlbumTitle = meta.Album()
		ti.TrackTitle = meta.Title()
		ti.TrackNo, _ = meta.Track()
		ti.DiscNo, _ = meta.Disc()
		ti.Year = extractYear(meta)
	}

	if ti.ArtistName == "" {
		ti.ArtistName = "Unknown Artist"
	}
	if ti.AlbumTitle == "" {
		ti.AlbumTitle = filepath.Base(filepath.Dir(path))
		if ti.AlbumTitle == "." || ti.AlbumTitle == "/" {
			ti.AlbumTitle = "Unknown Album"
		}
	}
	if ti.TrackTitle == "" {
		ti.TrackTitle = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Get audio metadata
	audioInfo := getAudioInfo(path)
	ti.DurationMs = audioInfo.DurationMs
	ti.Codec = audioInfo.Codec
	ti.BitrateKbps = audioInfo.BitrateKbps

	return ti, nil
}

func (p *Provider) Health(ctx context.Context) (bool, string) {
	if p.db == nil {
		return false, "db not initialized"
	}
	if err := p.db.PingContext(ctx); err != nil {
		return false, err.Error()
	}
	return true, "ok"
}

func (p *Provider) ListArtists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Artist], error) {
	return p.listArtists(ctx, req)
}

func (p *Provider) listArtists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Artist], error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	_, offset := parseCursor(req.Cursor)
	rows, err := p.db.QueryContext(ctx, `
		SELECT a.id, a.name, a.sort_name, COUNT(al.id) as album_count
		FROM artists a
		LEFT JOIN albums al ON al.artist_id = a.id
		GROUP BY a.id
		ORDER BY a.sort_name
		LIMIT ? OFFSET ?`, pageSize+1, offset)
	if err != nil {
		return provider.Page[provider.Artist]{}, err
	}
	defer rows.Close()
	var items []provider.Artist
	for rows.Next() {
		var a provider.Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.SortName, &a.AlbumCount); err != nil {
			return provider.Page[provider.Artist]{}, err
		}
		items = append(items, a)
	}
	next := ""
	if len(items) > pageSize {
		next = fmt.Sprintf("%d", offset+pageSize)
		items = items[:pageSize]
	}
	return provider.Page[provider.Artist]{Items: items, NextCursor: next, TotalHint: -1}, nil
}

func (p *Provider) GetArtist(ctx context.Context, id string) (provider.Artist, error) {
	var a provider.Artist
	err := p.db.QueryRowContext(ctx, `
		SELECT a.id, a.name, a.sort_name, COUNT(al.id) as album_count
		FROM artists a
		LEFT JOIN albums al ON al.artist_id = a.id
		WHERE a.id = ?
		GROUP BY a.id`, id).Scan(&a.ID, &a.Name, &a.SortName, &a.AlbumCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.Artist{}, provider.ErrNotFound
		}
		return provider.Artist{}, err
	}
	return a, nil
}

func (p *Provider) ListAlbums(ctx context.Context, artistId string, req provider.ListReq) (provider.Page[provider.Album], error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	_, offset := parseCursor(req.Cursor)
	query := `SELECT id,artist_id,title,year FROM albums `
	var args []any
	if artistId != "" {
		query += `WHERE artist_id=? `
		args = append(args, artistId)
	}
	query += `ORDER BY title LIMIT ? OFFSET ?`
	args = append(args, pageSize+1, offset)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return provider.Page[provider.Album]{}, err
	}
	defer rows.Close()
	var items []provider.Album
	for rows.Next() {
		var a provider.Album
		if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.Year); err != nil {
			return provider.Page[provider.Album]{}, err
		}
		items = append(items, a)
	}
	next := ""
	if len(items) > pageSize {
		next = fmt.Sprintf("%d", offset+pageSize)
		items = items[:pageSize]
	}
	return provider.Page[provider.Album]{Items: items, NextCursor: next, TotalHint: -1}, nil
}

func (p *Provider) GetAlbum(ctx context.Context, id string) (provider.Album, error) {
	var a provider.Album
	err := p.db.QueryRowContext(ctx, `SELECT id,artist_id,title,year FROM albums WHERE id=?`, id).Scan(&a.ID, &a.ArtistID, &a.Title, &a.Year)
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.Album{}, provider.ErrNotFound
		}
		return provider.Album{}, err
	}
	return a, nil
}

func (p *Provider) ListTracks(ctx context.Context, albumId string, artistId string, playlistId string, req provider.ListReq) (provider.Page[provider.Track], error) {
	if playlistId != "" {
		return provider.Page[provider.Track]{}, provider.ErrNotSupported
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	_, offset := parseCursor(req.Cursor)
	query := `SELECT id,title,artist_id,artist_name,album_id,album_title,year,duration_ms,track_number,disc_number,codec,bitrate,file_path FROM tracks `
	var args []any
	var clauses []string
	if albumId != "" {
		clauses = append(clauses, "album_id=?")
		args = append(args, albumId)
	}
	if artistId != "" {
		clauses = append(clauses, "artist_id=?")
		args = append(args, artistId)
	}
	if len(clauses) > 0 {
		query += "WHERE " + strings.Join(clauses, " AND ") + " "
	}
	query += `ORDER BY disc_number, track_number, title LIMIT ? OFFSET ?`
	args = append(args, pageSize+1, offset)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return provider.Page[provider.Track]{}, err
	}
	defer rows.Close()
	var items []provider.Track
	for rows.Next() {
		var t provider.Track
		var filePath string
		if err := rows.Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.Year, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps, &filePath); err != nil {
			return provider.Page[provider.Track]{}, err
		}
		t.ArtworkRef = filePath // Use file path for artwork extraction
		items = append(items, t)
	}
	next := ""
	if len(items) > pageSize {
		next = fmt.Sprintf("%d", offset+pageSize)
		items = items[:pageSize]
	}
	return provider.Page[provider.Track]{Items: items, NextCursor: next, TotalHint: -1}, nil
}

func (p *Provider) GetTrack(ctx context.Context, id string) (provider.Track, error) {
	var t provider.Track
	var filePath string
	err := p.db.QueryRowContext(ctx, `SELECT id,title,artist_id,artist_name,album_id,album_title,year,duration_ms,track_number,disc_number,codec,bitrate,file_path FROM tracks WHERE id=?`, id).Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.Year, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps, &filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.Track{}, provider.ErrNotFound
		}
		return provider.Track{}, err
	}
	t.ArtworkRef = filePath // Use file path for artwork extraction
	return t, nil
}

func (p *Provider) Search(ctx context.Context, q string, req provider.ListReq) (provider.SearchResults, error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	targetType, offset := parseCursor(req.Cursor)
	pattern := "%" + strings.ToLower(q) + "%"

	var res provider.SearchResults

	// Search Tracks
	if targetType == "" || targetType == "tracks" {
		rows, err := p.db.QueryContext(ctx, `SELECT id,title,artist_id,artist_name,album_id,album_title,year,duration_ms,track_number,disc_number,codec,bitrate,file_path FROM tracks WHERE lower(title) LIKE ? OR lower(artist_name) LIKE ? OR lower(album_title) LIKE ? ORDER BY artist_name LIMIT ? OFFSET ?`, pattern, pattern, pattern, pageSize+1, offset)
		if err != nil {
			return provider.SearchResults{}, err
		}
		defer rows.Close()
		var tracks []provider.Track
		for rows.Next() {
			var t provider.Track
			var filePath string
			if err := rows.Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.Year, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps, &filePath); err != nil {
				return provider.SearchResults{}, err
			}
			t.ArtworkRef = filePath // Use file path for artwork extraction
			tracks = append(tracks, t)
		}
		next := ""
		if len(tracks) > pageSize {
			next = fmt.Sprintf("tracks:%d", offset+pageSize)
			tracks = tracks[:pageSize]
		}
		res.Tracks = provider.Page[provider.Track]{Items: tracks, NextCursor: next, TotalHint: -1}
	}

	// Search Albums
	if targetType == "" || targetType == "albums" {
		rows, err := p.db.QueryContext(ctx, `SELECT id,artist_id,title,year FROM albums WHERE lower(title) LIKE ? ORDER BY title LIMIT ? OFFSET ?`, pattern, pageSize+1, offset)
		if err != nil {
			return provider.SearchResults{}, err
		}
		defer rows.Close()
		var albums []provider.Album
		for rows.Next() {
			var a provider.Album
			if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.Year); err != nil {
				return provider.SearchResults{}, err
			}
			albums = append(albums, a)
		}
		next := ""
		if len(albums) > pageSize {
			next = fmt.Sprintf("albums:%d", offset+pageSize)
			albums = albums[:pageSize]
		}
		res.Albums = provider.Page[provider.Album]{Items: albums, NextCursor: next, TotalHint: -1}
	}

	// Search Artists
	if targetType == "" || targetType == "artists" {
		rows, err := p.db.QueryContext(ctx, `
			SELECT a.id, a.name, a.sort_name, COUNT(al.id) as album_count
			FROM artists a
			LEFT JOIN albums al ON al.artist_id = a.id
			WHERE lower(a.name) LIKE ?
			GROUP BY a.id
			ORDER BY a.sort_name
			LIMIT ? OFFSET ?`, pattern, pageSize+1, offset)
		if err != nil {
			return provider.SearchResults{}, err
		}
		defer rows.Close()
		var artists []provider.Artist
		for rows.Next() {
			var a provider.Artist
			if err := rows.Scan(&a.ID, &a.Name, &a.SortName, &a.AlbumCount); err != nil {
				return provider.SearchResults{}, err
			}
			artists = append(artists, a)
		}
		next := ""
		if len(artists) > pageSize {
			next = fmt.Sprintf("artists:%d", offset+pageSize)
			artists = artists[:pageSize]
		}
		res.Artists = provider.Page[provider.Artist]{Items: artists, NextCursor: next, TotalHint: -1}
	}

	return res, nil
}

func (p *Provider) ListPlaylists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Playlist], error) {
	return provider.Page[provider.Playlist]{}, provider.ErrNotSupported
}

func (p *Provider) GetPlaylist(ctx context.Context, id string) (provider.Playlist, error) {
	return provider.Playlist{}, provider.ErrNotSupported
}

func (p *Provider) GetStream(ctx context.Context, trackId string) (provider.StreamInfo, error) {
	var path string
	err := p.db.QueryRowContext(ctx, `SELECT file_path FROM tracks WHERE id=?`, trackId).Scan(&path)
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.StreamInfo{}, provider.ErrNotFound
		}
		return provider.StreamInfo{}, err
	}
	if _, err := os.Stat(path); err != nil {
		return provider.StreamInfo{}, fmt.Errorf("track missing: %w", err)
	}
	u := url.URL{Scheme: "file", Path: path}
	return provider.StreamInfo{URL: u.String()}, nil
}

func (p *Provider) GetLyrics(ctx context.Context, trackId string) (provider.Lyrics, error) {
	// Get file path for track
	var filePath string
	err := p.db.QueryRowContext(ctx, `SELECT file_path FROM tracks WHERE id=?`, trackId).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.Lyrics{}, provider.ErrNotFound
		}
		return provider.Lyrics{}, err
	}

	// Try embedded lyrics from ID3 tags first
	lyrics, err := extractEmbeddedLyrics(filePath)
	if err == nil && lyrics != "" {
		return provider.Lyrics{Text: lyrics}, nil
	}

	// Try .lrc sidecar file
	lrcPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".lrc"
	if lrcData, err := os.ReadFile(lrcPath); err == nil {
		return provider.Lyrics{Text: string(lrcData)}, nil
	}

	// Try .txt sidecar file
	txtPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".txt"
	if txtData, err := os.ReadFile(txtPath); err == nil {
		return provider.Lyrics{Text: string(txtData)}, nil
	}

	return provider.Lyrics{}, provider.ErrNotFound
}

// extractEmbeddedLyrics reads lyrics from ID3v2 USLT frame or similar tags.
func extractEmbeddedLyrics(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		return "", err
	}

	// The tag library provides access to raw frames
	raw := meta.Raw()
	if raw == nil {
		return "", fmt.Errorf("no raw tags")
	}

	// Check for USLT (Unsynchronized Lyrics) frame - ID3v2
	if uslt, ok := raw["USLT"]; ok {
		if s, ok := uslt.(string); ok && s != "" {
			return s, nil
		}
	}

	// Check for LYRICS tag (common in Vorbis comments)
	if lyrics, ok := raw["LYRICS"]; ok {
		if s, ok := lyrics.(string); ok && s != "" {
			return s, nil
		}
	}

	// Check for UNSYNCEDLYRICS (Vorbis)
	if lyrics, ok := raw["UNSYNCEDLYRICS"]; ok {
		if s, ok := lyrics.(string); ok && s != "" {
			return s, nil
		}
	}

	return "", fmt.Errorf("no lyrics tag found")
}

func (p *Provider) GetArtwork(ctx context.Context, ref string, sizePx int) (provider.Artwork, error) {
	// ref is the file path for filesystem provider
	if ref == "" {
		return provider.Artwork{}, provider.ErrNotFound
	}

	// Try to extract embedded artwork from the audio file
	f, err := os.Open(ref)
	if err != nil {
		return provider.Artwork{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		// Try folder.jpg fallback
		return p.getFolderArtwork(ref)
	}

	pic := meta.Picture()
	if pic == nil || len(pic.Data) == 0 {
		// Try folder.jpg fallback
		return p.getFolderArtwork(ref)
	}

	mimeType := pic.MIMEType
	if mimeType == "" {
		// Guess from data
		if len(pic.Data) > 2 && pic.Data[0] == 0xFF && pic.Data[1] == 0xD8 {
			mimeType = "image/jpeg"
		} else if len(pic.Data) > 8 && string(pic.Data[1:4]) == "PNG" {
			mimeType = "image/png"
		} else {
			mimeType = "image/jpeg" // Default
		}
	}

	return provider.Artwork{
		Data:     pic.Data,
		MimeType: mimeType,
	}, nil
}

// getFolderArtwork looks for folder.jpg, cover.jpg, etc. in the same directory
func (p *Provider) getFolderArtwork(trackPath string) (provider.Artwork, error) {
	dir := filepath.Dir(trackPath)
	coverNames := []string{"folder.jpg", "cover.jpg", "album.jpg", "front.jpg", "folder.png", "cover.png", "album.png", "front.png"}

	for _, name := range coverNames {
		coverPath := filepath.Join(dir, name)
		data, err := os.ReadFile(coverPath)
		if err == nil && len(data) > 0 {
			mimeType := "image/jpeg"
			if strings.HasSuffix(strings.ToLower(name), ".png") {
				mimeType = "image/png"
			}
			return provider.Artwork{
				Data:     data,
				MimeType: mimeType,
			}, nil
		}
	}

	return provider.Artwork{}, provider.ErrNotFound
}

func parseCursor(cur string) (string, int) {
	if cur == "" {
		return "", 0
	}
	parts := strings.SplitN(cur, ":", 2)
	if len(parts) == 2 {
		var off int
		fmt.Sscanf(parts[1], "%d", &off)
		return parts[0], off
	}
	var off int
	fmt.Sscanf(cur, "%d", &off)
	return "", off
}

// getDurationMs uses ffprobe to get audio duration in milliseconds
func getDurationMs(path string) int {
	info := getAudioInfo(path)
	return info.DurationMs
}

// audioInfo holds metadata extracted from ffprobe
type audioInfo struct {
	DurationMs  int
	Codec       string
	BitrateKbps int
}

// getAudioInfo uses ffprobe to extract audio metadata
func getAudioInfo(path string) audioInfo {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		return audioInfo{}
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecName string `json:"codec_name"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return audioInfo{}
	}

	info := audioInfo{}

	// Duration
	if result.Format.Duration != "" {
		var secs float64
		fmt.Sscanf(result.Format.Duration, "%f", &secs)
		info.DurationMs = int(secs * 1000)
	}

	// Bitrate (convert from bps to kbps)
	if result.Format.BitRate != "" {
		var bps int
		fmt.Sscanf(result.Format.BitRate, "%d", &bps)
		info.BitrateKbps = bps / 1000
	}

	// Codec - find the audio stream
	for _, s := range result.Streams {
		if s.CodecType == "audio" && s.CodecName != "" {
			info.Codec = s.CodecName
			break
		}
	}

	return info
}
