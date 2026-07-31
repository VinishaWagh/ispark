package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
)

// allowedSettingStatuses is the set of statuses a system setting may hold.
var allowedSettingStatuses = map[string]bool{
	"Active":   true,
	"Enabled":  true,
	"Disabled": true,
}

// platformUser is the flattened shape the super admin user registry renders.
// Students and admins are different tables, so they are normalised here.
type platformUser struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Role     string `json:"role"`
	Dept     string `json:"dept"`
	Status   string `json:"status"`
	Email    string `json:"email,omitempty"`
	Semester int    `json:"semester,omitempty"`
}

// createPlatformUserInput is what the super admin "Create User" form submits.
type createPlatformUserInput struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	ID       string `json:"id"`
	Email    string `json:"email"`
	Dept     string `json:"dept"`
	Semester int    `json:"semester"`
}

// generateTemporaryPassword returns a random password for a newly created
// account. The account holder is expected to change it on first login.
func generateTemporaryPassword() (string, error) {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return "iSPARC-" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// GetPlatformStats returns platform-wide counts for the super admin dashboard.
func GetPlatformStats(c *fiber.Ctx) error {
	var (
		totalStudents       int64
		totalAdmins         int64
		totalActivities     int64
		totalCertificates   int64
		pendingCertificates int64
		activeTracks        int64
	)

	counts := []struct {
		model any
		where string
		args  []any
		into  *int64
	}{
		{model: &models.Student{}, into: &totalStudents},
		{model: &models.Admin{}, into: &totalAdmins},
		{model: &models.Activity{}, into: &totalActivities},
		{model: &models.Certificate{}, into: &totalCertificates},
		{model: &models.Certificate{}, where: "status = ?", args: []any{"Pending"}, into: &pendingCertificates},
		{model: &models.Track{}, where: "status = ?", args: []any{"Active"}, into: &activeTracks},
	}

	for _, count := range counts {
		query := config.DB.Model(count.model)
		if count.where != "" {
			query = query.Where(count.where, count.args...)
		}
		if err := query.Count(count.into).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to load platform statistics",
			})
		}
	}

	return c.JSON(fiber.Map{
		"total_students":       totalStudents,
		"total_admins":         totalAdmins,
		"total_users":          totalStudents + totalAdmins,
		"total_activities":     totalActivities,
		"total_certificates":   totalCertificates,
		"pending_certificates": pendingCertificates,
		"active_tracks":        activeTracks,
	})
}

// GetPlatformUsers returns every student and admin account for the super admin
// user registry.
func GetPlatformUsers(c *fiber.Ctx) error {
	var students []models.Student
	if err := config.DB.
		Select("roll_no", "name", "course_name", "email_id", "semester", "is_verified", "status").
		Order("created_at desc").
		Find(&students).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load users",
		})
	}

	var admins []models.Admin
	if err := config.DB.
		Select("admin_id", "name", "role", "email", "assigned_batch", "status").
		Order("created_at desc").
		Find(&admins).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load users",
		})
	}

	users := make([]platformUser, 0, len(students)+len(admins))

	for _, admin := range admins {
		dept := admin.AssignedBatch
		if dept == "" {
			dept = "All Batches"
		}

		role := "Admin"
		if admin.Role == "superadmin" {
			role = "Super Admin"
		}

		adminStatus := admin.Status
		if adminStatus == "" {
			adminStatus = "Active"
		}

		users = append(users, platformUser{
			Name:     admin.Name,
			ID:       admin.AdminID,
			Role:     role,
			Dept:     dept,
			Status:   adminStatus,
			Email:    admin.Email,
			Semester: 0,
		})
	}

	for _, student := range students {
		status := student.Status
		if status == "" {
			status = "Pending"
			if student.IsVerified {
				status = "Active"
			}
		}

		users = append(users, platformUser{
			Name:     student.Name,
			ID:       student.RollNo,
			Role:     "Student",
			Dept:     student.CourseName,
			Status:   status,
			Email:    student.EmailID,
			Semester: student.Semester,
		})
	}

	return c.JSON(fiber.Map{"users": users})
}

// CreatePlatformUser registers a new student or admin account from the super
// admin user registry. The account is created with a generated temporary
// password, which is returned once so the super admin can pass it on.
func CreatePlatformUser(c *fiber.Ctx) error {
	var input createPlatformUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	input.Name = strings.TrimSpace(input.Name)
	input.ID = strings.TrimSpace(input.ID)
	input.Email = utils.NormalizeEmail(input.Email)
	input.Dept = strings.TrimSpace(input.Dept)

	if input.Name == "" || input.ID == "" || input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, ID and email are required",
		})
	}

	if !utils.ValidateEmail(input.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email address format",
		})
	}

	if input.Role != "Student" && input.Role != "Admin" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role must be either Student or Admin",
		})
	}

	// Preemptive uniqueness check
	var count int64
	// Check Student RollNo / AdminID
	config.DB.Model(&models.Student{}).Where("roll_no = ?", input.ID).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "A student with this roll number already exists"})
	}
	config.DB.Model(&models.Admin{}).Where("admin_id = ?", input.ID).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An admin with this ID already exists"})
	}

	// Check Student Email / Admin Email (case-insensitive checks)
	config.DB.Model(&models.Student{}).Where("LOWER(email_id) = ?", input.Email).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "A student with this email already exists"})
	}
	config.DB.Model(&models.Admin{}).Where("LOWER(email) = ?", input.Email).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An admin with this email already exists"})
	}

	tempPassword, err := generateTemporaryPassword()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create account",
		})
	}

	hashed, err := utils.HashPassword(tempPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create account",
		})
	}

	var created platformUser

	if input.Role == "Student" {
		semester := input.Semester
		if semester <= 0 {
			semester = 1
		}

		student := models.Student{
			RollNo:       input.ID,
			Name:         input.Name,
			CourseName:   input.Dept,
			Semester:     semester,
			EmailID:      input.Email,
			EnrollmentNo: "EN-" + input.ID,
			Password:     hashed,
			IsVerified:   false,
			Status:       "Pending",
		}

		if err := config.DB.Create(&student).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create student account",
			})
		}

		created = platformUser{
			Name:   student.Name,
			ID:     student.RollNo,
			Role:   "Student",
			Dept:   student.CourseName,
			Status: "Pending",
		}
	} else {
		admin := models.Admin{
			AdminID:            input.ID,
			Name:               input.Name,
			Email:              input.Email,
			Password:           hashed,
			Role:               "admin",
			AssignedBatch:      input.Dept,
			MustChangePassword: true,
			Status:             "Active",
		}

		if err := config.DB.Create(&admin).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create admin account",
			})
		}

		dept := admin.AssignedBatch
		if dept == "" {
			dept = "All Batches"
		}

		created = platformUser{
			Name:   admin.Name,
			ID:     admin.AdminID,
			Role:   "Admin",
			Dept:   dept,
			Status: admin.Status,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":               created,
		"temporary_password": tempPassword,
	})
}

// DeletePlatformUser removes a student or admin account. A super admin cannot
// delete their own account, and super admin accounts cannot be deleted here.
func DeletePlatformUser(c *fiber.Ctx) error {
	id := c.Params("id")

	callerID, _ := c.Locals("roll_no").(string)
	if id == callerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You cannot delete your own account",
		})
	}

	var admin models.Admin
	err := config.DB.Where("admin_id = ?", id).First(&admin).Error
	switch {
	case err == nil:
		if admin.Role == "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Super admin accounts cannot be deleted here",
			})
		}

		if err := config.DB.Delete(&admin).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete user",
			})
		}

		return c.JSON(fiber.Map{"message": "User deleted successfully"})

	case errors.Is(err, gorm.ErrRecordNotFound):
		// Not an admin, fall through and try students.

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	var student models.Student
	if err := config.DB.Where("roll_no = ?", id).First(&student).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	tx := config.DB.Begin()
	// Delete related OTPs, enrollments, certificates
	if err := tx.Where("email = ?", student.EmailID).Delete(&models.OTP{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete student OTPs"})
	}
	if err := tx.Where("student_roll_no = ?", student.RollNo).Delete(&models.Enrollment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete student enrollments"})
	}
	if err := tx.Where("student_roll_no = ?", student.RollNo).Delete(&models.Certificate{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete student certificates"})
	}
	if err := tx.Delete(&student).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete student"})
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// ---------------------------------------------------------------------------
// System Settings
// ---------------------------------------------------------------------------

// groupSettingsByCategory loads every setting ordered by SortOrder and groups
// them under their category, which is the shape the super admin settings screen
// renders (one list per tab).
func groupSettingsByCategory() (map[string][]models.SystemSetting, error) {
	var settings []models.SystemSetting
	if err := config.DB.Order("sort_order asc").Find(&settings).Error; err != nil {
		return nil, err
	}

	grouped := make(map[string][]models.SystemSetting)
	for _, setting := range settings {
		grouped[setting.Category] = append(grouped[setting.Category], setting)
	}

	return grouped, nil
}

// GetPlatformSettings returns every platform setting grouped by category for the
// super admin System Settings screen.
func GetPlatformSettings(c *fiber.Ctx) error {
	grouped, err := groupSettingsByCategory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load settings",
		})
	}

	return c.JSON(fiber.Map{"settings": grouped})
}

// UpdatePlatformSetting updates a single setting's value and/or status.
func UpdatePlatformSetting(c *fiber.Ctx) error {
	key := c.Params("key")

	var input models.UpdateSettingInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	if input.Value == nil && input.Status == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nothing to update: provide a value or a status",
		})
	}

	if input.Status != nil && !allowedSettingStatuses[*input.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status must be one of Active, Enabled or Disabled",
		})
	}

	var setting models.SystemSetting
	if err := config.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Setting not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load setting"})
	}

	if input.Value != nil {
		setting.Value = strings.TrimSpace(*input.Value)
	}
	if input.Status != nil {
		setting.Status = *input.Status
	}

	if err := config.DB.Save(&setting).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update setting"})
	}

	return c.JSON(fiber.Map{"setting": setting})
}

// UpdatePlatformSettings applies several setting updates in one request. It is
// what the settings screen's "Save" action uses so a whole tab can be persisted
// at once. All updates run in a transaction: if any key is unknown or invalid,
// nothing is written.
func UpdatePlatformSettings(c *fiber.Ctx) error {
	var input struct {
		Settings []models.BulkSettingUpdate `json:"settings"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	if len(input.Settings) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No settings provided"})
	}

	// Validate everything up front so the transaction never partially applies.
	for _, item := range input.Settings {
		if strings.TrimSpace(item.Key) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Each setting must include a key"})
		}
		if item.Value == nil && item.Status == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Setting %q has nothing to update", item.Key),
			})
		}
		if item.Status != nil && !allowedSettingStatuses[*item.Status] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid status %q for setting %q", *item.Status, item.Key),
			})
		}
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range input.Settings {
			updates := map[string]any{}
			if item.Value != nil {
				updates["value"] = strings.TrimSpace(*item.Value)
			}
			if item.Status != nil {
				updates["status"] = *item.Status
			}

			res := tx.Model(&models.SystemSetting{}).Where("key = ?", item.Key).Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("setting %q not found", item.Key)
			}
		}
		return nil
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	grouped, err := groupSettingsByCategory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reload settings"})
	}

	return c.JSON(fiber.Map{"settings": grouped})
}

// ---------------------------------------------------------------------------
// Activity Management
// ---------------------------------------------------------------------------

// platformActivity represents the shape the super admin activity management renders.
type platformActivity struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Track        string `json:"track"` // "Personality Development" or "Skill Building"
	Type         string `json:"type"`  // "Workshop", "Seminar", "Webinar", "Course"
	Credits      int    `json:"credits"`
	Status       string `json:"status"` // "Active" or "Inactive"
	Category     string `json:"category"`
	Description  string `json:"description"`
	Mode         string `json:"mode"`
	RegDeadline  string `json:"reg_deadline"`
	ActivityDate string `json:"activity_date"`
	Venue        string `json:"venue"`
	Coordinator  string `json:"coordinator"`
}

// createActivityInput represents the body submitted by Svelte.
type createActivityInput struct {
	Name         string `json:"name"`
	Track        string `json:"track"`
	Type         string `json:"type"`
	Credits      *int   `json:"credits"`
	Status       string `json:"status"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	Mode         string `json:"mode"`
	RegDeadline  string `json:"reg_deadline"`
	ActivityDate string `json:"activity_date"`
	Venue        string `json:"venue"`
	Coordinator  string `json:"coordinator"`
}

func parseActivityID(str string) uint {
	clean := strings.TrimPrefix(str, "ACT")
	id, err := strconv.Atoi(clean)
	if err != nil {
		return 0
	}
	return uint(id)
}

// GetPlatformActivities fetches all registered activities for superadmin management.
func GetPlatformActivities(c *fiber.Ctx) error {
	var activities []models.Activity
	if err := config.DB.Preload("Track").Order("created_at desc").Find(&activities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load activities",
		})
	}

	res := make([]platformActivity, 0, len(activities))
	for _, act := range activities {
		track := act.Track.Name
		if track == "" {
			track = "Personality Development"
			catUpper := strings.ToUpper(act.Category)
			if catUpper == "TECHNICAL" || catUpper == "RESEARCH" || catUpper == "SPORTS" || catUpper == "CULTURAL" {
				track = "Skill Building"
			}
		}

		actType := act.Type
		if actType == "" {
			actType = "Workshop"
			catUpper := strings.ToUpper(act.Category)
			if act.Mode == "Online" {
				actType = "Course"
			} else if catUpper == "PUBLIC SPEAKING" || catUpper == "LEADERSHIP" {
				actType = "Seminar"
			}
		}

		// Map backend Status to frontend Status
		status := "Inactive"
		if act.Status == "Open" || act.Status == "Closing Soon" || act.Status == "Active" {
			status = "Active"
		}

		// Formatted deadlines and activity dates
		regDeadlineStr := ""
		if !act.RegDeadline.IsZero() {
			regDeadlineStr = act.RegDeadline.Format("2006-01-02")
		}
		activityDateStr := ""
		if !act.ActivityDate.IsZero() {
			activityDateStr = act.ActivityDate.Format("2006-01-02")
		}

		res = append(res, platformActivity{
			ID:           fmt.Sprintf("ACT%03d", act.ID),
			Name:         act.Name,
			Track:        track,
			Type:         actType,
			Credits:      act.Credits,
			Status:       status,
			Category:     act.Category,
			Description:  act.Description,
			Mode:         act.Mode,
			RegDeadline:  regDeadlineStr,
			ActivityDate: activityDateStr,
			Venue:        act.Venue,
			Coordinator:  act.Coordinator,
		})
	}

	return c.JSON(fiber.Map{"activities": res})
}

// CreatePlatformActivity registers a new activity.
func CreatePlatformActivity(c *fiber.Ctx) error {
	var input createActivityInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Activity name is required"})
	}

	status := "Closed"
	if input.Status == "Active" {
		status = "Open"
	}

	var regDeadline time.Time
	if input.RegDeadline != "" {
		t, err := time.Parse("2006-01-02", input.RegDeadline)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid registration deadline format. Must be YYYY-MM-DD"})
		}
		regDeadline = t
	} else {
		regDeadline = time.Now().AddDate(0, 0, 7)
	}

	var activityDate time.Time
	if input.ActivityDate != "" {
		t, err := time.Parse("2006-01-02", input.ActivityDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid activity date format. Must be YYYY-MM-DD"})
		}
		activityDate = t
	} else {
		activityDate = time.Now().AddDate(0, 0, 14)
	}

	if !regDeadline.IsZero() && !activityDate.IsZero() && regDeadline.After(activityDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Registration deadline must be on or before the activity date"})
	}

	credits := 0
	if input.Credits != nil {
		credits = *input.Credits
	}

	var trackRecord models.Track
	trackName := strings.TrimSpace(input.Track)
	if trackName == "" {
		trackName = "Personality Development"
		catUpper := strings.ToUpper(input.Category)
		if catUpper == "TECHNICAL" || catUpper == "RESEARCH" || catUpper == "SPORTS" || catUpper == "CULTURAL" {
			trackName = "Skill Building"
		}
	}

	if err := config.DB.Where("LOWER(name) = LOWER(?)", trackName).First(&trackRecord).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Track '%s' does not exist", trackName)})
	}

	act := models.Activity{
		Name:         input.Name,
		Category:     input.Category,
		TrackID:      &trackRecord.ID,
		Type:         input.Type,
		Description:  input.Description,
		Credits:      credits,
		Mode:         input.Mode,
		RegDeadline:  regDeadline,
		ActivityDate: activityDate,
		Venue:        input.Venue,
		Coordinator:  input.Coordinator,
		Status:       status,
	}

	// An activity that is open for registration is an event students are
	// entitled to hear about, so its creation and the fan-out notifying them
	// are written in one transaction: if the notifications cannot be stored the
	// activity is not created either and the request can simply be retried,
	// rather than leaving a live activity nobody was told about.
	//
	// An activity created as Closed is not yet open to register for, so it
	// notifies nobody; students are told when it is opened, not when it is
	// drafted.
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&act).Error; err != nil {
			return err
		}
		if act.Status != "Open" {
			return nil
		}

		var rollNos []string
		if err := tx.Model(&models.Student{}).Pluck("roll_no", &rollNos).Error; err != nil {
			return err
		}

		message := fmt.Sprintf("A new activity %q is now open for registration.", act.Name)
		if !act.RegDeadline.IsZero() {
			message = fmt.Sprintf("A new activity %q is now open for registration. Register by %s.",
				act.Name, act.RegDeadline.Format("2006-01-02"))
		}

		return createNotificationsForStudentsTx(tx, rollNos, models.NotificationTitleNewActivity, message, models.NotificationTypeActivity)
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create activity"})
	}

	regDeadlineStr := ""
	if !act.RegDeadline.IsZero() {
		regDeadlineStr = act.RegDeadline.Format("2006-01-02")
	}
	activityDateStr := ""
	if !act.ActivityDate.IsZero() {
		activityDateStr = act.ActivityDate.Format("2006-01-02")
	}

	frontStatus := "Inactive"
	if act.Status == "Open" || act.Status == "Closing Soon" || act.Status == "Active" {
		frontStatus = "Active"
	}

	created := platformActivity{
		ID:           fmt.Sprintf("ACT%03d", act.ID),
		Name:         act.Name,
		Track:        trackRecord.Name,
		Type:         act.Type,
		Credits:      act.Credits,
		Status:       frontStatus,
		Category:     act.Category,
		Description:  act.Description,
		Mode:         act.Mode,
		RegDeadline:  regDeadlineStr,
		ActivityDate: activityDateStr,
		Venue:        act.Venue,
		Coordinator:  act.Coordinator,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"activity": created})
}

// UpdatePlatformActivity updates an existing activity.
func UpdatePlatformActivity(c *fiber.Ctx) error {
	rawID := c.Params("id")
	id := parseActivityID(rawID)
	if id == 0 {
		if val, err := strconv.Atoi(rawID); err == nil {
			id = uint(val)
		}
	}

	if id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid activity ID"})
	}

	var act models.Activity
	if err := config.DB.Preload("Track").First(&act, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Activity not found"})
	}

	var input createActivityInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	if input.Name != "" {
		act.Name = strings.TrimSpace(input.Name)
	}
	if input.Category != "" {
		act.Category = input.Category
	}

	var trackRecord models.Track
	resTrackName := act.Track.Name
	if strings.TrimSpace(input.Track) != "" {
		trackName := strings.TrimSpace(input.Track)
		if err := config.DB.Where("LOWER(name) = LOWER(?)", trackName).First(&trackRecord).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Track '%s' does not exist", trackName)})
		}
		act.TrackID = &trackRecord.ID
		resTrackName = trackRecord.Name
	}

	if input.Type != "" {
		act.Type = input.Type
	}
	if input.Description != "" {
		act.Description = input.Description
	}
	if input.Mode != "" {
		act.Mode = input.Mode
	}
	newRegDeadline := act.RegDeadline
	newActivityDate := act.ActivityDate

	if input.RegDeadline != "" {
		t, err := time.Parse("2006-01-02", input.RegDeadline)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid registration deadline format. Must be YYYY-MM-DD"})
		}
		newRegDeadline = t
	}
	if input.ActivityDate != "" {
		t, err := time.Parse("2006-01-02", input.ActivityDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid activity date format. Must be YYYY-MM-DD"})
		}
		newActivityDate = t
	}

	if !newRegDeadline.IsZero() && !newActivityDate.IsZero() && newRegDeadline.After(newActivityDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Registration deadline must be on or before the activity date"})
	}

	act.RegDeadline = newRegDeadline
	act.ActivityDate = newActivityDate
	if input.Venue != "" {
		act.Venue = input.Venue
	}
	if input.Coordinator != "" {
		act.Coordinator = input.Coordinator
	}
	if input.Status != "" {
		if input.Status == "Active" {
			act.Status = "Open"
		} else {
			act.Status = "Closed"
		}
	}
	if input.Credits != nil {
		act.Credits = *input.Credits
	}

	if err := config.DB.Save(&act).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update activity"})
	}

	regDeadlineStr := ""
	if !act.RegDeadline.IsZero() {
		regDeadlineStr = act.RegDeadline.Format("2006-01-02")
	}
	activityDateStr := ""
	if !act.ActivityDate.IsZero() {
		activityDateStr = act.ActivityDate.Format("2006-01-02")
	}

	frontStatus := "Inactive"
	if act.Status == "Open" || act.Status == "Closing Soon" || act.Status == "Active" {
		frontStatus = "Active"
	}

	if resTrackName == "" {
		if act.TrackID != nil && *act.TrackID != 0 {
			var t models.Track
			if err := config.DB.First(&t, *act.TrackID).Error; err == nil {
				resTrackName = t.Name
			}
		}
	}

	updated := platformActivity{
		ID:           fmt.Sprintf("ACT%03d", act.ID),
		Name:         act.Name,
		Track:        resTrackName,
		Type:         act.Type,
		Credits:      act.Credits,
		Status:       frontStatus,
		Category:     act.Category,
		Description:  act.Description,
		Mode:         act.Mode,
		RegDeadline:  regDeadlineStr,
		ActivityDate: activityDateStr,
		Venue:        act.Venue,
		Coordinator:  act.Coordinator,
	}

	return c.JSON(fiber.Map{"activity": updated})
}

// DeletePlatformActivity removes an activity.
func DeletePlatformActivity(c *fiber.Ctx) error {
	rawID := c.Params("id")
	id := parseActivityID(rawID)
	if id == 0 {
		if val, err := strconv.Atoi(rawID); err == nil {
			id = uint(val)
		}
	}

	if id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid activity ID"})
	}

	var act models.Activity
	if err := config.DB.First(&act, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Activity not found"})
	}

	tx := config.DB.Begin()
	// Cascade delete related enrollments
	if err := tx.Where("activity_id = ?", act.ID).Delete(&models.Enrollment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete activity enrollments"})
	}
	if err := tx.Delete(&act).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete activity"})
	}
	tx.Commit()

	return c.JSON(fiber.Map{"message": "Activity deleted successfully"})
}

// ---------------------------------------------------------------------------
// User Management
// ---------------------------------------------------------------------------

type updatePlatformUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Dept     string `json:"dept"`
	Semester int    `json:"semester"`
	Status   string `json:"status"` // Active, Inactive, Pending
}

// UpdatePlatformUser edits student or admin details.
func UpdatePlatformUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID is required"})
	}

	var input updatePlatformUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}

	// First try to check if the user is an Admin
	var admin models.Admin
	err := config.DB.Where("admin_id = ?", id).First(&admin).Error
	if err == nil {
		if input.Name != "" {
			admin.Name = input.Name
		}
		if input.Email != "" {
			input.Email = utils.NormalizeEmail(input.Email)
			if !utils.ValidateEmail(input.Email) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email address format"})
			}
			var count int64
			config.DB.Model(&models.Admin{}).Where("LOWER(email) = ? AND admin_id != ?", input.Email, admin.AdminID).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An admin with this email already exists"})
			}
			config.DB.Model(&models.Student{}).Where("LOWER(email_id) = ?", input.Email).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "A student with this email already exists"})
			}
			admin.Email = input.Email
		}
		if input.Dept != "" {
			admin.AssignedBatch = input.Dept
		}
		if input.Status != "" {
			if input.Status != "Active" && input.Status != "Inactive" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Status must be either Active or Inactive",
				})
			}
			admin.Status = input.Status
		}

		if err := config.DB.Save(&admin).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update admin"})
		}

		return c.JSON(fiber.Map{"message": "Admin updated successfully"})
	}

	// If not an admin, try Student
	var student models.Student
	err = config.DB.Where("roll_no = ?", id).First(&student).Error
	if err == nil {
		if input.Name != "" {
			student.Name = input.Name
		}
		if input.Email != "" {
			input.Email = utils.NormalizeEmail(input.Email)
			if !utils.ValidateEmail(input.Email) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email address format"})
			}
			var count int64
			config.DB.Model(&models.Student{}).Where("LOWER(email_id) = ? AND roll_no != ?", input.Email, student.RollNo).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "A student with this email already exists"})
			}
			config.DB.Model(&models.Admin{}).Where("LOWER(email) = ?", input.Email).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An admin with this email already exists"})
			}
			student.EmailID = input.Email
		}
		if input.Dept != "" {
			student.CourseName = input.Dept
		}
		if input.Semester > 0 {
			student.Semester = input.Semester
		}

		// Map Status to IsVerified
		if input.Status != "" {
			if input.Status != "Active" && input.Status != "Inactive" && input.Status != "Pending" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Status must be Active, Inactive or Pending",
				})
			}
			student.Status = input.Status
			student.IsVerified = (input.Status == "Active")
		}

		if err := config.DB.Save(&student).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update student"})
		}

		return c.JSON(fiber.Map{"message": "Student updated successfully"})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
}
