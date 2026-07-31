package controllers

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
)

func errJSON(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

func AdminLogin(c *fiber.Ctx) error {
	var input models.AdminLoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse request body"})
	}
	if input.AdminID == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Admin ID and Password are required"})
	}

	var admin models.Admin
	if err := config.DB.Where("admin_id = ?", input.AdminID).First(&admin).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	if !utils.CheckPasswordHash(input.Password, admin.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	if admin.Status == "Inactive" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Your account is inactive. Please contact the administrator."})
	}

	accessToken, err := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate access token"})
	}
	return c.JSON(fiber.Map{
		"message":              "Admin logged in successfully",
		"access_token":         accessToken,
		"must_change_password": admin.MustChangePassword,
		"admin": fiber.Map{
			"admin_id": admin.AdminID,
			"name":     admin.Name,
			"role":     admin.Role,
		},
	})
}

// 1. POST /admin/change-password
func AdminChangePassword(c *fiber.Ctx) error {
	var input models.ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Cannot parse request body")
	}
	if input.CurrentPassword == "" || input.NewPassword == "" || input.ConfirmPassword == "" {
		return errJSON(c, fiber.StatusBadRequest, "All fields are required")
	}
	if input.NewPassword != input.ConfirmPassword {
		return errJSON(c, fiber.StatusBadRequest, "Passwords do not match")
	}

	// Validate password strength
	if err := utils.ValidatePasswordStrength(input.NewPassword); err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	adminID, ok := c.Locals("roll_no").(string)
	if !ok || adminID == "" {
		return errJSON(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var admin models.Admin
	if err := config.DB.Where("admin_id = ?", adminID).First(&admin).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Admin not found")
	}

	if !utils.CheckPasswordHash(input.CurrentPassword, admin.Password) {
		return errJSON(c, fiber.StatusUnauthorized, "Current password is incorrect")
	}

	newHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	admin.Password = newHash
	admin.MustChangePassword = false

	if err := config.DB.Save(&admin).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to update password")
	}

	return c.JSON(fiber.Map{"message": "Password changed successfully"})
}

func getAuthenticatedAdmin(c *fiber.Ctx) (*models.Admin, error) {
	adminID, ok := c.Locals("roll_no").(string)
	if !ok || adminID == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	var admin models.Admin
	if err := config.DB.Where("admin_id = ?", adminID).First(&admin).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	return &admin, nil
}

func scopeToAssignedBatch(query *gorm.DB, admin *models.Admin) (*gorm.DB, bool) {
	if admin.Role != "admin" {
		return query, true
	}
	if admin.AssignedBatch == "" {
		return query.Where("1 = 0"), false
	}

	return query.Where("roll_no LIKE ? AND LENGTH(roll_no) >= ?", admin.AssignedBatch+"%", len(admin.AssignedBatch)), true
}

func applyStudentStats(student *models.Student) {
	credits := 0
	pending := 0
	for _, cert := range student.Certificates {
		if cert.Status == "Approved" {
			credits += cert.Credits
		}
		if cert.Status == "Pending" {
			pending++
		}
	}

	student.CreditsEarned = credits
	student.PendingCertificates = pending
	student.TotalCertificates = len(student.Certificates)
	student.ActivityCount = len(student.Enrollments)
	if len(student.RollNo) >= 6 {
		student.Batch = student.RollNo[:6]
	}

	switch {
	case len(student.Enrollments) == 0 && len(student.Certificates) == 0:
		student.EngagementStatus = "Inactive"
	case pending > 0:
		student.EngagementStatus = "Pending Review"
	default:
		student.EngagementStatus = "Active"
	}
}

// 2. GET /api/admin/students -> View assigned students
func GetAllStudents(c *fiber.Ctx) error {
	currentUser, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var students []models.Student
	studentQuery := config.DB.Model(&models.Student{})
	studentQuery, scoped := scopeToAssignedBatch(studentQuery, currentUser)

	if !scoped {
		return c.JSON(fiber.Map{"students": []models.Student{}})
	}

	if err := studentQuery.Find(&students).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to retrieve students")
	}

	type StudentStats struct {
		RollNo              string `gorm:"column:roll_no"`
		CreditsEarned       int    `gorm:"column:credits_earned"`
		PendingCertificates int    `gorm:"column:pending_certificates"`
		TotalCertificates   int    `gorm:"column:total_certificates"`
		ActivityCount       int    `gorm:"column:activity_count"`
	}

	var stats []StudentStats
	statsQuery := config.DB.Model(&models.Student{}).Select(`
		roll_no,
		(SELECT COALESCE(SUM(credits), 0) FROM certificates WHERE certificates.student_roll_no = students.roll_no AND certificates.status = 'Approved') as credits_earned,
		(SELECT COUNT(*) FROM certificates WHERE certificates.student_roll_no = students.roll_no AND certificates.status = 'Pending') as pending_certificates,
		(SELECT COUNT(*) FROM certificates WHERE certificates.student_roll_no = students.roll_no) as total_certificates,
		(SELECT COUNT(*) FROM enrollments WHERE enrollments.student_roll_no = students.roll_no) as activity_count
	`)

	statsQuery, _ = scopeToAssignedBatch(statsQuery, currentUser)
	statsQuery.Scan(&stats)

	statsMap := make(map[string]StudentStats)
	for _, s := range stats {
		statsMap[s.RollNo] = s
	}

	for i, student := range students {
		if len(student.RollNo) >= 6 {
			students[i].Batch = student.RollNo[:6]
		}
		if stat, exists := statsMap[student.RollNo]; exists {
			students[i].CreditsEarned = stat.CreditsEarned
			students[i].PendingCertificates = stat.PendingCertificates
			students[i].TotalCertificates = stat.TotalCertificates
			students[i].ActivityCount = stat.ActivityCount

			if students[i].ActivityCount == 0 &&
				students[i].CreditsEarned == 0 &&
				students[i].PendingCertificates == 0 {
				students[i].EngagementStatus = "Inactive"
			} else if students[i].PendingCertificates > 0 {
				students[i].EngagementStatus = "Pending Review"
			} else {
				students[i].EngagementStatus = "Active"
			}
		}
	}

	return c.JSON(fiber.Map{"students": students})
}

// 3. GET /api/admin/students/:roll -> One student's detail
func GetStudentDetail(c *fiber.Ctx) error {
	roll := c.Params("roll")

	currentUser, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	query, scoped := scopeToAssignedBatch(
		config.DB.Preload("Enrollments.Activity").
			Preload("Certificates").
			Where("roll_no = ?", roll),
		currentUser,
	)
	if !scoped {
		return errJSON(c, fiber.StatusNotFound, "Student not found")
	}

	var student models.Student
	if err := query.First(&student).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Student not found")
	}
	applyStudentStats(&student)

	return c.JSON(fiber.Map{"student": student})
}

// Fetches student if they exist AND are within admin's batch scope
func getScopedStudent(_ *fiber.Ctx, roll string, admin *models.Admin) (*models.Student, error) {
	var student models.Student
	dbQuery, scoped := scopeToAssignedBatch(config.DB.Where("roll_no = ?", roll), admin)

	if !scoped || dbQuery.First(&student).Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Student not found or unauthorized")
	}
	return &student, nil
}

func GetMentorObservations(c *fiber.Ctx) error {
	roll := c.Params("roll")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	if _, err := getScopedStudent(c, roll, admin); err != nil {
		return errJSON(c, fiber.StatusNotFound, err.Error())
	}

	var observations []models.AdminNote
	config.DB.Where("student_roll_no = ?", roll).Order("created_at asc").Find(&observations)
	return c.JSON(fiber.Map{"observations": observations})
}

func AddMentorObservation(c *fiber.Ctx) error {
	roll := c.Params("roll")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	if _, err := getScopedStudent(c, roll, admin); err != nil {
		return errJSON(c, fiber.StatusNotFound, err.Error())
	}

	var input models.ObservationInput
	if err := c.BodyParser(&input); err != nil || input.Text == "" {
		return errJSON(c, fiber.StatusBadRequest, "Observation text is required")
	}

	note := models.AdminNote{
		StudentRollNo: roll,
		AdminID:       admin.AdminID,
		AuthorName:    admin.Name,
		Role:          "Mentor",
		Text:          input.Text,
	}

	if err := config.DB.Create(&note).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to save observation")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Observation added", "observation": note})
}

func EditMentorObservation(c *fiber.Ctx) error {
	roll := c.Params("roll")
	id := c.Params("id")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	if _, err := getScopedStudent(c, roll, admin); err != nil {
		return errJSON(c, fiber.StatusNotFound, err.Error())
	}

	var input models.ObservationInput
	if err := c.BodyParser(&input); err != nil || input.Text == "" {
		return errJSON(c, fiber.StatusBadRequest, "Observation text is required")
	}

	var note models.AdminNote
	if err := config.DB.Where("id = ? AND student_roll_no = ?", id, roll).First(&note).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Observation not found")
	}

	if note.AdminID != admin.AdminID {
		return errJSON(c, fiber.StatusForbidden, "Cannot edit another admin's observation")
	}

	note.Text = input.Text
	if err := config.DB.Save(&note).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to update observation")
	}

	return c.JSON(fiber.Map{"message": "Observation updated", "observation": note})
}

func SendStudentNotice(c *fiber.Ctx) error {
	roll := c.Params("roll")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	// 1. Parse the incoming JSON to get the custom message from the frontend
	var input models.NoticeInput
	if err := c.BodyParser(&input); err != nil || input.Message == "" {
		return errJSON(c, fiber.StatusBadRequest, "Notice message is required")
	}

	student, err := getScopedStudent(c, roll, admin)
	if err != nil {
		return errJSON(c, fiber.StatusNotFound, err.Error())
	}

	// 2. Inject the custom input.Message into the email template
	msg := fmt.Sprintf("Dear %s,\n\n%s\n\nRegards,\n%s", student.Name, input.Message, admin.Name)

	// A notice has to reach the student's notification bell as well as their
	// inbox, so the endpoint must not report success when only one of the two
	// landed. The notification is written first and the email is dispatched
	// inside the same transaction: if the notification cannot be stored no
	// email goes out at all, and if the email fails the notification is rolled
	// back, so a retry re-sends the whole notice rather than duplicating the
	// bell entry for an email the student never received.
	//
	// KNOWN LIMITATION (tracked reliability follow-up): the email is an
	// external side effect inside a database transaction, so the one window
	// this does not cover is the commit itself failing after the mail server
	// has already accepted the message. The rollback cannot unsend that email,
	// and the admin's retry sends a second copy. Closing it properly needs the
	// email to become a queued outbox row committed with the notification and
	// drained by a dispatcher, which is deliberately out of scope here rather
	// than half-built: an outbox without a retrying worker would simply move
	// the silent-loss window instead of removing it.
	var emailErr error
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := createNotificationTx(tx, student.RollNo, "New Notice", input.Message, models.NotificationTypeNotice); err != nil {
			return err
		}
		emailErr = utils.SendEmail(student.EmailID, "iSPARC Notice", msg)
		return emailErr
	}); err != nil {
		if emailErr != nil {
			return errJSON(c, fiber.StatusInternalServerError, "Failed to dispatch email")
		}
		return errJSON(c, fiber.StatusInternalServerError, "Failed to record the notice notification, please retry")
	}

	return c.JSON(fiber.Map{"message": "Notice sent successfully"})
}

func GetAdminDashboardStats(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	batchPrefix := ""
	if admin.Role == "admin" {
		batchPrefix = admin.AssignedBatch
	}

	// If admin has no assigned batch, return 0s
	if admin.Role == "admin" && batchPrefix == "" {
		return c.JSON(fiber.Map{
			"total_students":  0,
			"active_students": 0,
			"pending_reviews": 0,
			"average_credits": 0,
		})
	}

	var totalStudents int64
	studentQuery := config.DB.Model(&models.Student{})
	if batchPrefix != "" {
		studentQuery = studentQuery.Where("roll_no LIKE ?", batchPrefix+"%")
	}
	studentQuery.Count(&totalStudents)

	var activeStudents int64
	activeQuery := config.DB.Model(&models.Student{}).
		Where("EXISTS (SELECT 1 FROM enrollments WHERE enrollments.student_roll_no = students.roll_no) OR EXISTS (SELECT 1 FROM certificates WHERE certificates.student_roll_no = students.roll_no)")
	if batchPrefix != "" {
		activeQuery = activeQuery.Where("roll_no LIKE ?", batchPrefix+"%")
	}
	activeQuery.Count(&activeStudents)

	var pendingReviews int64
	certQuery := config.DB.Model(&models.Certificate{}).Where("certificates.status = ?", "Pending")
	if batchPrefix != "" {
		certQuery = certQuery.Joins("JOIN students on students.roll_no = certificates.student_roll_no").Where("students.roll_no LIKE ?", batchPrefix+"%")
	}
	certQuery.Count(&pendingReviews)

	var totalCredits int64
	type CreditResult struct {
		Total sql.NullInt64
	}
	var res CreditResult
	creditsQuery := config.DB.Model(&models.Certificate{}).Select("SUM(credits) as total").Where("certificates.status = ?", "Approved")
	if batchPrefix != "" {
		creditsQuery = creditsQuery.Joins("JOIN students on students.roll_no = certificates.student_roll_no").Where("students.roll_no LIKE ?", batchPrefix+"%")
	}
	creditsQuery.Scan(&res)
	if res.Total.Valid {
		totalCredits = res.Total.Int64
	}

	avgCredits := 0.0
	if totalStudents > 0 {
		avgCredits = float64(totalCredits) / float64(totalStudents)
	}

	return c.JSON(fiber.Map{
		"total_students":  totalStudents,
		"active_students": activeStudents,
		"pending_reviews": pendingReviews,
		"average_credits": avgCredits,
	})
}

func GetRecentActivities(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	// A batch-scoped admin with no assigned batch must not see other batches'
	// submissions, matching the scoping of the other admin endpoints.
	if admin.Role == "admin" && admin.AssignedBatch == "" {
		return c.JSON(fiber.Map{"recent_activities": []models.Certificate{}})
	}

	var recentCerts []models.Certificate
	query := config.DB.Preload("Student").Order("created_at desc").Limit(5)

	if admin.Role == "admin" && admin.AssignedBatch != "" {
		query = query.Where("student_roll_no LIKE ?", admin.AssignedBatch+"%")
	}

	query.Find(&recentCerts)
	return c.JSON(fiber.Map{"recent_activities": recentCerts})
}

func GetAdminProfile(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	batchPrefix := ""
	if admin.Role == "admin" {
		batchPrefix = admin.AssignedBatch
	}

	var assignedStudents int64
	var verifiedCertificates int64
	var pendingReviews int64
	var supervisedActivities int64

	if admin.Role == "superadmin" || batchPrefix != "" {
		studentQuery := config.DB.Model(&models.Student{})
		if batchPrefix != "" {
			studentQuery = studentQuery.Where("roll_no LIKE ?", batchPrefix+"%")
		}
		studentQuery.Count(&assignedStudents)

		certQuery := config.DB.Model(&models.Certificate{})
		if batchPrefix != "" {
			certQuery = certQuery.Joins("JOIN students on students.roll_no = certificates.student_roll_no").Where("students.roll_no LIKE ?", batchPrefix+"%")
		}
		certQuery.Where("certificates.status = ?", "Approved").Count(&verifiedCertificates)

		pendingQuery := config.DB.Model(&models.Certificate{})
		if batchPrefix != "" {
			pendingQuery = pendingQuery.Joins("JOIN students on students.roll_no = certificates.student_roll_no").Where("students.roll_no LIKE ?", batchPrefix+"%")
		}
		pendingQuery.Where("certificates.status = ?", "Pending").Count(&pendingReviews)

		config.DB.Model(&models.Activity{}).Where("coordinator_id = ?", admin.AdminID).Count(&supervisedActivities)

	}

	return c.JSON(fiber.Map{
		"admin": fiber.Map{
			"admin_id":       admin.AdminID,
			"name":           admin.Name,
			"email":          admin.Email,
			"role":           admin.Role,
			"assigned_batch": admin.AssignedBatch,
		},
		"stats": fiber.Map{
			"assigned_students":     assignedStudents,
			"verified_certificates": verifiedCertificates,
			"pending_reviews":       pendingReviews,
			"supervised_activities": supervisedActivities,
		},
	})
}

func UpdateAdminProfile(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Invalid request payload")
	}

	// Validate name
	if input.Name != "" {
		trimmedName := strings.TrimSpace(input.Name)
		if trimmedName == "" {
			return errJSON(c, fiber.StatusBadRequest, "Name cannot be blank")
		}
		if len(trimmedName) > 100 {
			return errJSON(c, fiber.StatusBadRequest, "Name cannot exceed 100 characters")
		}
		admin.Name = trimmedName
	}

	// Validate email
	if input.Email != "" {
		trimmedEmail := strings.TrimSpace(input.Email)
		if trimmedEmail == "" {
			return errJSON(c, fiber.StatusBadRequest, "Email cannot be blank")
		}
		if len(trimmedEmail) > 100 {
			return errJSON(c, fiber.StatusBadRequest, "Email cannot exceed 100 characters")
		}
		if !utils.IsValidEmail(trimmedEmail) {
			return errJSON(c, fiber.StatusBadRequest, "Invalid email format")
		}

		// Check for duplicate email across Admin and Student tables (case-insensitive)
		var existingAdmin models.Admin
		if err := config.DB.Where("LOWER(email) = LOWER(?) AND admin_id != ?", trimmedEmail, admin.AdminID).First(&existingAdmin).Error; err == nil {
			return errJSON(c, fiber.StatusConflict, "Email already in use")
		}

		var existingStudent models.Student
		if err := config.DB.Where("LOWER(email_id) = LOWER(?)", trimmedEmail).First(&existingStudent).Error; err == nil {
			return errJSON(c, fiber.StatusConflict, "Email already in use")
		}

		admin.Email = trimmedEmail
	}

	if err := config.DB.Save(&admin).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"admin": fiber.Map{
			"admin_id":       admin.AdminID,
			"name":           admin.Name,
			"email":          admin.Email,
			"role":           admin.Role,
			"assigned_batch": admin.AssignedBatch,
		},
	})
}

func GetCertificatesQueue(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	if admin.Role == "admin" && admin.AssignedBatch == "" {
		return c.JSON(fiber.Map{"certificates": []models.Certificate{}})
	}

	var certs []models.Certificate

	// Removed the fragile JOIN. Preload is all you need to attach the student data.
	certQuery := config.DB.Preload("Student").Order("created_at desc")

	if admin.Role == "admin" && admin.AssignedBatch != "" {
		// Filter using the column that already exists on the certificates table
		certQuery = certQuery.Where("student_roll_no LIKE ?", admin.AssignedBatch+"%")
	}

	if err := certQuery.Find(&certs).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to fetch certificates queue")
	}

	return c.JSON(fiber.Map{
		"certificates": certs,
	})
}

func ApproveCertificate(c *fiber.Ctx) error {
	return updateCertificateStatus(c, "Approved")
}

func RejectCertificate(c *fiber.Ctx) error {
	return updateCertificateStatus(c, "Rejected")
}

func updateCertificateStatus(c *fiber.Ctx, status string) error {
	id := c.Params("id")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var cert models.Certificate

	// Removed the fragile JOIN here as well
	certQuery := config.DB.Preload("Student").Where("id = ?", id)

	if admin.Role == "admin" {
		if admin.AssignedBatch == "" {
			return errJSON(c, fiber.StatusForbidden, "Admin not assigned to any batch")
		}
		// Validate access using the certificate's foreign key
		certQuery = certQuery.Where("student_roll_no LIKE ?", admin.AssignedBatch+"%")
	}

	if err := certQuery.First(&cert).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Certificate not found or access denied")
	}

	if status == "Rejected" {
		var input struct {
			Reason string `json:"reason"`
		}
		if err := c.BodyParser(&input); err == nil && input.Reason != "" {
			cert.RejectionReason = input.Reason
		}
	}

	previousStatus := cert.Status
	cert.Status = status

	// The decision and the student's notification are written in one
	// transaction. A student is entitled to be told about a decision on their
	// certificate, so a decision that cannot be delivered must not be recorded
	// either: the rollback leaves the certificate at its previous status, the
	// endpoint reports failure rather than success, and the same request can be
	// retried once notification storage recovers.
	//
	// On approval the notification doubles as the credits-awarded notice, since
	// credits are granted when a certificate is approved.
	//
	// Only an actual status transition is notifiable. Re-sending the same
	// decision (a retry of an already-delivered approval, or a double-clicked
	// approve button) writes no second notification, so the student never
	// receives two for one decision.
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&cert).Error; err != nil {
			return err
		}
		if previousStatus == status {
			return nil
		}

		switch status {
		case "Approved":
			message := fmt.Sprintf("Your certificate for %q has been approved.", cert.ActivityName)
			if cert.Credits > 0 {
				message = fmt.Sprintf("Your certificate for %q has been approved. You earned %d credits.", cert.ActivityName, cert.Credits)
			}
			return createNotificationTx(tx, cert.StudentRollNo, "Certificate Approved", message, models.NotificationTypeCertificate)
		case "Rejected":
			message := fmt.Sprintf("Your certificate for %q was rejected.", cert.ActivityName)
			if cert.RejectionReason != "" {
				message = fmt.Sprintf("Your certificate for %q was rejected. Reason: %s", cert.ActivityName, cert.RejectionReason)
			}
			return createNotificationTx(tx, cert.StudentRollNo, "Certificate Rejected", message, models.NotificationTypeCertificate)
		}
		return nil
	}); err != nil {
		cert.Status = previousStatus
		return errJSON(c, fiber.StatusInternalServerError, "Failed to update certificate status and notify the student, please retry")
	}

	return c.JSON(fiber.Map{"message": "Certificate status updated to " + status, "certificate": cert})
}

// AdminDownloadCertificate serves the submitted physical certificate file to admins with batch scoping.
func AdminDownloadCertificate(c *fiber.Ctx) error {
	id := c.Params("id")
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var cert models.Certificate
	certQuery := config.DB.Preload("Student").Where("id = ?", id)

	// Enforce batch scoping for regular admins
	if admin.Role == "admin" {
		if admin.AssignedBatch == "" {
			return errJSON(c, fiber.StatusForbidden, "Admin not assigned to any batch")
		}
		certQuery = certQuery.Where("student_roll_no LIKE ? AND LENGTH(student_roll_no) >= ?", admin.AssignedBatch+"%", len(admin.AssignedBatch))
	}

	if err := certQuery.First(&cert).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Certificate not found or access denied")
	}

	targetPath := cert.FilePath
	if targetPath == "" || func() bool { _, err := os.Stat(targetPath); return err != nil }() {
		if cert.FileName != "" {
			targetPath = filepath.Join("./uploads/certificates", filepath.Base(cert.FileName))
		}
	}

	if _, err := os.Stat(targetPath); err != nil {
		return errJSON(c, fiber.StatusNotFound, "Certificate file is no longer available on the server")
	}

	downloadName := cert.FileName
	if downloadName == "" {
		downloadName = filepath.Base(targetPath)
	}
	return c.Download(targetPath, downloadName)
}
