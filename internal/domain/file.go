package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File represents a file in the system
type File struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	FolderID   *uuid.UUID     `json:"folder_id,omitempty" gorm:"type:uuid;index"`
	Folder     *Folder        `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	Size       int64          `json:"size" gorm:"not null;check:size >= 0"`
	MimeType   string         `json:"mime_type" gorm:"type:varchar(100)"`
	StorageKey string         `json:"storage_key" gorm:"type:varchar(500);not null"`
	Checksum   string         `json:"checksum,omitempty" gorm:"type:varchar(64)"`
	CreatedAt  time.Time      `json:"created_at" gorm:"index:idx_files_created_at,sort:desc"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for the File model
func (File) TableName() string {
	return "files"
}

// Folder represents a folder in the system
type Folder struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	User      *User          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ParentID  *uuid.UUID     `json:"parent_id,omitempty" gorm:"type:uuid;index"`
	Parent    *Folder        `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for the Folder model
func (Folder) TableName() string {
	return "folders"
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
