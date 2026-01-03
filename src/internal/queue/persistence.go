package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tunez/tunez/internal/provider"
	_ "modernc.org/sqlite"
)

// PersistenceStore handles queue state persistence to SQLite.
type PersistenceStore struct {
	db *sql.DB
}

// NewPersistenceStore creates a new persistence store at the given path.
// If dbPath is empty, uses the default location.
func NewPersistenceStore(dbPath string) (*PersistenceStore, error) {
	if dbPath == "" {
		var err error
		dbPath, err = defaultQueueDBPath()
		if err != nil {
			return nil, fmt.Errorf("resolve queue db path: %w", err)
		}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open queue db: %w", err)
	}

	store := &PersistenceStore{db: db}
	if err := store.ensureSchema(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func defaultQueueDBPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez", "state")
	case "windows":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "Tunez", "state")
	default:
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez", "state")
	}
	return filepath.Join(base, "queue.db"), nil
}

func (s *PersistenceStore) ensureSchema(ctx context.Context) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS queue_items (
			position INTEGER PRIMARY KEY,
			track_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			track_json TEXT NOT NULL,
			added_at INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS queue_state (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			current_index INTEGER NOT NULL DEFAULT -1,
			shuffle_enabled INTEGER NOT NULL DEFAULT 0,
			repeat_mode INTEGER NOT NULL DEFAULT 0,
			profile_id TEXT NOT NULL DEFAULT ''
		);`,
		// Ensure there's always exactly one state row
		`INSERT OR IGNORE INTO queue_state (id, current_index, shuffle_enabled, repeat_mode, profile_id)
		 VALUES (1, -1, 0, 0, '');`,
	}
	for _, stmt := range schema {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate queue schema: %w", err)
		}
	}
	return nil
}

// Save persists the queue state to SQLite.
func (s *PersistenceStore) Save(ctx context.Context, q *Queue, providerID, profileID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing items
	if _, err := tx.ExecContext(ctx, `DELETE FROM queue_items`); err != nil {
		return fmt.Errorf("clear queue items: %w", err)
	}

	// Insert current items
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO queue_items (position, track_id, provider_id, track_json, added_at)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	items := q.Items()
	for i, track := range items {
		trackJSON, err := json.Marshal(track)
		if err != nil {
			return fmt.Errorf("marshal track %s: %w", track.ID, err)
		}
		if _, err := stmt.ExecContext(ctx, i, track.ID, providerID, string(trackJSON), 0); err != nil {
			return fmt.Errorf("insert track %s: %w", track.ID, err)
		}
	}

	// Update state
	shuffleInt := 0
	if q.IsShuffled() {
		shuffleInt = 1
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE queue_state SET current_index = ?, shuffle_enabled = ?, repeat_mode = ?, profile_id = ? WHERE id = 1`,
		q.CurrentIndex(), shuffleInt, int(q.RepeatMode()), profileID)
	if err != nil {
		return fmt.Errorf("update queue state: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// LoadResult contains the result of loading a queue from persistence.
type LoadResult struct {
	Tracks       []provider.Track
	CurrentIndex int
	Shuffled     bool
	Repeat       RepeatMode
	ProfileID    string
}

// Load reads the queue state from SQLite.
func (s *PersistenceStore) Load(ctx context.Context) (LoadResult, error) {
	result := LoadResult{CurrentIndex: -1}

	// Load state
	var shuffleInt int
	err := s.db.QueryRowContext(ctx,
		`SELECT current_index, shuffle_enabled, repeat_mode, profile_id FROM queue_state WHERE id = 1`).
		Scan(&result.CurrentIndex, &shuffleInt, &result.Repeat, &result.ProfileID)
	if err != nil && err != sql.ErrNoRows {
		return result, fmt.Errorf("load queue state: %w", err)
	}
	result.Shuffled = shuffleInt == 1

	// Load items
	rows, err := s.db.QueryContext(ctx,
		`SELECT track_json FROM queue_items ORDER BY position ASC`)
	if err != nil {
		return result, fmt.Errorf("load queue items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trackJSON string
		if err := rows.Scan(&trackJSON); err != nil {
			return result, fmt.Errorf("scan track: %w", err)
		}

		var track provider.Track
		if err := json.Unmarshal([]byte(trackJSON), &track); err != nil {
			// Skip corrupted entries
			continue
		}
		result.Tracks = append(result.Tracks, track)
	}

	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate tracks: %w", err)
	}

	// Validate current index
	if result.CurrentIndex >= len(result.Tracks) {
		result.CurrentIndex = len(result.Tracks) - 1
	}
	if result.CurrentIndex < 0 && len(result.Tracks) > 0 {
		result.CurrentIndex = 0
	}

	return result, nil
}

// Clear removes all persisted queue data.
func (s *PersistenceStore) Clear(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM queue_items`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE queue_state SET current_index = -1, shuffle_enabled = 0, repeat_mode = 0 WHERE id = 1`); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes the database connection.
func (s *PersistenceStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
