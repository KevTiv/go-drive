package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName     string         `json:"first_name" gorm:"column:firstname;type:varchar(100);not null"`
	Surname       string         `json:"surname" gorm:"type:varchar(100);not null"`
	Email         string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone         string         `json:"phone,omitempty" gorm:"type:varchar(50)"`
	Country       string         `json:"country" gorm:"type:varchar(100);default:'The netherlands'"`
	Region        string         `json:"region,omitempty" gorm:"type:varchar(100)"`
	City          string         `json:"city,omitempty" gorm:"type:varchar(100)"`
	Type          string         `json:"type" gorm:"type:varchar(50);default:'standard';index"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	IsActive      bool           `json:"is_active" gorm:"default:true;index"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index:idx_users_created_at,sort:desc"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// UserType constants
const (
	UserTypeStandard = "standard"
	UserTypePremium  = "premium"
	UserTypeAdmin    = "admin"
)

// IsValidUserType checks if a user type is valid
func IsValidUserType(userType string) bool {
	switch userType {
	case UserTypeStandard, UserTypePremium, UserTypeAdmin:
		return true
	default:
		return false
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	Surname   string `json:"surname" validate:"required,min=1,max=100"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Country   string `json:"country,omitempty" validate:"omitempty,max=100"`
	Region    string `json:"region,omitempty" validate:"omitempty,max=100"`
	City      string `json:"city,omitempty" validate:"omitempty,max=100"`
	Type      string `json:"type,omitempty" validate:"omitempty,oneof=standard premium admin"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	FirstName *string   `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	Surname   *string   `json:"surname,omitempty" validate:"omitempty,min=1,max=100"`
	Email     *string   `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone     *string   `json:"phone,omitempty" validate:"omitempty,max=50"`
	Country   *string   `json:"country,omitempty" validate:"omitempty,max=100"`
	Region    *string   `json:"region,omitempty" validate:"omitempty,max=100"`
	City      *string   `json:"city,omitempty" validate:"omitempty,max=100"`
	Type      *string   `json:"type,omitempty" validate:"omitempty,oneof=standard premium admin"`
}
