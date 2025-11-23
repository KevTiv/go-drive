package database

import (
	"fmt"
	"log"

	"go-drive/internal/domain"

	"gorm.io/gorm"
)

// AutoMigrate runs automatic migrations for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Enable pgcrypto extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return fmt.Errorf("failed to create pgcrypto extension: %w", err)
	}

	// AutoMigrate all models
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Folder{},
		&domain.File{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	// Add custom indexes and constraints that GORM doesn't handle automatically
	if err := addCustomConstraints(db); err != nil {
		return fmt.Errorf("failed to add custom constraints: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// addCustomConstraints adds custom database constraints
func addCustomConstraints(db *gorm.DB) error {
	// Add email format constraint for users
	if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'users_email_format_check'
			) THEN
				ALTER TABLE users
				ADD CONSTRAINT users_email_format_check
				CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
			END IF;
		END $$;
	`).Error; err != nil {
		return fmt.Errorf("failed to add email format constraint: %w", err)
	}

	// Add no self-parent constraint for folders
	if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'folders_no_self_parent_check'
			) THEN
				ALTER TABLE folders
				ADD CONSTRAINT folders_no_self_parent_check
				CHECK (id != parent_id);
			END IF;
		END $$;
	`).Error; err != nil {
		return fmt.Errorf("failed to add no self-parent constraint: %w", err)
	}

	// Create or replace function to update updated_at timestamp
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`).Error; err != nil {
		return fmt.Errorf("failed to create update_updated_at_column function: %w", err)
	}

	// Create triggers for auto-updating updated_at
	tables := []string{"users", "files", "folders"}
	for _, table := range tables {
		triggerName := fmt.Sprintf("update_%s_updated_at", table)
		if err := db.Exec(fmt.Sprintf(`
			DROP TRIGGER IF EXISTS %s ON %s;
			CREATE TRIGGER %s
			BEFORE UPDATE ON %s
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`, triggerName, table, triggerName, table)).Error; err != nil {
			return fmt.Errorf("failed to create trigger for %s: %w", table, err)
		}
	}

	return nil
}

// CreateServiceRoles creates database roles for microservices
func CreateServiceRoles(db *gorm.DB) error {
	log.Println("Creating service roles...")

	roles := []struct {
		name     string
		password string
	}{
		{"user_service", "user_service_password"},
		{"file_service", "file_service_password"},
		{"analytics_reader", "analytics_password"},
	}

	for _, role := range roles {
		if err := db.Exec(fmt.Sprintf(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN
					CREATE ROLE %s WITH LOGIN PASSWORD '%s';
				END IF;
			END
			$$;
		`, role.name, role.name, role.password)).Error; err != nil {
			return fmt.Errorf("failed to create role %s: %w", role.name, err)
		}
		log.Printf("Role %s created/verified", role.name)
	}

	// Grant permissions
	if err := grantPermissions(db); err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}

	log.Println("Service roles created successfully")
	return nil
}

// grantPermissions grants appropriate permissions to service roles
func grantPermissions(db *gorm.DB) error {
	// Grant permissions to user_service
	if err := db.Exec(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON users TO user_service;
		GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO user_service;
	`).Error; err != nil {
		return fmt.Errorf("failed to grant permissions to user_service: %w", err)
	}

	// Grant permissions to file_service
	if err := db.Exec(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON files, folders TO file_service;
		GRANT SELECT ON users TO file_service;
		GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO file_service;
	`).Error; err != nil {
		return fmt.Errorf("failed to grant permissions to file_service: %w", err)
	}

	// Grant read-only permissions to analytics_reader
	if err := db.Exec(`
		GRANT SELECT ON users, files, folders TO analytics_reader;
	`).Error; err != nil {
		return fmt.Errorf("failed to grant permissions to analytics_reader: %w", err)
	}

	return nil
}

// DropAllTables drops all tables (use with caution!)
func DropAllTables(db *gorm.DB) error {
	log.Println("WARNING: Dropping all tables...")

	if err := db.Migrator().DropTable(
		&domain.File{},
		&domain.Folder{},
		&domain.User{},
	); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	log.Println("All tables dropped")
	return nil
}
