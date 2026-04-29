package blackboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteBlackboard implements the Blackboard interface using a persistent SQLite backend.
type SQLiteBlackboard struct {
	db    *sql.DB
	mutex sync.RWMutex

	// Pub/Sub fields
	subscribers []chan any
	subMutex    sync.Mutex
}

// NewSQLiteBlackboard initializes a new SQLite database with performance and integrity optimizations.
func NewSQLiteBlackboard(dbPath string) (*SQLiteBlackboard, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// ── Optimization: WAL mode for concurrency, Foreign Keys for integrity
	initialPRAGMAs := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, pragma := range initialPRAGMAs {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	bb := &SQLiteBlackboard{db: db}
	if err := bb.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return bb, nil
}

func (bb *SQLiteBlackboard) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		workspace_path TEXT,
		created_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS turns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		turn_index INTEGER,
		user_prompt TEXT,
		system_response TEXT,
		created_at TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		turn_id INTEGER,
		role TEXT,
		status INTEGER,
		input_data TEXT,
		output_data TEXT,
		created_at TIMESTAMP,
		FOREIGN KEY(turn_id) REFERENCES turns(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS artifacts (
		id TEXT PRIMARY KEY,
		task_id TEXT,
		file_path TEXT,
		content TEXT,
		confidence_score REAL,
		version INTEGER,
		created_at TIMESTAMP,
		FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS kv_store (
		key TEXT PRIMARY KEY,
		value_json TEXT,
		updated_at TIMESTAMP
	);
	`
	_, err := bb.db.Exec(schema)
	return err
}

// ── Pub/Sub Implementation ──────────────────────────────────────────────────

func (bb *SQLiteBlackboard) Subscribe(ctx context.Context) <-chan any {
	ch := make(chan any, 10)
	bb.subMutex.Lock()
	bb.subscribers = append(bb.subscribers, ch)
	bb.subMutex.Unlock()

	go func() {
		<-ctx.Done()
		bb.subMutex.Lock()
		for i, sub := range bb.subscribers {
			if sub == ch {
				bb.subscribers = append(bb.subscribers[:i], bb.subscribers[i+1:]...)
				close(ch)
				break
			}
		}
		bb.subMutex.Unlock()
	}()

	return ch
}

func (bb *SQLiteBlackboard) notify(event any) {
	bb.subMutex.Lock()
	defer bb.subMutex.Unlock()
	for _, sub := range bb.subscribers {
		select {
		case sub <- event:
		default:
			// Buffer full, skip to prevent blocking
		}
	}
}

// ── Session & Turn Management ────────────────────────────────────────────────

func (bb *SQLiteBlackboard) NewSession(id, path string) error {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	_, err := bb.db.Exec(`INSERT INTO sessions (id, workspace_path, created_at) VALUES (?, ?, ?)`, id, path, time.Now())
	if err == nil {
		bb.notify(fmt.Sprintf("New session created: %s", id))
	}
	return err
}

func (bb *SQLiteBlackboard) GetLatestSession(path string) (*Session, error) {
	bb.mutex.RLock()
	defer bb.mutex.RUnlock()
	var s Session
	query := `SELECT id, workspace_path, created_at FROM sessions WHERE workspace_path = ? ORDER BY created_at DESC LIMIT 1`
	err := bb.db.QueryRow(query, path).Scan(&s.ID, &s.WorkspacePath, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (bb *SQLiteBlackboard) CreateTurn(sessionID, prompt string, index int) (int64, error) {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	query := `INSERT INTO turns (session_id, user_prompt, turn_index, created_at) VALUES (?, ?, ?, ?)`
	res, err := bb.db.Exec(query, sessionID, prompt, index, time.Now())
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	bb.notify(fmt.Sprintf("Turn %d started for session %s", index, sessionID))
	return id, nil
}

// ── Task & Artifact Management ──────────────────────────────────────────────

func (bb *SQLiteBlackboard) UpsertTask(t *Task) error {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	query := `INSERT INTO tasks (id, turn_id, role, status, input_data, output_data, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)
	          ON CONFLICT(id) DO UPDATE SET status=excluded.status, output_data=excluded.output_data`
	_, err := bb.db.Exec(query, t.ID, t.TurnID, t.Role, t.Status, t.InputData, t.OutputData, time.Now())
	if err == nil {
		bb.notify(t)
	}
	return err
}

func (bb *SQLiteBlackboard) SaveArtifact(a *Artifact) error {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	query := `INSERT INTO artifacts (id, task_id, file_path, content, confidence_score, version, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := bb.db.Exec(query, a.ID, a.TaskID, a.FilePath, a.Content, a.ConfidenceScore, a.Version, time.Now())
	if err == nil {
		bb.notify(a)
	}
	return err
}

func (bb *SQLiteBlackboard) GetArtifactsForTurn(turnID int64) ([]*Artifact, error) {
	bb.mutex.RLock()
	defer bb.mutex.RUnlock()
	query := `SELECT a.id, a.task_id, a.file_path, a.content, a.confidence_score, a.version, a.created_at 
	          FROM artifacts a JOIN tasks t ON a.task_id = t.id WHERE t.turn_id = ?`
	rows, err := bb.db.Query(query, turnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*Artifact
	for rows.Next() {
		var a Artifact
		if err := rows.Scan(&a.ID, &a.TaskID, &a.FilePath, &a.Content, &a.ConfidenceScore, &a.Version, &a.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &a)
	}
	return res, nil
}

// ── KV Persistence ──────────────────────────────────────────────────────────

func (bb *SQLiteBlackboard) Set(key string, value any) {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	jsonData, _ := json.Marshal(value)
	query := `INSERT INTO kv_store (key, value_json, updated_at) VALUES (?, ?, ?)
	          ON CONFLICT(key) DO UPDATE SET value_json=excluded.value_json, updated_at=excluded.updated_at`
	_, _ = bb.db.Exec(query, key, string(jsonData), time.Now())
	bb.notify(key)
}

func (bb *SQLiteBlackboard) Get(key string) (any, bool) {
	bb.mutex.RLock()
	defer bb.mutex.RUnlock()
	var valJSON string
	err := bb.db.QueryRow("SELECT value_json FROM kv_store WHERE key = ?", key).Scan(&valJSON)
	if err != nil {
		return nil, false
	}
	var val any
	_ = json.Unmarshal([]byte(valJSON), &val)
	return val, true
}

func (bb *SQLiteBlackboard) Delete(key string) {
	bb.mutex.Lock()
	defer bb.mutex.Unlock()
	_, _ = bb.db.Exec("DELETE FROM kv_store WHERE key = ?", key)
}

func (bb *SQLiteBlackboard) Keys() []string {
	bb.mutex.RLock()
	defer bb.mutex.RUnlock()
	rows, err := bb.db.Query("SELECT key FROM kv_store")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		rows.Scan(&k)
		keys = append(keys, k)
	}
	return keys
}

func (bb *SQLiteBlackboard) Close() error {
	return bb.db.Close()
}
