# Tunez — Filesystem Provider (Built-in)

**Last updated:** 2026-01-02  
**Provider ID:** `filesystem`

## 1. Overview
The Filesystem Provider spans local directories to build a rich, searchable music library. It is designed to handle large collections (100k+ tracks) with sub-second query response times and fast incremental startup scans.

## 2. Architecture & Storage

We use **SQLite** as the backing store for the index.
- **Why:** Zero-conf, highly optimized for read-heavy workloads, supports Full Text Search (FTS5), and handles reliable concurrency.
- **Driver:** `modernc.org/sqlite` (Pure Go, CGO-free) is recommended for easier cross-compilation.

### 2.1 Schema Design
The schema is normalized to support fast browsing (Artists -> Albums -> Tracks) and efficient search.

```sql
-- Metadata for the scanner state (e.g. last scan time, schema version)
CREATE TABLE meta (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE artists (
    id TEXT PRIMARY KEY,       -- normalized hash (e.g. SHA256 of lowercase name)
    name TEXT NOT NULL,
    sort_name TEXT NOT NULL    -- normalized for sorting (removed "The ", etc.)
);

CREATE TABLE albums (
    id TEXT PRIMARY KEY,       -- hash(artist_id + album_title)
    artist_id TEXT NOT NULL,
    title TEXT NOT NULL,
    year INTEGER,
    artwork_path TEXT,
    FOREIGN KEY(artist_id) REFERENCES artists(id)
);

CREATE TABLE tracks (
    id TEXT PRIMARY KEY,       -- hash(file_path) needed for stable ID across re-tags
    album_id TEXT NOT NULL,
    artist_id TEXT NOT NULL,
    
    title TEXT NOT NULL,
    album_title TEXT NOT NULL, -- Cached for faster denormalized searches
    artist_name TEXT NOT NULL, -- Cached for faster denormalized searches
    
    track_number INTEGER,
    disc_number INTEGER,
    duration_ms INTEGER,
    
    file_path TEXT NOT NULL UNIQUE,
    file_size INTEGER,         -- extraction caching
    file_mtime INTEGER,        -- extraction caching
    
    codec TEXT,
    bitrate INTEGER,
    
    FOREIGN KEY(album_id) REFERENCES albums(id),
    FOREIGN KEY(artist_id) REFERENCES artists(id)
);

-- Virtual table for lightning-fast global search
CREATE VIRTUAL TABLE search_index USING fts5(
    title, 
    artist, 
    album, 
    content='tracks', 
    content_rowid='rowid'
);

-- Triggers to keep FTS index in sync with tracks table would be added here.
```

### 2.2 Indices for Performance
Indices MUST be created to support specific UI access patterns:
- `CREATE INDEX idx_tracks_album_disc_seq ON tracks (album_id, disc_number, track_number);` (Album view)
- `CREATE INDEX idx_albums_artist_year ON albums (artist_id, year, title);` (Artist view)
- `CREATE INDEX idx_artists_sort ON artists (sort_name);` (Library browsing)

## 3. High-Performance Scanning Strategy

Scanning large libraries is I/O intensive. Naive approaches (serial walk + single insert) are too slow.

### 3.1 Two-Phase Scan (Walk & Diff)
1.  **Layout Phase (Fast Walk)**:
    - Recursively walk user directories (`godirwalk` or `filepath.WalkDir`).
    - Collect `(path, size, mtime)` tuples.
    - Compare against DB `tracks` table.
    - **Result**: specific lists of `ToInsert`, `ToUpdate`, `ToDelete`.
    - *Goal*: Avoid opening/parsing files designed that haven't changed.

2.  **Processing Phase (Concurrent)**:
    - For `ToInsert` / `ToUpdate`:
        - Spin up a **Worker Pool** (e.g., `runtime.NumCPU() * 2` workers).
        - Each worker reads the file tags (use `github.com/dhowden/tag` or similar).
        - Worker sends parsed `Track` struct to a **Batch Writer** channel.

### 3.2 Batch Writing (ACID)
- **Problem**: 10,000 separate `INSERT` statements are slow due to fsync overhead.
- **Solution**: The Batch Writer collects tracks into chunks (e.g., 500 items) and writes them inside a **Single Transaction**.
- **FTS**: Defer FTS index updates until the end of the batch or scan for maximum throughput.

### 3.3 Incremental Updates
- On startup, run the Two-Phase scan.
- Because `mtime` is checked first, startup on an unchanged library of 100k files should take milliseconds to seconds, not minutes.
- Only parsing changed/new files keeps the UI responsive.

## 4. Playback Implementation
- **Stream URL**: Returns `file://<absolute_path>`.
- **Latency**: Zero. The scanner ensures the path existed at scan time. If `mpv` fails to load (file deleted externally), the Provider returns a specific error, and the core removes it from the queue.

## 5. Configuration & Limits
- **Extensions**: Allowlist (mp3, flac, m4a, ogg, wav, opus).
- **Hidden Files**: Ignored by default.
- **Symlinks**: Followed (configurable), with loop detection.

## 6. Recommended Libraries
- **DB**: `modernc.org/sqlite` (Embedded, C-free)
- **Tag Parsing**: `github.com/dhowden/tag` (Robust, widely used)
- **File Walking**: `github.com/karrick/godirwalk` (Faster than `path/filepath` on some systems) or standard library `filepath.WalkDir` (Go 1.16+ is quite optimized).
