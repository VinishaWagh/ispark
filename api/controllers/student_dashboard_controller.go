package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
)

const certificateUploadDir = "./uploads/certificates"

// keeping it like this soo that adding more file types in future is easy.
// If we want to allow more file types, we can just add them here.
var allowedCertificateTypes = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}

// max certificate file size is 5 MB but i highly doublt that we should keep it 5 MB.
const maxCertificateSize = 5 * 1024 * 1024

// StudentNotificationResponse represents notification items sent to the student portal.
type StudentNotificationResponse struct {
	ID        uint      `json:"id"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Time      string    `json:"time"`
	Unread    bool      `json:"unread"`
}

// Deprecated: GetStudentNotifications reads the admin-note trail and is no
// longer routed. GET /api/student/notifications is served by GetNotifications,
// which reads the notifications table the portal's bell renders.
//
// Do not register this on that path again: it was previously registered ahead
// of GetNotifications and shadowed it, so events that wrote a real notification
// were invisible to students while the endpoint still returned 200. Tests must
// assert against the routed handler for the same reason.
func GetStudentNotifications(c *fiber.Ctx) error {
	rollNo, ok := c.Locals("roll_no").(string)
	if !ok || rollNo == "" {
		return errJSON(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var notes []models.AdminNote
	if err := config.DB.Where("student_roll_no = ?", rollNo).Order("created_at desc").Find(&notes).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to fetch notifications")
	}

	response := make([]StudentNotificationResponse, 0, len(notes))
	for _, note := range notes {
		timeAgo := time.Since(note.CreatedAt)
		timeStr := "Just now"
		if timeAgo.Hours() >= 24 {
			timeStr = fmt.Sprintf("%d days ago", int(timeAgo.Hours()/24))
		} else if timeAgo.Hours() >= 1 {
			timeStr = fmt.Sprintf("%d hours ago", int(timeAgo.Hours()))
		} else if timeAgo.Minutes() >= 1 {
			timeStr = fmt.Sprintf("%d mins ago", int(timeAgo.Minutes()))
		}

		cleanText := strings.TrimPrefix(note.Text, "[ALERT REMINDER] ")
		cleanText = strings.TrimPrefix(cleanText, "[NOTICE] ")

		response = append(response, StudentNotificationResponse{
			ID:        note.ID,
			Text:      cleanText,
			Author:    note.AuthorName,
			Role:      note.Role,
			CreatedAt: note.CreatedAt,
			Time:      timeStr,
			Unread:    true,
		})
	}

	return c.JSON(response)
}

// GetCertificates returns student's uploaded certificates
func GetCertificates(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var certificates []models.Certificate
	if err := config.DB.Where("student_roll_no = ?", rollNo).Order("created_at desc").Find(&certificates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch certificates",
		})
	}

	return c.JSON(certificates)
}

// UploadCertificate uploads a new certificate (handles file upload + details)
func UploadCertificate(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	activityName := c.FormValue("activity_name")
	activityCategory := c.FormValue("activity_category")
	activityDateStr := c.FormValue("activity_date")
	organizerName := c.FormValue("organizer_name")
	eventLevel := c.FormValue("event_level")
	certNumber := c.FormValue("cert_number")
	issueDateStr := c.FormValue("issue_date")
	participationType := c.FormValue("participation_type")
	description := c.FormValue("description")

	if activityName == "" || participationType == "" || activityDateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Required fields are missing",
		})
	}

	// Parse dates
	activityDate, err := time.Parse("2006-01-02", activityDateStr)
	if err != nil || activityDate.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid activity date format. Expected YYYY-MM-DD",
		})
	}
	var issueDate *time.Time
	if issueDateStr != "" {
		parsedDate, parseErr := time.Parse("2006-01-02", issueDateStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid issue date format. Expected YYYY-MM-DD",
			})
		}
		issueDate = &parsedDate
	}

	credits := config.CreditsForCertificate(participationType, eventLevel)

	// Handle File Upload
	file, err := c.FormFile("certificate_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Certificate file is required",
		})
	}

	// Validate file size (max 5 MB)
	if file.Size > maxCertificateSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File size exceeds 5 MB limit",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	expectedType, extAllowed := allowedCertificateTypes[ext]
	if !extAllowed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unsupported file format. Please upload PDF, PNG, or JPG",
		})
	}

	detectedType, err := detectFileType(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read uploaded file",
		})
	}
	if detectedType != expectedType {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File contents do not match its extension. Please upload a valid PDF, PNG, or JPG",
		})
	}

	// Ensure directory exists
	if err := os.MkdirAll(certificateUploadDir, 0o755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}

	fileName := fmt.Sprintf("%s_%d_%s", rollNo, time.Now().UnixNano(), filepath.Base(file.Filename))
	filePath := filepath.Join(certificateUploadDir, fileName)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	cert := models.Certificate{
		StudentRollNo:     rollNo,
		ActivityName:      activityName,
		ActivityCategory:  activityCategory,
		ActivityDate:      activityDate,
		OrganizerName:     organizerName,
		EventLevel:        eventLevel,
		CertNumber:        certNumber,
		IssueDate:         issueDate,
		ParticipationType: participationType,
		Description:       description,
		FileName:          fileName,
		FilePath:          filePath,
		Credits:           credits,
		Status:            "Pending",
	}

	if err := config.DB.Create(&cert).Error; err != nil {
		_ = os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save certificate record",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Certificate uploaded successfully",
		"certificate": cert,
	})
}

func detectFileType(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	head := make([]byte, 512)
	n, err := f.Read(head)
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.SplitN(http.DetectContentType(head[:n]), ";", 2)[0], nil
}

// now a student can only download their own certificate files.
func DownloadCertificate(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var cert models.Certificate
	if err := config.DB.Where("id = ? AND student_roll_no = ?", c.Params("id"), rollNo).First(&cert).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Certificate not found",
		})
	}

	filePath := filepath.Join(certificateUploadDir, filepath.Base(cert.FileName))
	if _, err := os.Stat(filePath); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Certificate file is no longer available",
		})
	}

	return c.Download(filePath, cert.FileName)
}

// getAcademicYearDateRange returns start and end date of the academic year
func getAcademicYearDateRange(yearStr string) (time.Time, time.Time, error) {
	if len(yearStr) != 7 || yearStr[4] != '-' {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid year format, expected YYYY-YY")
	}

	startYear, err := strconv.Atoi(yearStr[0:4])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start year: %v", err)
	}

	endYearPart, err := strconv.Atoi(yearStr[5:7])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end year suffix: %v", err)
	}

	expectedEndYearPart := (startYear + 1) % 100
	if endYearPart != expectedEndYearPart {
		return time.Time{}, time.Time{}, fmt.Errorf("academic year mismatch: suffix must be %02d", expectedEndYearPart)
	}

	endYear := startYear + 1
	startDate := time.Date(startYear, time.July, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(endYear, time.June, 30, 23, 59, 59, 999999999, time.UTC)

	return startDate, endDate, nil
}

// getCurrentAcademicYearRange returns the current academic year dates
func getCurrentAcademicYearRange() (time.Time, time.Time) {
	now := time.Now()
	var startYear int
	if now.Month() >= time.July {
		startYear = now.Year()
	} else {
		startYear = now.Year() - 1
	}
	startDate := time.Date(startYear, time.July, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(startYear+1, time.June, 30, 23, 59, 59, 999999999, time.UTC)
	return startDate, endDate
}

// GetLeaderboard returns the leaderboard sorted by total credits for a given academic year
func GetLeaderboard(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)
	year := c.Query("year")

	var startDate, endDate time.Time
	var err error
	if year != "" {
		startDate, endDate, err = getAcademicYearDateRange(year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid year format. Expected YYYY-YY (e.g., 2025-26)",
			})
		}
	} else {
		startDate, endDate = getCurrentAcademicYearRange()
	}

	type LeaderboardEntry struct {
		RollNo     string `json:"roll_no"`
		Name       string `json:"name"`
		CourseName string `json:"course_name"`
		Semester   int    `json:"semester"`
		Points     int    `json:"points"`
		IsSelf     bool   `json:"is_self"`
	}

	var entries []LeaderboardEntry

	err = config.DB.Raw(`
		SELECT
			s.roll_no,
			s.name,
			s.course_name,
			s.semester,
			COALESCE(SUM(c.credits), 0) as points
		FROM students s
		LEFT JOIN certificates c ON c.student_roll_no = s.roll_no
			AND c.status = 'Approved'
			AND c.activity_date >= ?
			AND c.activity_date <= ?
		GROUP BY s.roll_no, s.name, s.course_name, s.semester
		ORDER BY points DESC, s.name ASC
	`, startDate, endDate).Scan(&entries).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch leaderboard",
		})
	}

	// Mark the authenticated student
	for i := range entries {
		if entries[i].RollNo == rollNo {
			entries[i].IsSelf = true
		}
	}

	return c.JSON(entries)
}

// GetCategoryChampions returns the top student per activity category for a given academic year
func GetCategoryChampions(c *fiber.Ctx) error {
	year := c.Query("year")

	var startDate, endDate time.Time
	var err error
	if year != "" {
		startDate, endDate, err = getAcademicYearDateRange(year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid year format. Expected YYYY-YY (e.g., 2025-26)",
			})
		}
	} else {
		startDate, endDate = getCurrentAcademicYearRange()
	}

	type ChampionResult struct {
		Track   string `json:"track"`
		RollNo  string `json:"roll_no"`
		Name    string `json:"name"`
		Credits int    `json:"credits"`
	}

	var champions []ChampionResult

	err = config.DB.Raw(`
		WITH CategoryCredits AS (
			SELECT
				UPPER(c.activity_category) AS track,
				s.roll_no,
				s.name,
				SUM(c.credits) AS total_credits
			FROM students s
			JOIN certificates c ON c.student_roll_no = s.roll_no
			WHERE c.status = 'Approved'
			  AND c.activity_date >= ?
			  AND c.activity_date <= ?
			GROUP BY UPPER(c.activity_category), s.roll_no, s.name
		),
		RankedCategoryCredits AS (
			SELECT
				track,
				roll_no,
				name,
				total_credits,
				ROW_NUMBER() OVER (PARTITION BY track ORDER BY total_credits DESC, name ASC) as rn
			FROM CategoryCredits
		)
		SELECT track, roll_no, name, total_credits as credits
		FROM RankedCategoryCredits
		WHERE rn = 1
	`, startDate, endDate).Scan(&champions).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch category champions",
		})
	}

	return c.JSON(champions)
}

// GetActivities returns a list of activities
func GetActivities(c *fiber.Ctx) error {
	category := c.Query("category")
	status := c.Query("status")
	search := c.Query("search")

	query := config.DB.Model(&models.Activity{})
	if category != "" {
		query = query.Where("UPPER(category) = UPPER(?)", category)
	}
	if status != "" {
		query = query.Where("UPPER(status) = UPPER(?)", status)
	}
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR coordinator ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var activities []models.Activity
	if err := query.Order("activity_date asc").Find(&activities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch activities",
		})
	}

	resolveActivitiesCoordinators(activities)

	return c.JSON(activities)
}

// UpdateProfile updates student's editable profile info
func UpdateProfile(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	type ProfileUpdate struct {
		EmailID   string `json:"email_id"`
		ContactNo string `json:"contact_no"`
		DOB       string `json:"dob"`
		Gender    string `json:"gender"`
	}

	var input ProfileUpdate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	var student models.Student
	if err := config.DB.Where("roll_no = ?", rollNo).First(&student).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Student not found",
		})
	}

	// Update fields if provided
	if input.EmailID != "" {
		input.EmailID = utils.NormalizeEmail(input.EmailID)
		if !utils.ValidateEmail(input.EmailID) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid email address format",
			})
		}
		if input.EmailID != student.EmailID {
			var count int64
			config.DB.Model(&models.Student{}).Where("LOWER(email_id) = ? AND roll_no != ?", input.EmailID, student.RollNo).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "A student with this email already exists"})
			}
			config.DB.Model(&models.Admin{}).Where("LOWER(email) = ?", input.EmailID).Count(&count)
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An admin with this email already exists"})
			}
		}
		student.EmailID = input.EmailID
	}
	if input.ContactNo != "" {
		student.ContactNo = input.ContactNo
	}
	if input.DOB != "" {
		student.DOB = input.DOB
	}
	if input.Gender != "" {
		student.Gender = input.Gender
	}

	if err := config.DB.Save(&student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"student": fiber.Map{
			"roll_no":       student.RollNo,
			"name":          student.Name,
			"email_id":      student.EmailID,
			"course_name":   student.CourseName,
			"semester":      student.Semester,
			"contact_no":    student.ContactNo,
			"dob":           student.DOB,
			"gender":        student.Gender,
			"enrollment_no": student.EnrollmentNo,
			"is_verified":   student.IsVerified,
		},
	})
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword changes the authenticated student's password
func ChangePassword(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var input ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON input",
		})
	}

	if len(input.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters long",
		})
	}

	var student models.Student
	if err := config.DB.Where("roll_no = ?", rollNo).First(&student).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Student not found",
		})
	}

	// Verify current password
	if !utils.CheckPasswordHash(input.CurrentPassword, student.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect current password",
		})
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	student.Password = hashedPassword
	if err := config.DB.Save(&student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// EnrollActivity enrolls a student in an activity
func EnrollActivity(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)
	activityID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid activity ID",
		})
	}

	// Check if activity exists
	var activity models.Activity
	if err := config.DB.First(&activity, activityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Activity not found",
		})
	}

	// Check if already enrolled
	var existing models.Enrollment
	if err := config.DB.Where("student_roll_no = ? AND activity_id = ?", rollNo, activityID).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "You are already enrolled in this activity",
		})
	}

	enrollment := models.Enrollment{
		StudentRollNo: rollNo,
		ActivityID:    uint(activityID),
		Status:        "Enrolled",
	}

	// The enrollment and the student's notification are written in one
	// transaction. A student is entitled to be told that their enrolment was
	// registered, so an enrolment whose notification cannot be delivered must
	// not be recorded either: the rollback leaves the student unenrolled, the
	// endpoint reports failure rather than success, and the same request can be
	// retried once notification storage recovers.
	//
	// Without the rollback the enrolment would commit and the notification
	// would be lost for good, because the retry hits the already-enrolled
	// conflict above and never reaches the notification write again.
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&enrollment).Error; err != nil {
			return err
		}
		return createNotificationTx(
			tx,
			rollNo,
			"Activity Enrollment",
			fmt.Sprintf("You have successfully enrolled in %q.", activity.Name),
			models.NotificationTypeEnrollment,
		)
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to enroll in activity and notify you, please retry",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Enrolled successfully",
		"enrollment": enrollment,
	})
}

// GetEnrollments returns enrollments for the student
func GetEnrollments(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var enrollments []models.Enrollment
	if err := config.DB.Preload("Activity").Where("student_roll_no = ?", rollNo).Order("created_at desc").Find(&enrollments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch enrollments",
		})
	}

	// Resolve coordinator names dynamically for preloaded activities
	var admins []models.Admin
	if err := config.DB.Find(&admins).Error; err == nil {
		adminMap := make(map[string]string)
		for _, admin := range admins {
			adminMap[admin.AdminID] = admin.Name
		}
		for i := range enrollments {
			if enrollments[i].Activity.CoordinatorID != "" {
				if name, ok := adminMap[enrollments[i].Activity.CoordinatorID]; ok {
					enrollments[i].Activity.Coordinator = name
				}
			}
		}
	}

	return c.JSON(enrollments)
}

// GetDashboardStats returns stats for the student dashboard home page
func GetDashboardStats(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	// 1. Total activities participated (count of enrollments)
	var activitiesCount int64
	if err := config.DB.Model(&models.Enrollment{}).Where("student_roll_no = ? AND status IN ('Enrolled', 'Completed', 'Registered', 'Verified', 'Pending Verification')", rollNo).Count(&activitiesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate activities count",
		})
	}

	// 2. Total certificates uploaded and pending/approved/rejected
	var certificatesCount int64
	if err := config.DB.Model(&models.Certificate{}).Where("student_roll_no = ?", rollNo).Count(&certificatesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate certificates count",
		})
	}

	var pendingCertificatesCount int64
	if err := config.DB.Model(&models.Certificate{}).Where("student_roll_no = ? AND status = 'Pending'", rollNo).Count(&pendingCertificatesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate pending certificates count",
		})
	}

	var approvedCertificatesCount int64
	if err := config.DB.Model(&models.Certificate{}).Where("student_roll_no = ? AND status = 'Approved'", rollNo).Count(&approvedCertificatesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate approved certificates count",
		})
	}

	var rejectedCertificatesCount int64
	if err := config.DB.Model(&models.Certificate{}).Where("student_roll_no = ? AND status = 'Rejected'", rollNo).Count(&rejectedCertificatesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate rejected certificates count",
		})
	}

	// 3. Credits earned from approved certificates
	type SumResult struct {
		Total int
	}
	var sumResult SumResult
	if err := config.DB.Raw("SELECT COALESCE(SUM(credits), 0) as total FROM certificates WHERE student_roll_no = ? AND status = 'Approved'", rollNo).Scan(&sumResult).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate total credits",
		})
	}

	// 4. Current Rank on the Leaderboard
	type StudentRank struct {
		RollNo       string
		TotalCredits int
	}
	var ranks []StudentRank
	if err := config.DB.Raw(`
		SELECT s.roll_no, COALESCE(SUM(c.credits), 0) as total_credits
		FROM students s
		LEFT JOIN certificates c ON c.student_roll_no = s.roll_no AND c.status = 'Approved'
		GROUP BY s.roll_no
		ORDER BY total_credits DESC, s.roll_no ASC
	`).Scan(&ranks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate student ranks",
		})
	}

	rank := len(ranks)
	totalStudents := len(ranks)
	for i, r := range ranks {
		if r.RollNo == rollNo {
			rank = i + 1
			break
		}
	}

	// 5. Recent extracurricular activities list (approved/pending certificates)
	var recentActivities []models.Certificate
	if err := config.DB.Where("student_roll_no = ?", rollNo).Order("created_at desc").Limit(5).Find(&recentActivities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch recent activities",
		})
	}
	return c.JSON(fiber.Map{
		"activities_participated": activitiesCount,
		"certificates_uploaded":   certificatesCount,
		"pending_certificates":    pendingCertificatesCount,
		"approved_certificates":   approvedCertificatesCount,
		"rejected_certificates":   rejectedCertificatesCount,
		"credits_earned":          sumResult.Total,
		"current_rank":            rank,
		"total_students":          totalStudents,
		"recent_activities":       recentActivities,
	})
}

func resolveActivitiesCoordinators(activities []models.Activity) {
	if len(activities) == 0 {
		return
	}
	var admins []models.Admin
	if err := config.DB.Find(&admins).Error; err != nil {
		return
	}
	adminMap := make(map[string]string)
	for _, admin := range admins {
		adminMap[admin.AdminID] = admin.Name
	}

	for i := range activities {
		if activities[i].CoordinatorID != "" {
			if name, ok := adminMap[activities[i].CoordinatorID]; ok {
				activities[i].Coordinator = name
			}
		}
	}
}

// StudentInfoResponse represents student details for the marksheet
type StudentInfoResponse struct {
	Name         string `json:"name"`
	RollNo       string `json:"rollNo"`
	EnrollmentNo string `json:"enrollmentNo"`
	Semester     string `json:"semester"`
	Course       string `json:"course"`
	Department   string `json:"department"`
	Batch        string `json:"batch"`
	Institute    string `json:"institute"`
	AcademicYear string `json:"academicYear"`
}

// CreditCategoryResponse represents category-wise aggregated credits
type CreditCategoryResponse struct {
	Category     string `json:"category"`
	Activities   int    `json:"activities"`
	Credits      int    `json:"credits"`
	Contribution string `json:"contribution"`
}

// SemesterSummaryEntry represents semester-wise credits and activities
type SemesterSummaryEntry struct {
	Semester        string `json:"semester"`
	Name            string `json:"name"`
	Credits         int    `json:"credits"`
	Activities      int    `json:"activities"`
	ActivitiesCount int    `json:"activitiesCount"`
	Cumulative      int    `json:"cumulative"`
	Grade           string `json:"grade"`
}

// MarksheetResponse represents the full payload of the marksheet
type MarksheetResponse struct {
	StudentInfo              StudentInfoResponse      `json:"student_info"`
	CreditCategories         []CreditCategoryResponse `json:"credit_categories"`
	TotalActivities          int                      `json:"total_activities"`
	TotalCredits             int                      `json:"total_credits"`
	TotalContribution        string                   `json:"total_contribution"`
	SemesterSummary          []SemesterSummaryEntry   `json:"semester_summary"`
	FinalGrade               string                   `json:"final_grade"`
	TargetCredits            int                      `json:"target_credits"`
	CreditsNeededToNextGrade int                      `json:"credits_needed_to_next_grade"`
	NextGrade                string                   `json:"next_grade"`
	GradeScales              []fiber.Map              `json:"grade_scales"`
	Insights                 []fiber.Map              `json:"insights"`
}

// GetMarksheet generates the extracurricular marksheet, credits, and progress data
func GetMarksheet(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	var student models.Student
	if err := config.DB.Where("roll_no = ?", rollNo).First(&student).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Student not found",
		})
	}

	// Fetch all approved certificates
	var approvedCerts []models.Certificate
	if err := config.DB.Where("student_roll_no = ? AND status = 'Approved'", rollNo).Find(&approvedCerts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch approved certificates",
		})
	}

	admissionYear := parseAdmissionYear(rollNo)

	// Format Student Info
	duration := 5
	courseLower := strings.ToLower(student.CourseName)

	// 2-Year Courses
	if strings.Contains(courseLower, "apr") ||
		(strings.Contains(courseLower, "mba") && (strings.Contains(courseLower, "2") || strings.Contains(courseLower, "entrepreneurship"))) ||
		strings.Contains(courseLower, "entrepreneurship") {
		duration = 2
	} else if strings.Contains(courseLower, "b.com") || strings.Contains(courseLower, "bcom") {
		// 3-Year Courses
		duration = 3
	} else if strings.Contains(courseLower, "mca") ||
		strings.Contains(courseLower, "m.tech") || strings.Contains(courseLower, "mtech") ||
		strings.Contains(courseLower, "computer science") ||
		strings.Contains(courseLower, "information technology") ||
		strings.Contains(courseLower, "tourism") ||
		(strings.Contains(courseLower, "mba") && (strings.Contains(courseLower, "5") || strings.Contains(courseLower, "ms"))) {
		// 5-Year Courses
		duration = 5
	}
	batchStr := fmt.Sprintf("%d – %d", admissionYear, admissionYear+duration)

	deptStr := "-"
	if strings.Contains(courseLower, "mca") {
		deptStr = "-"
	} else if strings.Contains(courseLower, "computer science") {
		deptStr = "Computer Science"
	} else if strings.Contains(courseLower, "information") || strings.Contains(courseLower, "it") {
		deptStr = "Information Technology"
	}

	// Calculate Academic Year
	now := time.Now()
	var startYear int
	if now.Month() >= time.July {
		startYear = now.Year()
	} else {
		startYear = now.Year() - 1
	}
	endYearPart := (startYear + 1) % 100
	academicYearStr := fmt.Sprintf("%d-%02d", startYear, endYearPart)

	studentInfo := StudentInfoResponse{
		Name:         student.Name,
		RollNo:       student.RollNo,
		EnrollmentNo: student.EnrollmentNo,
		Semester:     semesterName(student.Semester),
		Course:       student.CourseName,
		Department:   deptStr,
		Batch:        batchStr,
		Institute:    "IIPS, DAVV Indore",
		AcademicYear: academicYearStr,
	}

	// Category aggregates
	type CatGroup struct {
		Category string
		Count    int
		Credits  int
	}
	var groups []CatGroup
	if err := config.DB.Raw(`
		SELECT UPPER(activity_category) as category, COUNT(*) as count, SUM(credits) as credits
		FROM certificates
		WHERE student_roll_no = ? AND status = 'Approved'
		GROUP BY UPPER(activity_category)
	`, rollNo).Scan(&groups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to aggregate categories",
		})
	}

	standardCategories := []string{"TECHNICAL", "PUBLIC SPEAKING", "RESEARCH", "SOCIAL SERVICE", "SPORTS", "CULTURAL", "LEADERSHIP"}
	catMap := make(map[string]CatGroup)
	for _, sc := range standardCategories {
		catMap[sc] = CatGroup{Category: sc, Count: 0, Credits: 0}
	}
	for _, g := range groups {
		catMap[g.Category] = g
	}

	var creditCategories []CreditCategoryResponse
	totalActivities := 0
	totalCredits := 0

	for _, sc := range standardCategories {
		g := catMap[sc]
		creditCategories = append(creditCategories, CreditCategoryResponse{
			Category:   formatCategoryName(sc),
			Activities: g.Count,
			Credits:    g.Credits,
		})
		totalActivities += g.Count
		totalCredits += g.Credits
	}
	for _, g := range groups {
		found := false
		for _, sc := range standardCategories {
			if sc == g.Category {
				found = true
				break
			}
		}
		if !found {
			creditCategories = append(creditCategories, CreditCategoryResponse{
				Category:   formatCategoryName(g.Category),
				Activities: g.Count,
				Credits:    g.Credits,
			})
			totalActivities += g.Count
			totalCredits += g.Credits
		}
	}

	for i := range creditCategories {
		if totalCredits > 0 {
			pct := float64(creditCategories[i].Credits) / float64(totalCredits) * 100
			creditCategories[i].Contribution = fmt.Sprintf("%.0f%%", pct)
		} else {
			creditCategories[i].Contribution = "0%"
		}
	}

	// Semester records
	semCerts := make(map[int][]models.Certificate)
	for _, cert := range approvedCerts {
		semNum := getSemesterForActivity(cert.ActivityDate, admissionYear)
		semCerts[semNum] = append(semCerts[semNum], cert)
	}

	startSem := student.Semester
	for semNum := range semCerts {
		if semNum < startSem {
			startSem = semNum
		}
	}
	if startSem < 1 {
		startSem = 1
	}
	currentSemester := student.Semester
	if currentSemester < startSem {
		currentSemester = startSem
	}

	var semesterSummary []SemesterSummaryEntry
	cumulative := 0
	for sem := startSem; sem <= currentSemester; sem++ {
		certs := semCerts[sem]
		semCredits := 0
		semActCount := len(certs)
		for _, cert := range certs {
			semCredits += cert.Credits
		}
		cumulative += semCredits
		grade := getSemesterGrade(semCredits)

		semesterSummary = append(semesterSummary, SemesterSummaryEntry{
			Semester:        romanSemester(sem),
			Name:            fmt.Sprintf("Semester %d", sem),
			Credits:         semCredits,
			Activities:      semActCount,
			ActivitiesCount: semActCount,
			Cumulative:      cumulative,
			Grade:           grade,
		})
	}

	finalGrade, neededToNext, nextGrade, scales := getGradeAndNextInfo(totalCredits)
	insights := generateInsights(creditCategories, neededToNext, nextGrade)

	contributionStr := "100%"
	if totalCredits == 0 {
		contributionStr = "0%"
	}

	return c.JSON(MarksheetResponse{
		StudentInfo:              studentInfo,
		CreditCategories:         creditCategories,
		TotalActivities:          totalActivities,
		TotalCredits:             totalCredits,
		TotalContribution:        contributionStr,
		SemesterSummary:          semesterSummary,
		FinalGrade:               finalGrade,
		TargetCredits:            200,
		CreditsNeededToNextGrade: neededToNext,
		NextGrade:                nextGrade,
		GradeScales:              scales,
		Insights:                 insights,
	})
}

func parseAdmissionYear(rollNo string) int {
	idx := strings.Index(strings.ToUpper(rollNo), "2K")
	if idx != -1 && len(rollNo) >= idx+4 {
		yearPart := rollNo[idx+2 : idx+4]
		if yr, err := strconv.Atoi(yearPart); err == nil {
			return 2000 + yr
		}
	}
	for i := 0; i <= len(rollNo)-4; i++ {
		if yr, err := strconv.Atoi(rollNo[i : i+4]); err == nil && yr >= 1990 && yr <= 2100 {
			return yr
		}
	}
	return time.Now().Year() - 3
}

func semesterName(sem int) string {
	mapping := map[int]string{
		1:  "I (First)",
		2:  "II (Second)",
		3:  "III (Third)",
		4:  "IV (Fourth)",
		5:  "V (Fifth)",
		6:  "VI (Sixth)",
		7:  "VII (Seventh)",
		8:  "VIII (Eighth)",
		9:  "IX (Ninth)",
		10: "X (Tenth)",
	}
	if name, ok := mapping[sem]; ok {
		return name
	}
	return strconv.Itoa(sem)
}

func romanSemester(sem int) string {
	roman := map[int]string{
		1:  "Semester I",
		2:  "Semester II",
		3:  "Semester III",
		4:  "Semester IV",
		5:  "Semester V",
		6:  "Semester VI",
		7:  "Semester VII",
		8:  "Semester VIII",
		9:  "Semester IX",
		10: "Semester X",
	}
	if name, ok := roman[sem]; ok {
		return name
	}
	return fmt.Sprintf("Semester %d", sem)
}

func formatCategoryName(category string) string {
	mapping := map[string]string{
		"TECHNICAL":       "Technical Skills",
		"PUBLIC SPEAKING": "Public Speaking",
		"RESEARCH":        "Research",
		"SOCIAL SERVICE":  "Social Service",
		"SPORTS":          "Sports",
		"LEADERSHIP":      "Leadership",
		"CULTURAL":        "Cultural Activities",
	}
	if val, ok := mapping[strings.ToUpper(category)]; ok {
		return val
	}
	lower := strings.ToLower(category)
	if len(lower) == 0 {
		return ""
	}
	runes := []rune(lower)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 'a' + 'A'
	}
	return string(runes)
}

func getSemesterForActivity(activityDate time.Time, admissionYear int) int {
	admissionStartDate := time.Date(admissionYear, time.July, 1, 0, 0, 0, 0, time.UTC)
	if activityDate.Before(admissionStartDate) {
		return 1
	}
	y1, m1, _ := admissionStartDate.Date()
	y2, m2, _ := activityDate.Date()
	totalMonths := int(y2-y1)*12 + int(m2-m1)
	if totalMonths < 0 {
		return 1
	}
	return (totalMonths / 6) + 1
}

func getSemesterGrade(credits int) string {
	if credits >= 30 {
		return "Grade O"
	} else if credits >= 20 {
		return "Grade A"
	} else if credits >= 15 {
		return "Grade B"
	} else if credits >= 10 {
		return "Grade C"
	}
	return "Grade D"
}

func getGradeAndNextInfo(totalCredits int) (string, int, string, []fiber.Map) {
	var currentGrade string
	var nextGrade string
	var neededToNext int

	scales := []fiber.Map{
		{"name": "Grade O", "range": "140+", "isActive": false},
		{"name": "Grade A", "range": "100-139", "isActive": false},
		{"name": "Grade B", "range": "70-99", "isActive": false},
		{"name": "Grade C", "range": "40-69", "isActive": false},
		{"name": "Grade D", "range": "Below 40", "isActive": false},
	}

	switch {
	case totalCredits >= 140:
		currentGrade = "Grade O"
		nextGrade = ""
		neededToNext = 0
		scales[0]["isActive"] = true
	case totalCredits >= 100:
		currentGrade = "Grade A"
		nextGrade = "Grade O"
		neededToNext = 140 - totalCredits
		scales[1]["isActive"] = true
	case totalCredits >= 70:
		currentGrade = "Grade B"
		nextGrade = "Grade A"
		neededToNext = 100 - totalCredits
		scales[2]["isActive"] = true
	case totalCredits >= 40:
		currentGrade = "Grade C"
		nextGrade = "Grade B"
		neededToNext = 70 - totalCredits
		scales[3]["isActive"] = true
	default:
		currentGrade = "Grade D"
		nextGrade = "Grade C"
		neededToNext = 40 - totalCredits
		scales[4]["isActive"] = true
	}

	return currentGrade, neededToNext, nextGrade, scales
}

func generateInsights(creditCategories []CreditCategoryResponse, neededToNext int, nextGrade string) []fiber.Map {
	var insights []fiber.Map

	var maxCat string
	var maxCredits int
	var minCat string
	var minCredits = 999999

	for _, cat := range creditCategories {
		if cat.Credits > maxCredits {
			maxCredits = cat.Credits
			maxCat = cat.Category
		}
		if cat.Credits < minCredits {
			minCredits = cat.Credits
			minCat = cat.Category
		}
	}

	if maxCat != "" && maxCredits > 0 {
		insights = append(insights, fiber.Map{
			"text": fmt.Sprintf("Your strongest contribution area is %s.", maxCat),
			"type": "success",
		})
	}

	if neededToNext > 0 && nextGrade != "" {
		insights = append(insights, fiber.Map{
			"text": fmt.Sprintf("You need %d more credits to reach %s.", neededToNext, nextGrade),
			"type": "danger",
		})
	} else if nextGrade == "" {
		insights = append(insights, fiber.Map{
			"text": "Congratulations! You have reached the highest Grade (O).",
			"type": "success",
		})
	}

	if minCat != "" && minCredits < maxCredits {
		insights = append(insights, fiber.Map{
			"text": fmt.Sprintf("%s activities have the lowest contribution.", minCat),
			"type": "warning",
		})
	}

	insights = append(insights, fiber.Map{
		"text": "Participating in upcoming hackathons or events may improve your score faster.",
		"type": "info",
	})

	return insights
}
