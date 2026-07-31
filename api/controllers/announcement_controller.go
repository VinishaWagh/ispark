package controllers

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"gorm.io/gorm"
)

const (
	announcementDateLayout     = "2006-01-02"
	maxAnnouncementTitleLength = 200
	indiaUTCOffsetSeconds      = 5*60*60 + 30*60
)

var announcementCategories = map[string]string{
	"general":    "General",
	"academic":   "Academic",
	"activities": "Activities",
	"events":     "Events",
}

var announcementAudiences = map[string]string{
	"students":  "Students",
	"mentors":   "Mentors",
	"all users": "All Users",
}

var announcementPriorities = map[string]string{
	"low":    "Low",
	"medium": "Medium",
	"high":   "High",
}

var announcementStatuses = map[string]string{
	"draft":     "draft",
	"scheduled": "scheduled",
	"active":    "active",
	"expired":   "expired",
}

type announcementInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Audience    *string `json:"audience"`
	Priority    *string `json:"priority"`
	PublishDate *string `json:"publish_date"`
	ExpiryDate  *string `json:"expiry_date"`
	Status      *string `json:"status"`
}

type announcementResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Audience    string `json:"audience"`
	Priority    string `json:"priority"`
	PublishDate string `json:"publish_date"`
	ExpiryDate  string `json:"expiry_date"`
	Status      string `json:"status"`
}

func announcementToday() time.Time {
	now := time.Now().In(time.FixedZone("Asia/Kolkata", indiaUTCOffsetSeconds))
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func announcementToResponse(announcement models.Announcement) announcementResponse {
	return announcementResponse{
		ID:          announcement.ID,
		Title:       announcement.Title,
		Description: announcement.Description,
		Category:    announcement.Category,
		Audience:    announcement.Audience,
		Priority:    announcement.Priority,
		PublishDate: announcement.PublishDate.Format(announcementDateLayout),
		ExpiryDate:  announcement.ExpiryDate.Format(announcementDateLayout),
		Status:      announcement.Status,
	}
}

func canonicalAnnouncementValue(value string, allowed map[string]string, field string) (string, error) {
	canonical, ok := allowed[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return "", fmt.Errorf("invalid announcement %s", field)
	}
	return canonical, nil
}

func parseAnnouncementDate(value, field string) (time.Time, error) {
	parsed, err := time.Parse(announcementDateLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must use YYYY-MM-DD format", field)
	}
	return parsed, nil
}

func announcementInputEmpty(input announcementInput) bool {
	return input.Title == nil && input.Description == nil && input.Category == nil &&
		input.Audience == nil && input.Priority == nil && input.PublishDate == nil &&
		input.ExpiryDate == nil && input.Status == nil
}

func applyAnnouncementInput(announcement *models.Announcement, input announcementInput, creating bool) error {
	if input.Title != nil {
		announcement.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		announcement.Description = strings.TrimSpace(*input.Description)
	}
	if input.Category != nil {
		category, err := canonicalAnnouncementValue(*input.Category, announcementCategories, "category")
		if err != nil {
			return err
		}
		announcement.Category = category
	}
	if input.Audience != nil {
		audience, err := canonicalAnnouncementValue(*input.Audience, announcementAudiences, "audience")
		if err != nil {
			return err
		}
		announcement.Audience = audience
	}
	if input.Priority != nil {
		priority, err := canonicalAnnouncementValue(*input.Priority, announcementPriorities, "priority")
		if err != nil {
			return err
		}
		announcement.Priority = priority
	}
	if input.PublishDate != nil {
		publishDate, err := parseAnnouncementDate(*input.PublishDate, "publish_date")
		if err != nil {
			return err
		}
		announcement.PublishDate = publishDate
	}
	if input.ExpiryDate != nil {
		expiryDate, err := parseAnnouncementDate(*input.ExpiryDate, "expiry_date")
		if err != nil {
			return err
		}
		announcement.ExpiryDate = expiryDate
	}
	if input.Status != nil {
		status, err := canonicalAnnouncementValue(*input.Status, announcementStatuses, "status")
		if err != nil {
			return err
		}
		announcement.Status = status
	} else if creating {
		announcement.Status = "draft"
	}
	if input.Priority == nil && creating {
		announcement.Priority = "Medium"
	}

	return validateAnnouncement(announcement)
}

func validateAnnouncement(announcement *models.Announcement) error {
	if announcement.Title == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(announcement.Title) > maxAnnouncementTitleLength {
		return fmt.Errorf("title must be %d characters or fewer", maxAnnouncementTitleLength)
	}
	if _, ok := announcementCategories[strings.ToLower(announcement.Category)]; !ok {
		return errors.New("category is required")
	}
	if _, ok := announcementAudiences[strings.ToLower(announcement.Audience)]; !ok {
		return errors.New("audience is required")
	}
	if _, ok := announcementPriorities[strings.ToLower(announcement.Priority)]; !ok {
		return errors.New("priority is required")
	}
	if _, ok := announcementStatuses[strings.ToLower(announcement.Status)]; !ok {
		return errors.New("status is required")
	}
	if announcement.PublishDate.IsZero() {
		return errors.New("publish_date is required")
	}
	if announcement.ExpiryDate.IsZero() {
		return errors.New("expiry_date is required")
	}
	if announcement.ExpiryDate.Before(announcement.PublishDate) {
		return errors.New("expiry_date cannot be before publish_date")
	}

	today := announcementToday()
	switch announcement.Status {
	case "scheduled":
		if !announcement.PublishDate.After(today) {
			return errors.New("scheduled announcements require a future publish_date")
		}
	case "active":
		if announcement.PublishDate.After(today) {
			return errors.New("active announcements cannot have a future publish_date")
		}
		if announcement.ExpiryDate.Before(today) {
			return errors.New("active announcements cannot have a past expiry_date")
		}
	case "expired":
		if !announcement.ExpiryDate.Before(today) {
			return errors.New("expired announcements require a past expiry_date")
		}
	}

	return nil
}

// refreshAnnouncementStatuses applies date-driven lifecycle transitions. Drafts
// remain drafts until explicitly published, even after their configured date.
func refreshAnnouncementStatuses() error {
	today := announcementToday()
	if err := config.DB.Model(&models.Announcement{}).
		Where("status IN ? AND expiry_date < ?", []string{"active", "scheduled"}, today).
		Update("status", "expired").Error; err != nil {
		return err
	}

	// A scheduled announcement reaching its publish date is a publication like
	// any other, so it goes live through the publication service rather than a
	// bare status update — otherwise students are never told it exists.
	var due []models.Announcement
	if err := config.DB.
		Where("status = ? AND publish_date <= ? AND expiry_date >= ?", "scheduled", today, today).
		Find(&due).Error; err != nil {
		return err
	}

	// Each announcement activates in its own transaction: one that cannot be
	// delivered stays scheduled and is retried by the next refresh, without
	// holding back the ones that can.
	var firstErr error
	for _, announcement := range due {
		if _, err := activateAnnouncement(announcement, nil); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func GetAnnouncements(c *fiber.Ctx) error {
	if err := refreshAnnouncementStatuses(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to refresh announcements"})
	}

	query := config.DB.Model(&models.Announcement{}).Order("created_at desc")
	if status := strings.TrimSpace(c.Query("status")); status != "" && !strings.EqualFold(status, "All") {
		canonical, err := canonicalAnnouncementValue(status, announcementStatuses, "status")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		query = query.Where("status = ?", canonical)
	}
	if search := strings.ToLower(strings.TrimSpace(c.Query("q"))); search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(category) LIKE ? OR LOWER(audience) LIKE ?", like, like, like)
	}

	var announcements []models.Announcement
	if err := query.Find(&announcements).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcements"})
	}

	result := make([]announcementResponse, 0, len(announcements))
	for _, announcement := range announcements {
		result = append(result, announcementToResponse(announcement))
	}
	return c.JSON(fiber.Map{"announcements": result})
}

func GetAnnouncementStats(c *fiber.Ctx) error {
	if err := refreshAnnouncementStatuses(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to refresh announcements"})
	}

	var total, active, scheduled, expired int64
	if err := config.DB.Model(&models.Announcement{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement statistics"})
	}
	for status, count := range map[string]*int64{
		"active": &active, "scheduled": &scheduled, "expired": &expired,
	} {
		if err := config.DB.Model(&models.Announcement{}).Where("status = ?", status).Count(count).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement statistics"})
		}
	}

	return c.JSON(fiber.Map{
		"total_announcements":     total,
		"active_announcements":    active,
		"scheduled_announcements": scheduled,
		"expired_announcements":   expired,
	})
}

func GetAnnouncement(c *fiber.Ctx) error {
	if err := refreshAnnouncementStatuses(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to refresh announcements"})
	}

	var announcement models.Announcement
	if err := config.DB.First(&announcement, c.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Announcement not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement"})
	}
	return c.JSON(fiber.Map{"announcement": announcementToResponse(announcement)})
}

func CreateAnnouncement(c *fiber.Ctx) error {
	var input announcementInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	announcement := models.Announcement{}
	if err := applyAnnouncementInput(&announcement, input, true); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Creating an announcement directly as "active" publishes it, so the insert
	// and the student fan-out share one transaction: the row cannot exist as
	// active without its notifications.
	if announcement.Status == "active" {
		created := announcement
		if err := config.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
			activated, err := activateAnnouncementTx(tx, created, nil)
			if err != nil {
				return err
			}
			created = activated
			return nil
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": announcementPublishFailedMessage})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"announcement": announcementToResponse(created)})
	}

	if err := config.DB.Create(&announcement).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create announcement"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"announcement": announcementToResponse(announcement)})
}

func UpdateAnnouncement(c *fiber.Ctx) error {
	var announcement models.Announcement
	if err := config.DB.First(&announcement, c.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Announcement not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement"})
	}

	var input announcementInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}
	if announcementInputEmpty(input) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No announcement fields provided"})
	}
	if err := applyAnnouncementInput(&announcement, input, false); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// An update that sets the status to "active" publishes the announcement, so
	// it takes the publication path in the same transaction as the save. An
	// already-delivered announcement re-saved as active is left alone by the
	// NotifiedAt claim, so editing a live announcement never re-notifies.
	if announcement.Status == "active" {
		updated := announcement
		if err := config.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&updated).Error; err != nil {
				return err
			}
			activated, err := activateAnnouncementTx(tx, updated, nil)
			if err != nil {
				return err
			}
			updated = activated
			return nil
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": announcementPublishFailedMessage})
		}
		return c.JSON(fiber.Map{"announcement": announcementToResponse(updated)})
	}

	if err := config.DB.Save(&announcement).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update announcement"})
	}
	return c.JSON(fiber.Map{"announcement": announcementToResponse(announcement)})
}

func DeleteAnnouncement(c *fiber.Ctx) error {
	var announcement models.Announcement
	if err := config.DB.First(&announcement, c.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Announcement not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement"})
	}
	if err := config.DB.Delete(&announcement).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete announcement"})
	}
	return c.JSON(fiber.Map{"message": "Announcement deleted successfully"})
}

func PublishAnnouncement(c *fiber.Ctx) error {
	var announcement models.Announcement
	if err := config.DB.First(&announcement, c.Params("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Announcement not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load announcement"})
	}

	today := announcementToday()
	if announcement.ExpiryDate.Before(today) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot publish an expired announcement"})
	}

	// Publishing brings the announcement live and delivers it to students'
	// notification bells atomically. A delivery failure rolls the whole
	// transition back and is reported as an error, so the announcement is never
	// left active-but-undelivered and this request can simply be retried.
	published, err := activateAnnouncement(announcement, &today)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": announcementPublishFailedMessage})
	}

	return c.JSON(fiber.Map{"announcement": announcementToResponse(published)})
}
