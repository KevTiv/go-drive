-- Migration: Add additional indexes for user queries
-- Version: 002_add_user_indexes
-- Description: Add performance indexes for common user query patterns

-- Add index for email verification status
CREATE INDEX IF NOT EXISTS idx_users_email_verified
ON users(email_verified)
WHERE deleted_at IS NULL;

-- Add composite index for type and active status
CREATE INDEX IF NOT EXISTS idx_users_type_active
ON users(type, is_active)
WHERE deleted_at IS NULL;

-- Add index for phone lookups
CREATE INDEX IF NOT EXISTS idx_users_phone
ON users(phone)
WHERE deleted_at IS NULL AND phone IS NOT NULL;

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('002_add_user_indexes', 'Add additional indexes for user queries')
ON CONFLICT (version) DO NOTHING;
