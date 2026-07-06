package models

import (
	"time"

	"gorm.io/gorm"
)

type Enrollment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	StudentRollNo string         `gorm:"type:varchar(50);not null" json:"student_roll_no"`
	Student       Student        `gorm:"foreignKey:StudentRollNo;references:RollNo" json:"student"`
	ActivityID    uint           `gorm:"not null" json:"activity_id"`
	Activity      Activity       `gorm:"foreignKey:ActivityID;references:ID" json:"activity"`
	Status        string         `gorm:"type:varchar(50);default:'Enrolled'" json:"status"` // Enrolled, Completed, Cancelled
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
