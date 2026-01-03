package filesystem

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	Roots      []string
	IndexDB    string
	ScanOnInit bool
	PageSize   int
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
	return provider.Capabilities{}
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
	if v, ok := raw["page_size"].(int64); ok && v > 0 {
		cfg.PageSize = int(v)
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
		`CREATE TABLE IF NOT EXISTS tracks (id TEXT PRIMARY KEY, album_id TEXT NOT NULL, artist_id TEXT NOT NULL, title TEXT NOT NULL, album_title TEXT NOT NULL, artist_name TEXT NOT NULL, track_number INTEGER, disc_number INTEGER, duration_ms INTEGER, file_path TEXT NOT NULL UNIQUE, file_size INTEGER, file_mtime INTEGER, codec TEXT, bitrate INTEGER, FOREIGN KEY(album_id) REFERENCES albums(id), FOREIGN KEY(artist_id) REFERENCES artists(id));`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_album ON tracks(album_id, disc_number, track_number);`,
		`CREATE INDEX IF NOT EXISTS idx_albums_artist ON albums(artist_id, year, title);`,
		`CREATE INDEX IF NOT EXISTS idx_artists_sort ON artists(sort_name);`,
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

func (p *Provider) scan(ctx context.Context) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM tracks; DELETE FROM albums; DELETE FROM artists;`); err != nil {
		return fmt.Errorf("clear tables: %w", err)
	}
	insertArtist, _ := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO artists(id,name,sort_name) VALUES(?,?,?)`)
	insertAlbum, _ := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO albums(id,artist_id,title,year,artwork_path) VALUES(?,?,?,?,?)`)
	insertTrack, _ := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO tracks(id,album_id,artist_id,title,album_title,artist_name,track_number,disc_number,duration_ms,file_path,file_size,file_mtime,codec,bitrate) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	for _, root := range p.cfg.Roots {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !allowedExtensions[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer f.Close()
			meta, err := tag.ReadFrom(f)
			var artistName, albumTitle, trackTitle string
			var trackNo, discNo int
			if err == nil {
				artistName = meta.Artist()
				albumTitle = meta.Album()
				trackTitle = meta.Title()
				trackNo, _ = meta.Track()
				discNo, _ = meta.Disc()
			}
			if artistName == "" {
				artistName = "Unknown Artist"
			}
			if albumTitle == "" {
				albumTitle = filepath.Base(filepath.Dir(path))
				if albumTitle == "." || albumTitle == "/" {
					albumTitle = "Unknown Album"
				}
			}
			if trackTitle == "" {
				trackTitle = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			artistID := hash(strings.ToLower(artistName))
			albumID := hash(artistID, strings.ToLower(albumTitle))
			trackID := hash(path)
			_, _ = insertArtist.ExecContext(ctx, artistID, artistName, strings.ToLower(artistName))
			_, _ = insertAlbum.ExecContext(ctx, albumID, artistID, albumTitle, 0, "")
			// Skip ffprobe during scan for speed - duration will be 0 initially
			// TODO: Add background job to populate durations, or get on-demand
			durationMs := 0
			format := ""
			if err == nil {
				format = fmt.Sprint(meta.Format())
			}
			_, _ = insertTrack.ExecContext(ctx, trackID, albumID, artistID, trackTitle, albumTitle, artistName, trackNo, discNo, durationMs, path, info.Size(), info.ModTime().Unix(), format, 0)
			return nil
		})
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit scan: %w", err)
	}
	return nil
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
	rows, err := p.db.QueryContext(ctx, `SELECT id,name,sort_name FROM artists ORDER BY sort_name LIMIT ? OFFSET ?`, pageSize+1, offset)
	if err != nil {
		return provider.Page[provider.Artist]{}, err
	}
	defer rows.Close()
	var items []provider.Artist
	for rows.Next() {
		var a provider.Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.SortName); err != nil {
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
	err := p.db.QueryRowContext(ctx, `SELECT id,name,sort_name FROM artists WHERE id=?`, id).Scan(&a.ID, &a.Name, &a.SortName)
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
	query := `SELECT id,title,artist_id,artist_name,album_id,album_title,duration_ms,track_number,disc_number,codec,bitrate FROM tracks `
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
		if err := rows.Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps); err != nil {
			return provider.Page[provider.Track]{}, err
		}
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
	err := p.db.QueryRowContext(ctx, `SELECT id,title,artist_id,artist_name,album_id,album_title,duration_ms,track_number,disc_number,codec,bitrate,file_path FROM tracks WHERE id=?`, id).Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps, new(string))
	if err != nil {
		if err == sql.ErrNoRows {
			return provider.Track{}, provider.ErrNotFound
		}
		return provider.Track{}, err
	}
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
		rows, err := p.db.QueryContext(ctx, `SELECT id,title,artist_id,artist_name,album_id,album_title,duration_ms,track_number,disc_number,codec,bitrate FROM tracks WHERE lower(title) LIKE ? OR lower(artist_name) LIKE ? OR lower(album_title) LIKE ? ORDER BY artist_name LIMIT ? OFFSET ?`, pattern, pattern, pattern, pageSize+1, offset)
		if err != nil {
			return provider.SearchResults{}, err
		}
		defer rows.Close()
		var tracks []provider.Track
		for rows.Next() {
			var t provider.Track
			if err := rows.Scan(&t.ID, &t.Title, &t.ArtistID, &t.ArtistName, &t.AlbumID, &t.AlbumTitle, &t.DurationMs, &t.TrackNo, &t.DiscNo, &t.Codec, &t.BitrateKbps); err != nil {
				return provider.SearchResults{}, err
			}
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
		rows, err := p.db.QueryContext(ctx, `SELECT id,name,sort_name FROM artists WHERE lower(name) LIKE ? ORDER BY sort_name LIMIT ? OFFSET ?`, pattern, pageSize+1, offset)
		if err != nil {
			return provider.SearchResults{}, err
		}
		defer rows.Close()
		var artists []provider.Artist
		for rows.Next() {
			var a provider.Artist
			if err := rows.Scan(&a.ID, &a.Name, &a.SortName); err != nil {
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
	return provider.Lyrics{}, provider.ErrNotSupported
}

func (p *Provider) GetArtwork(ctx context.Context, ref string, sizePx int) (provider.Artwork, error) {
	return provider.Artwork{}, provider.ErrNotSupported
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
	// Try ffprobe first (most accurate)
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", path)
	out, err := cmd.Output()
	if err == nil {
		var result struct {
			Format struct {
				Duration string `json:"duration"`
			} `json:"format"`
		}
		if json.Unmarshal(out, &result) == nil && result.Format.Duration != "" {
			var secs float64
			fmt.Sscanf(result.Format.Duration, "%f", &secs)
			return int(secs * 1000)
		}
	}
	return 0
}
