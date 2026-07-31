package config

import (
	"log"
	"strings"

	"github.com/iips-oss/ispark/api/models"
)

// RunMigrations applies data migrations that must run on every boot, regardless
// of whether demo seeding is enabled. AutoMigrate (in ConnectDB) handles the
// schema; this handles data that needs to be reshaped to match current
// expectations — for example canonicalising legacy course names so production
// records line up with the report filters' canonical program list.
//
// Every migration here must be idempotent: it runs on each start and must be a
// no-op once the data is already in its target shape.
func RunMigrations() {
	if err := normalizeStudentCourses(); err != nil {
		// A migration failure should not stop the server from booting, but it
		// must be visible in the logs so it can be investigated.
		log.Printf("Migration warning: normalising student courses failed: %v", err)
	}
	if err := backfillReminderNotifications(); err != nil {
		log.Printf("Migration warning: backfilling reminder notifications failed: %v", err)
	}
}

// backfillReminderNotifications lifts activity-monitoring reminders that were
// recorded only as admin notes into the notifications table.
//
// Reminders used to be stored as an AdminNote and surfaced by a student
// endpoint that read the note trail. That endpoint now reads the notifications
// table, so without this backfill every reminder sent before the change would
// silently disappear from the student's bell.
//
// It is idempotent: a reminder that already has its matching notification is
// skipped, so this is a no-op from the second boot onwards. The note's original
// timestamp is carried over so backfilled reminders keep their place in the
// newest-first ordering instead of all surfacing as if they arrived today.
func backfillReminderNotifications() error {
	var notes []models.AdminNote
	if err := DB.Where("text LIKE ?", models.AdminNoteReminderPrefix+"%").Find(&notes).Error; err != nil {
		return err
	}

	for _, note := range notes {
		message := strings.TrimPrefix(note.Text, models.AdminNoteReminderPrefix)

		var existing int64
		if err := DB.Model(&models.Notification{}).
			Where("student_roll_no = ? AND title = ? AND message = ?",
				note.StudentRollNo, models.NotificationTitleActivityReminder, message).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}

		notification := models.Notification{
			StudentRollNo: note.StudentRollNo,
			Title:         models.NotificationTitleActivityReminder,
			Message:       message,
			Type:          models.NotificationTypeActivity,
			CreatedAt:     note.CreatedAt,
			UpdatedAt:     note.CreatedAt,
		}
		if err := DB.Create(&notification).Error; err != nil {
			return err
		}
	}

	return nil
}

// normalizeStudentCourses rewrites any legacy course name to its canonical
// equivalent. It is idempotent: rows already on a canonical name match nothing.
// It runs as a migration (see RunMigrations) so production records are
// normalised without depending on development seeding.
func normalizeStudentCourses() error {
	for legacy, canonical := range models.CourseNameAliases() {
		if err := DB.Model(&models.Student{}).
			Where("LOWER(TRIM(course_name)) = ?", legacy).
			Update("course_name", canonical).Error; err != nil {
			return err
		}
	}
	return nil
}
