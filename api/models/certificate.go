package models

import (
	"time"

	"gorm.io/gorm"
)

type Certificate struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	StudentRollNo     string         `gorm:"type:varchar(50);not null" json:"student_roll_no"`
	Student           Student        `gorm:"foreignKey:StudentRollNo;references:RollNo" json:"student"`
	ActivityName      string         `gorm:"type:varchar(255);not null" json:"activity_name"`
	ActivityCategory  string         `gorm:"type:varchar(100);not null" json:"activity_category"`
	ActivityDate      time.Time      `json:"activity_date"`
	OrganizerName     string         `gorm:"type:varchar(255)" json:"organizer_name"`
	EventLevel        string         `gorm:"type:varchar(100)" json:"event_level"`
	CertNumber        string         `gorm:"type:varchar(100);not null" json:"cert_number"`
	IssueDate         time.Time      `json:"issue_date"`
	ParticipationType string         `gorm:"type:varchar(100)" json:"participation_type"` // Winner, Runner Up, Participant, Coordinator, Volunteer
	Description       string         `gorm:"type:text" json:"description"`
	FileName          string         `gorm:"type:varchar(255)" json:"file_name"`
	FilePath          string         `gorm:"type:varchar(255)" json:"file_path"`
	Credits           int            `gorm:"default:0" json:"credits"`
	Status            string         `gorm:"type:varchar(50);default:'Pending'" json:"status"` // Pending, Approved, Rejected
	RejectionReason   string         `gorm:"type:text" json:"rejection_reason"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
