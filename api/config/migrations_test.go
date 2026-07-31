package config

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/iips-oss/ispark/api/models"
	"gorm.io/gorm"
)

func TestNormalizeStudentCourses(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Student{}); err != nil {
		t.Fatalf("migrate students: %v", err)
	}

	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })

	tests := []struct {
		legacy    string
		canonical string
	}{
		{"M.Tech (Computer Science - CS)", models.CourseMTechCS},
		{"M.Tech (Information Technology - IT)", models.CourseMTechIT},
		{"MCA (Master of Computer Applications)", models.CourseMCA5Yr},
		{"MBA (Management Science - MS)", models.CourseMBAMS5Yr},
		{"MBA (Management Science)", models.CourseMBAMS2Yr},
		{"MBA (Advertising and Public Relations - APR)", models.CourseMBAAPR},
		{"MBA (Entrepreneurship)", models.CourseMBAEnt},
		{"B.Com. (Hons.)", models.CourseBComHons},
		{"  mca  ", models.CourseMCA5Yr},
	}

	for i, test := range tests {
		student := models.Student{
			RollNo:       fmt.Sprintf("R%02d", i),
			Name:         "Migration Test",
			CourseName:   test.legacy,
			Semester:     1,
			EmailID:      fmt.Sprintf("student-%d@example.com", i),
			EnrollmentNo: fmt.Sprintf("E%02d", i),
			Password:     "test",
		}
		if err := db.Create(&student).Error; err != nil {
			t.Fatalf("seed %q: %v", test.legacy, err)
		}
	}

	if err := normalizeStudentCourses(); err != nil {
		t.Fatalf("normalize courses: %v", err)
	}
	if err := normalizeStudentCourses(); err != nil {
		t.Fatalf("normalize courses a second time: %v", err)
	}

	for i, test := range tests {
		var student models.Student
		if err := db.First(&student, "roll_no = ?", fmt.Sprintf("R%02d", i)).Error; err != nil {
			t.Fatalf("load %q: %v", test.legacy, err)
		}
		if student.CourseName != test.canonical {
			t.Errorf("%q: expected %q, got %q", test.legacy, test.canonical, student.CourseName)
		}
	}
}

// Reminders sent before the student endpoint moved to the notifications table
// exist only as admin notes. The backfill has to lift them across so they stay
// visible in the bell, exactly once however many times the server restarts.
func TestBackfillReminderNotifications(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Student{}, &models.AdminNote{}, &models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })

	student := models.Student{
		RollNo:       "IT2K24001",
		Name:         "Aarav Sharma",
		CourseName:   models.CourseMCA5Yr,
		Semester:     4,
		ContactNo:    "9876543211",
		EmailID:      "aarav@example.com",
		EnrollmentNo: "EN-IT2K24001",
	}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("seed student: %v", err)
	}

	sentAt := time.Now().AddDate(0, 0, -3).Truncate(time.Second)
	reminder := models.AdminNote{
		StudentRollNo: student.RollNo,
		AdminID:       "ADM24",
		AuthorName:    "Mentor 24",
		Role:          "admin",
		Text:          models.AdminNoteReminderPrefix + "Reminder for Python Workshop: Pending Verification.",
		CreatedAt:     sentAt,
	}
	if err := db.Create(&reminder).Error; err != nil {
		t.Fatalf("seed reminder note: %v", err)
	}

	// An ordinary observation is not a reminder and must not be lifted across.
	observation := models.AdminNote{
		StudentRollNo: student.RollNo,
		AdminID:       "ADM24",
		AuthorName:    "Mentor 24",
		Role:          "admin",
		Text:          "Met the student about their project.",
	}
	if err := db.Create(&observation).Error; err != nil {
		t.Fatalf("seed observation note: %v", err)
	}

	if err := backfillReminderNotifications(); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	var notifications []models.Notification
	if err := db.Find(&notifications).Error; err != nil {
		t.Fatalf("load notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected exactly 1 backfilled notification, got %d", len(notifications))
	}

	got := notifications[0]
	if got.StudentRollNo != student.RollNo {
		t.Errorf("expected the notification to belong to %s, got %s", student.RollNo, got.StudentRollNo)
	}
	if got.Title != models.NotificationTitleActivityReminder {
		t.Errorf("expected title %q, got %q", models.NotificationTitleActivityReminder, got.Title)
	}
	if got.Type != models.NotificationTypeActivity {
		t.Errorf("expected type %q, got %q", models.NotificationTypeActivity, got.Type)
	}
	if strings.Contains(got.Message, models.AdminNoteReminderPrefix) {
		t.Errorf("expected the internal marker to be stripped, got %q", got.Message)
	}
	if !strings.Contains(got.Message, "Python Workshop") {
		t.Errorf("expected the reminder text to be carried over, got %q", got.Message)
	}
	// The original timestamp keeps a backfilled reminder in its real place in
	// the newest-first ordering rather than surfacing as if it just arrived.
	if diff := got.CreatedAt.Sub(sentAt); diff > time.Second || diff < -time.Second {
		t.Errorf("expected the note's original timestamp %v to be preserved, got %v", sentAt, got.CreatedAt)
	}
	if got.IsRead {
		t.Errorf("expected a backfilled reminder to remain unread")
	}

	// Running again on an already-migrated database must change nothing.
	if err := backfillReminderNotifications(); err != nil {
		t.Fatalf("second backfill: %v", err)
	}
	var count int64
	if err := db.Model(&models.Notification{}).Count(&count).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected the backfill to be idempotent, got %d notifications after a second run", count)
	}
}
