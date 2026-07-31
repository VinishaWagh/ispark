package controllers

import (
	"strings"
	"time"

	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"gorm.io/gorm"
)

// announcementPublishFailedMessage is returned whenever an announcement could
// not both go live and reach students. The transaction is rolled back before it
// is sent, so the announcement keeps its previous status and the same request
// can simply be retried.
const announcementPublishFailedMessage = "Failed to publish announcement and notify students, please retry"

// activateAnnouncementTx is the single publication path for an announcement
// going live. Every transition to "active" routes through it: an explicit
// publish, a create or update that names the active status, and the scheduler
// flipping a due announcement. Nothing else may write status = "active".
//
// The status change, the NotifiedAt claim and the student notification inserts
// all happen on the caller's transaction, so the three are atomic. If the
// notifications cannot be written, the whole activation rolls back: the
// announcement stays in its previous status with NotifiedAt still NULL, the
// caller gets an error instead of a success response, and the event remains
// retryable — by re-issuing the request, or by the next status refresh for a
// scheduled announcement.
//
// Delivery stays exactly-once. NotifiedAt is claimed with a conditional update,
// so a second activation (a retry of an already-delivered publish, a
// double-clicked button, or an update that re-saves an active announcement)
// matches zero rows and skips the fan-out.
//
// publishDate, when non-nil, also moves the announcement's publish date — an
// explicit publish brings the date forward to today. The updated announcement
// is returned rather than mutated in place so a caller never holds state that a
// rolled-back transaction has undone.
func activateAnnouncementTx(tx *gorm.DB, announcement models.Announcement, publishDate *time.Time) (models.Announcement, error) {
	// Update only the lifecycle columns rather than saving the whole record, so
	// this write can never clobber the notified_at claim made below.
	updates := map[string]any{"status": "active"}
	if publishDate != nil {
		updates["publish_date"] = *publishDate
	}
	if err := tx.Model(&models.Announcement{}).
		Where("id = ?", announcement.ID).
		Updates(updates).Error; err != nil {
		return announcement, err
	}
	announcement.Status = "active"
	if publishDate != nil {
		announcement.PublishDate = *publishDate
	}

	notifiedAt := time.Now()
	claim := tx.Model(&models.Announcement{}).
		Where("id = ? AND notified_at IS NULL", announcement.ID).
		Update("notified_at", notifiedAt)
	if claim.Error != nil {
		return announcement, claim.Error
	}
	if claim.RowsAffected == 0 {
		// Someone already delivered this announcement; nothing left to send.
		return announcement, nil
	}
	announcement.NotifiedAt = &notifiedAt

	if err := fanOutAnnouncementToStudents(tx, announcement); err != nil {
		return announcement, err
	}
	return announcement, nil
}

// activateAnnouncement runs activateAnnouncementTx in its own transaction, for
// callers that have no wider unit of work to join. On failure the caller's
// announcement is returned unchanged.
func activateAnnouncement(announcement models.Announcement, publishDate *time.Time) (models.Announcement, error) {
	activated := announcement
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		activated, err = activateAnnouncementTx(tx, announcement, publishDate)
		return err
	}); err != nil {
		return announcement, err
	}
	return activated, nil
}

// fanOutAnnouncementToStudents creates a notification for every student when a
// published announcement is addressed to students (audience "Students" or
// "All Users"). "Mentors"-only announcements do not reach students, and are a
// successful no-op here.
func fanOutAnnouncementToStudents(tx *gorm.DB, announcement models.Announcement) error {
	audience := strings.ToLower(announcement.Audience)
	if audience != "students" && audience != "all users" {
		return nil
	}

	var rollNos []string
	if err := tx.Model(&models.Student{}).Pluck("roll_no", &rollNos).Error; err != nil {
		return err
	}

	return createNotificationsForStudentsTx(tx, rollNos, announcement.Title, announcement.Description, models.NotificationTypeAnnouncement)
}
