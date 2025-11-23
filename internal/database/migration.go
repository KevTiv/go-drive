package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	SQL         string
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version     string
	AppliedAt   time.Time
	Description string
}

// Migrator handles database migrations
type Migrator struct {
	conn *Connection
}

// NewMigrator creates a new migrator
func NewMigrator(conn *Connection) *Migrator {
	return &Migrator{conn: conn}
}

// Apply applies pending migrations
func (m *Migrator) Apply(ctx context.Context, migrations []Migration) error {
	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Create a map of applied versions
	appliedMap := make(map[string]bool)
	for _, record := range applied {
		appliedMap[record.Version] = true
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Apply pending migrations
	for _, migration := range migrations {
		if appliedMap[migration.Version] {
			continue
		}

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		fmt.Printf("Applied migration: %s - %s\n", migration.Version, migration.Description)
	}

	return nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	return m.conn.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record migration
		query := `
			INSERT INTO schema_migrations (version, description, applied_at)
			VALUES ($1, $2, $3)
		`
		_, err := tx.ExecContext(ctx, query, migration.Version, migration.Description, time.Now())
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	})
}

// GetAppliedMigrations returns all applied migrations
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error) {
	query := `
		SELECT version, applied_at, description
		FROM schema_migrations
		ORDER BY version
	`

	rows, err := m.conn.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.AppliedAt, &record.Description); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)
	`

	_, err := m.conn.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}
