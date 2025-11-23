package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Users struct {
	ID        string  `json:"id" gorm:"default:uuid_generate_v4()"`
	FirstName string  `json:"first_name" gorm:"type:varchar(255);not null"`
	Surname   string  `json:"surname" gorm:"type:varchar(255);not null"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Country   *string `json:"country" gorm:"default:'The netherlands'"`
	Region    *string `json:"region"`
	City      *string `json:"city"`
	Type      string  `json:"type"`

	EmailVerified bool `gorm:"default:false"`
	IsActive      bool `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (u *Users) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *Users) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
