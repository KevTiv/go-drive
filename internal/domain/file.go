package domain

import (
	"time"

	"github.com/google/uuid"
)

// File represents a file in the system
type File struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	UserID     uuid.UUID  `json:"user_id"`
	FolderID   *uuid.UUID `json:"folder_id,omitempty"`
	Size       int64      `json:"size"`
	MimeType   string     `json:"mime_type"`
	StorageKey string     `json:"storage_key"`
	Checksum   string     `json:"checksum,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// Folder represents a folder in the system
type Folder struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	UserID    uuid.UUID  `json:"user_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CreateFileRequest represents a request to create a file
type CreateFileRequest struct {
	Name       string     `json:"name" validate:"required,min=1,max=255"`
	UserID     uuid.UUID  `json:"user_id" validate:"required"`
	FolderID   *uuid.UUID `json:"folder_id,omitempty"`
	Size       int64      `json:"size" validate:"required,min=0"`
	MimeType   string     `json:"mime_type,omitempty" validate:"omitempty,max=100"`
	StorageKey string     `json:"storage_key" validate:"required,max=500"`
	Checksum   string     `json:"checksum,omitempty" validate:"omitempty,max=64"`
}

// CreateFolderRequest represents a request to create a folder
type CreateFolderRequest struct {
	Name     string     `json:"name" validate:"required,min=1,max=255"`
	UserID   uuid.UUID  `json:"user_id" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}
