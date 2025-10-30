package state

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFileName = "state.db"
	schema     = `
CREATE TABLE IF NOT EXISTS migrations (
	id TEXT PRIMARY KEY,
	source TEXT NOT NULL,
	target TEXT NOT NULL,
	domain TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS env_vars (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	migration_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	target_key TEXT,
	FOREIGN KEY (migration_id) REFERENCES migrations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dns_records (
	id TEXT PRIMARY KEY,
	migration_id TEXT,
	domain TEXT NOT NULL,
	record_type TEXT NOT NULL,
	record_name TEXT NOT NULL,
	record_value TEXT NOT NULL,
	ttl INTEGER DEFAULT 300,
	rollback_id TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (migration_id) REFERENCES migrations(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	migration_id TEXT,
	level TEXT NOT NULL,
	message TEXT NOT NULL,
	metadata TEXT,
	ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (migration_id) REFERENCES migrations(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_migrations_status ON migrations(status);
CREATE INDEX IF NOT EXISTS idx_env_vars_migration ON env_vars(migration_id);
CREATE INDEX IF NOT EXISTS idx_dns_records_migration ON dns_records(migration_id);
CREATE INDEX IF NOT EXISTS idx_logs_migration ON logs(migration_id);
CREATE INDEX IF NOT EXISTS idx_logs_ts ON logs(ts);
`
)

// DB wraps the SQLite database
type DB struct {
	db   *sql.DB
	path string
}

// Migration represents a migration record
type Migration struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnvVar represents an environment variable mapping
type EnvVar struct {
	ID          int    `json:"id"`
	MigrationID string `json:"migration_id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	TargetKey   string `json:"target_key,omitempty"`
}

// DnsRecord represents a DNS record
type DnsRecord struct {
	ID          string    `json:"id"`
	MigrationID *string   `json:"migration_id,omitempty"`
	Domain      string    `json:"domain"`
	RecordType  string    `json:"record_type"`
	RecordName  string    `json:"record_name"`
	RecordValue string    `json:"record_value"`
	TTL         int       `json:"ttl"`
	RollbackID  *string   `json:"rollback_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID          int       `json:"id"`
	MigrationID *string   `json:"migration_id,omitempty"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Metadata    *string   `json:"metadata,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Open opens or creates the state database
func Open(configDir string) (*DB, error) {
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		configDir = filepath.Join(home, ".deploy-tunnel")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	dbPath := filepath.Join(configDir, dbFileName)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &DB{db: db, path: dbPath}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// Path returns the database file path
func (d *DB) Path() string {
	return d.path
}

// CreateMigration creates a new migration record
func (d *DB) CreateMigration(id, source, target, domain string) error {
	_, err := d.db.Exec(`
		INSERT INTO migrations (id, source, target, domain, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, id, source, target, domain)
	return err
}

// GetMigration retrieves a migration by ID
func (d *DB) GetMigration(id string) (*Migration, error) {
	var m Migration
	err := d.db.QueryRow(`
		SELECT id, source, target, domain, status, created_at, updated_at
		FROM migrations WHERE id = ?
	`, id).Scan(&m.ID, &m.Source, &m.Target, &m.Domain, &m.Status, &m.CreatedAt, &m.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateMigrationStatus updates the status of a migration
func (d *DB) UpdateMigrationStatus(id, status string) error {
	_, err := d.db.Exec(`
		UPDATE migrations
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, id)
	return err
}

// ListMigrations lists all migrations, optionally filtered by status
func (d *DB) ListMigrations(status string) ([]Migration, error) {
	query := "SELECT id, source, target, domain, status, created_at, updated_at FROM migrations"
	var args []interface{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.ID, &m.Source, &m.Target, &m.Domain, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

// SaveEnvVar saves an environment variable mapping
func (d *DB) SaveEnvVar(migrationID, key, value, targetKey string) error {
	_, err := d.db.Exec(`
		INSERT INTO env_vars (migration_id, key, value, target_key)
		VALUES (?, ?, ?, ?)
	`, migrationID, key, value, targetKey)
	return err
}

// GetEnvVars retrieves all environment variables for a migration
func (d *DB) GetEnvVars(migrationID string) ([]EnvVar, error) {
	rows, err := d.db.Query(`
		SELECT id, migration_id, key, value, target_key
		FROM env_vars WHERE migration_id = ?
	`, migrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []EnvVar
	for rows.Next() {
		var e EnvVar
		if err := rows.Scan(&e.ID, &e.MigrationID, &e.Key, &e.Value, &e.TargetKey); err != nil {
			return nil, err
		}
		envVars = append(envVars, e)
	}

	return envVars, rows.Err()
}

// SaveDnsRecord saves a DNS record
func (d *DB) SaveDnsRecord(record *DnsRecord) error {
	_, err := d.db.Exec(`
		INSERT INTO dns_records (id, migration_id, domain, record_type, record_name, record_value, ttl, rollback_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, record.ID, record.MigrationID, record.Domain, record.RecordType, record.RecordName, record.RecordValue, record.TTL, record.RollbackID)
	return err
}

// GetDnsRecords retrieves DNS records for a migration
func (d *DB) GetDnsRecords(migrationID string) ([]DnsRecord, error) {
	rows, err := d.db.Query(`
		SELECT id, migration_id, domain, record_type, record_name, record_value, ttl, rollback_id, created_at
		FROM dns_records WHERE migration_id = ?
	`, migrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DnsRecord
	for rows.Next() {
		var r DnsRecord
		if err := rows.Scan(&r.ID, &r.MigrationID, &r.Domain, &r.RecordType, &r.RecordName, &r.RecordValue, &r.TTL, &r.RollbackID, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

// Log adds a log entry
func (d *DB) Log(migrationID *string, level, message, metadata string) error {
	_, err := d.db.Exec(`
		INSERT INTO logs (migration_id, level, message, metadata)
		VALUES (?, ?, ?, ?)
	`, migrationID, level, message, metadata)
	return err
}

// GetLogs retrieves logs for a migration
func (d *DB) GetLogs(migrationID string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := d.db.Query(`
		SELECT id, migration_id, level, message, metadata, ts
		FROM logs WHERE migration_id = ?
		ORDER BY ts DESC LIMIT ?
	`, migrationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.MigrationID, &l.Level, &l.Message, &l.Metadata, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}
