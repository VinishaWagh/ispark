package models

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Category     string         `gorm:"type:varchar(100);not null" json:"category"` // TECHNICAL, SPORTS, CULTURAL, SOCIAL SERVICE, RESEARCH, PUBLIC SPEAKING, LEADERSHIP, etc.
	Description  string         `gorm:"type:text" json:"description"`
	Credits      int            `gorm:"not null" json:"credits"`
	Mode         string         `gorm:"type:varchar(50);not null" json:"mode"` // Online, Offline, Hybrid
	RegDeadline  time.Time      `json:"reg_deadline"`
	ActivityDate time.Time      `json:"activity_date"`
	Venue        string         `gorm:"type:varchar(255)" json:"venue"`
	Coordinator  string         `gorm:"type:varchar(100)" json:"coordinator"`
	Status       string         `gorm:"type:varchar(50);default:'Open'" json:"status"` // Open, Closed, Closing Soon
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
