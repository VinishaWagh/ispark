package models

import (
	"time"

	"gorm.io/gorm"
)

// Announcement is a platform-wide notice managed by a super administrator.
// Dates are stored as date-only UTC values and formatted by the API so clients
// do not shift the displayed day when converting time zones.
type Announcement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(200);not null;index" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"type:varchar(30);not null;index" json:"category"`
	Audience    string    `gorm:"type:varchar(30);not null;index" json:"audience"`
	Priority    string    `gorm:"type:varchar(20);not null;default:'Medium'" json:"priority"`
	PublishDate time.Time `gorm:"type:date;not null;index" json:"publish_date"`
	ExpiryDate  time.Time `gorm:"type:date;not null;index" json:"expiry_date"`
	Status      string    `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	// NotifiedAt records when this announcement's notifications were fanned out
	// to students. It is set exactly once, the first time the announcement goes
	// live, so re-publishing an already-published announcement never delivers a
	// second copy of the same message.
	NotifiedAt *time.Time     `gorm:"index" json:"-"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
