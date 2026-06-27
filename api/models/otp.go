package models

import (
	"time"
)

type OTP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(100);index;not null" json:"email"`
	Code      string    `gorm:"type:varchar(10);not null" json:"code"`
	Purpose   string    `gorm:"type:varchar(50);not null" json:"purpose"` // "register" or "forgot_password"
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
