package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// RLSManager manages Row Level Security policies
type RLSManager struct {
	db *sql.DB
}

// NewRLSManager creates a new RLS manager
func NewRLSManager(db *sql.DB) *RLSManager {
	return &RLSManager{db: db}
}

// EnableRLS enables Row Level Security on a table
func (m *RLSManager) EnableRLS(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", tableName)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable RLS on %s: %w", tableName, err)
	}
	return nil
}

// DisableRLS disables Row Level Security on a table
func (m *RLSManager) DisableRLS(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", tableName)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to disable RLS on %s: %w", tableName, err)
	}
	return nil
}

// PolicyConfig represents an RLS policy configuration
type PolicyConfig struct {
	Name      string
	Table     string
	Command   string // ALL, SELECT, INSERT, UPDATE, DELETE
	Role      string
	Using     string // USING clause
	WithCheck string // WITH CHECK clause (optional)
}

// CreatePolicy creates a new RLS policy
func (m *RLSManager) CreatePolicy(ctx context.Context, cfg PolicyConfig) error {
	query := fmt.Sprintf(
		"CREATE POLICY %s ON %s FOR %s TO %s USING (%s)",
		cfg.Name, cfg.Table, cfg.Command, cfg.Role, cfg.Using,
	)

	if cfg.WithCheck != "" {
		query += fmt.Sprintf(" WITH CHECK (%s)", cfg.WithCheck)
	}

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create policy %s: %w", cfg.Name, err)
	}
	return nil
}

// DropPolicy drops an RLS policy
func (m *RLSManager) DropPolicy(ctx context.Context, policyName, tableName string) error {
	query := fmt.Sprintf("DROP POLICY IF EXISTS %s ON %s", policyName, tableName)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop policy %s: %w", policyName, err)
	}
	return nil
}

// CreateServiceRole creates a database role for a microservice
func (m *RLSManager) CreateServiceRole(ctx context.Context, roleName, password string) error {
	// Check if role exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT FROM pg_roles WHERE rolname = $1)"
	if err := m.db.QueryRowContext(ctx, checkQuery, roleName).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if exists {
		return nil // Role already exists
	}

	// Create role
	query := fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'", roleName, password)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create role %s: %w", roleName, err)
	}

	return nil
}

// GrantTablePermissions grants permissions on a table to a role
func (m *RLSManager) GrantTablePermissions(ctx context.Context, tableName, roleName string, permissions []string) error {
	for _, perm := range permissions {
		query := fmt.Sprintf("GRANT %s ON %s TO %s", perm, tableName, roleName)
		_, err := m.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to grant %s on %s to %s: %w", perm, tableName, roleName, err)
		}
	}
	return nil
}

// RevokeTablePermissions revokes permissions on a table from a role
func (m *RLSManager) RevokeTablePermissions(ctx context.Context, tableName, roleName string, permissions []string) error {
	for _, perm := range permissions {
		query := fmt.Sprintf("REVOKE %s ON %s FROM %s", perm, tableName, roleName)
		_, err := m.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to revoke %s on %s from %s: %w", perm, tableName, roleName, err)
		}
	}
	return nil
}

// SetupServiceRLS sets up RLS for a microservice
// This is a helper function to configure RLS for common service patterns
func SetupServiceRLS(ctx context.Context, db *sql.DB, serviceName, tableName string, fullAccess bool) error {
	mgr := NewRLSManager(db)

	// Enable RLS on table
	if err := mgr.EnableRLS(ctx, tableName); err != nil {
		return err
	}

	// Create policy based on access level
	var policy PolicyConfig
	if fullAccess {
		// Full CRUD access
		policy = PolicyConfig{
			Name:    fmt.Sprintf("%s_all_%s", serviceName, tableName),
			Table:   tableName,
			Command: "ALL",
			Role:    serviceName,
			Using:   "true",
		}
	} else {
		// Read-only access
		policy = PolicyConfig{
			Name:    fmt.Sprintf("%s_select_%s", serviceName, tableName),
			Table:   tableName,
			Command: "SELECT",
			Role:    serviceName,
			Using:   "true",
		}
	}

	return mgr.CreatePolicy(ctx, policy)
}
