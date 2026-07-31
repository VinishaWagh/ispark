package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"gorm.io/gorm"
)

// defaultNotificationPageSize and maxNotificationPageSize bound how many
// notifications a single list request returns, keeping payloads predictable.
const (
	defaultNotificationPageSize = 20
	maxNotificationPageSize     = 100
)

// createNotificationTx persists a single notification for one student on the
// given handle. It is the single entry point every application event uses so
// notification creation stays consistent.
//
// Pass the transaction that carries the triggering state change when the two
// must be all-or-nothing (a certificate decision, an announcement going live);
// the returned error then rolls both back and leaves the event retryable. Pass
// config.DB for events where the notification is genuinely incidental — or use
// the best-effort createNotification wrapper below.
func createNotificationTx(tx *gorm.DB, rollNo, title, message, notificationType string) error {
	if rollNo == "" || title == "" {
		return nil
	}
	if notificationType == "" {
		notificationType = models.NotificationTypeGeneral
	}

	notification := models.Notification{
		StudentRollNo: rollNo,
		Title:         title,
		Message:       message,
		Type:          notificationType,
	}

	return tx.Create(&notification).Error
}

// There is deliberately no best-effort "fire and forget" form of the helper
// above. Every notification this system creates accompanies a state change the
// student is entitled to hear about, so a notification that cannot be persisted
// must fail its triggering request instead of being logged and swallowed: a
// swallowed failure commits the state change, leaves no notification, and the
// retry is then rejected as a duplicate, losing the notification permanently.
// Callers pass the transaction carrying their state change to createNotificationTx.

// createNotificationsForStudentsTx fans a single message out to many students in
// one batch insert (used by platform-wide events such as published
// announcements). Like createNotificationTx it reports failure to the caller so
// the triggering event can be rolled back and retried.
func createNotificationsForStudentsTx(tx *gorm.DB, rollNos []string, title, message, notificationType string) error {
	if len(rollNos) == 0 || title == "" {
		return nil
	}
	if notificationType == "" {
		notificationType = models.NotificationTypeGeneral
	}

	notifications := make([]models.Notification, 0, len(rollNos))
	for _, rollNo := range rollNos {
		if rollNo == "" {
			continue
		}
		notifications = append(notifications, models.Notification{
			StudentRollNo: rollNo,
			Title:         title,
			Message:       message,
			Type:          notificationType,
		})
	}
	if len(notifications) == 0 {
		return nil
	}

	return tx.Create(&notifications).Error
}

// GetNotifications returns the logged-in student's notifications, newest first,
// paginated. The response also carries the total and unread counts so the
// client can render the bell badge and pagination without extra round-trips.
func GetNotifications(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultNotificationPageSize)))
	if limit < 1 {
		limit = defaultNotificationPageSize
	}
	if limit > maxNotificationPageSize {
		limit = maxNotificationPageSize
	}
	offset := (page - 1) * limit

	var total int64
	if err := config.DB.Model(&models.Notification{}).Where("student_roll_no = ?", rollNo).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count notifications",
		})
	}

	var unreadCount int64
	if err := config.DB.Model(&models.Notification{}).Where("student_roll_no = ? AND is_read = ?", rollNo, false).Count(&unreadCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count unread notifications",
		})
	}

	var notifications []models.Notification
	// id is a monotonic tiebreaker so notifications created within the same
	// timestamp still return in a stable, newest-first order.
	if err := config.DB.Where("student_roll_no = ?", rollNo).
		Order("created_at desc, id desc").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch notifications",
		})
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// GetUnreadNotificationCount returns only the unread count, kept as a separate
// lightweight endpoint so the bell badge can poll without pulling the full list.
func GetUnreadNotificationCount(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var unreadCount int64
	if err := config.DB.Model(&models.Notification{}).Where("student_roll_no = ? AND is_read = ?", rollNo, false).Count(&unreadCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count unread notifications",
		})
	}

	return c.JSON(fiber.Map{"unread_count": unreadCount})
}

// MarkNotificationRead marks a single notification as read. The student can only
// mark their own notifications, enforced by scoping the update to their roll no.
func MarkNotificationRead(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)
	id := c.Params("id")

	result := config.DB.Model(&models.Notification{}).
		Where("id = ? AND student_roll_no = ?", id, rollNo).
		Update("is_read", true)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update notification",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	return c.JSON(fiber.Map{"message": "Notification marked as read"})
}

// MarkAllNotificationsRead marks every unread notification for the logged-in
// student as read and reports how many rows changed.
func MarkAllNotificationsRead(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	result := config.DB.Model(&models.Notification{}).
		Where("student_roll_no = ? AND is_read = ?", rollNo, false).
		Update("is_read", true)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update notifications",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "All notifications marked as read",
		"updated_count": result.RowsAffected,
	})
}
