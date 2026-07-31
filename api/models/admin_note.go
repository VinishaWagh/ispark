package models

import (
	"gorm.io/gorm"
	"time"
)

// AdminNoteReminderPrefix marks an admin note as an activity-monitoring
// reminder. It is an internal marker for the staff note trail and is stripped
// before the text is shown to a student, so it lives here rather than being
// spelled out at each site that writes or recognises one.
const AdminNoteReminderPrefix = "[ALERT REMINDER] "

type AdminNote struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	StudentRollNo string         `gorm:"type:varchar(50);not null" json:"student_roll_no"`
	AdminID       string         `gorm:"type:varchar(50);not null" json:"admin_id"`
	AuthorName    string         `gorm:"type:varchar(100)" json:"author"`
	Role          string         `gorm:"type:varchar(100)" json:"role"`
	Text          string         `gorm:"type:text;not null" json:"text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ObservationInput struct {
	Text string `json:"text"`
}

type NoticeInput struct {
	Message string `json:"message"`
}
