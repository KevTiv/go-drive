package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id"`
	FirstName     string     `json:"first_name"`
	Surname       string     `json:"surname"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone,omitempty"`
	Country       string     `json:"country"`
	Region        string     `json:"region,omitempty"`
	City          string     `json:"city,omitempty"`
	Type          string     `json:"type"`
	EmailVerified bool       `json:"email_verified"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
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
