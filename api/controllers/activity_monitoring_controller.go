package controllers

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"gorm.io/gorm"
)

// MonitoredActivityResponse represents the DTO returned to the frontend for Activity Monitoring.
type MonitoredActivityResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Credits     int    `json:"credits"`
	Registered  int    `json:"registered"`
	Completed   int    `json:"completed"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Coordinator string `json:"coordinator,omitempty"`
}

// UpdateMonitoredActivityInput represents the payload to update a monitored activity.
type UpdateMonitoredActivityInput struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Credits     *int   `json:"credits"`
	Registered  *int   `json:"registered"`
	Completed   *int   `json:"completed"`
	Status      string `json:"status"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Coordinator string `json:"coordinator"`
}

// AttentionStudentResponse represents a student requiring admin attention.
type AttentionStudentResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Initials      string `json:"initials"`
	Enrollment    string `json:"enrollment"`
	Activity      string `json:"activity"`
	Issue         string `json:"issue"`
	StatusColor   string `json:"statusColor"`
	Days          int    `json:"days"`
	CourseName    string `json:"courseName,omitempty"`
	CreditsEarned int    `json:"creditsEarned,omitempty"`
}

// ActivityReminderInput represents payload for sending a reminder.
type ActivityReminderInput struct {
	StudentEnrollment string `json:"student_enrollment"`
	StudentRollNo     string `json:"student_roll_no"`
	ActivityName      string `json:"activity_name"`
	Issue             string `json:"issue"`
	Message           string `json:"message"`
}

// Helper functions for batch-scoped access
func isBatchScopedAdmin(admin *models.Admin) bool {
	return admin.Role == "admin"
}

func scopeStudentQuery(query *gorm.DB, admin *models.Admin) *gorm.DB {
	if isBatchScopedAdmin(admin) {
		if admin.AssignedBatch == "" {
			return query.Where("1 = 0")
		}
		return query.Where("roll_no LIKE ? OR enrollment_no LIKE ?", admin.AssignedBatch+"%", "%"+admin.AssignedBatch+"%")
	}
	return query
}

func scopeCertOrEnrollmentQuery(query *gorm.DB, admin *models.Admin) *gorm.DB {
	if isBatchScopedAdmin(admin) {
		if admin.AssignedBatch == "" {
			return query.Where("1 = 0")
		}
		return query.Where("student_roll_no LIKE ?", admin.AssignedBatch+"%")
	}
	return query
}

func isStudentInAdminScope(student *models.Student, admin *models.Admin) bool {
	if !isBatchScopedAdmin(admin) {
		return true
	}
	if admin.AssignedBatch == "" {
		return false
	}
	return strings.HasPrefix(student.RollNo, admin.AssignedBatch) || strings.Contains(student.EnrollmentNo, admin.AssignedBatch)
}

// GetActivityMonitoringStats handles GET /api/admin/monitoring/stats
func GetActivityMonitoringStats(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var totalMonitored int64
	if err := config.DB.Model(&models.Activity{}).Count(&totalMonitored).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to count activities")
	}

	var ongoingCount int64
	if err := config.DB.Model(&models.Activity{}).
		Where("LOWER(status) IN (?, ?, ?)", "ongoing", "open", "closing soon").
		Count(&ongoingCount).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to count ongoing activities")
	}

	var completedCount int64
	if err := config.DB.Model(&models.Activity{}).
		Where("LOWER(status) IN (?, ?)", "completed", "closed").
		Count(&completedCount).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to count completed activities")
	}

	var pendingCertificates int64
	pendingCertQuery := scopeCertOrEnrollmentQuery(config.DB.Model(&models.Certificate{}), admin).
		Where("LOWER(status) = ?", "pending")
	if err := pendingCertQuery.Count(&pendingCertificates).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to count pending certificates")
	}

	var pendingActivities int64
	if err := config.DB.Model(&models.Activity{}).
		Where("LOWER(status) = ?", "pending verification").
		Count(&pendingActivities).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to count pending activities")
	}

	totalPending := pendingActivities + pendingCertificates

	var recentActivities int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	config.DB.Model(&models.Activity{}).Where("created_at >= ?", sevenDaysAgo).Count(&recentActivities)

	totalMonitoredChange := fmt.Sprintf("+%d this week", recentActivities)
	if recentActivities == 0 {
		totalMonitoredChange = "Active registry"
	}

	var nearingDeadlineCount int64
	sevenDaysLater := time.Now().AddDate(0, 0, 7)
	config.DB.Model(&models.Activity{}).
		Where("LOWER(status) IN (?, ?, ?) AND reg_deadline >= ? AND reg_deadline <= ?", "ongoing", "open", "closing soon", time.Now().AddDate(0, 0, -1), sevenDaysLater).
		Count(&nearingDeadlineCount)

	nearingDeadlineStr := fmt.Sprintf("%d nearing deadline", nearingDeadlineCount)

	var recentCompleted int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentCompletedQuery := scopeCertOrEnrollmentQuery(config.DB.Model(&models.Certificate{}), admin).
		Where("LOWER(status) = ? AND updated_at >= ?", "approved", thirtyDaysAgo)
	recentCompletedQuery.Count(&recentCompleted)

	completedChange := fmt.Sprintf("+%d this month", recentCompleted)
	if recentCompleted == 0 {
		completedChange = "Total completed"
	}

	requiresReviewStr := "All verified"
	if totalPending > 0 {
		requiresReviewStr = fmt.Sprintf("%d pending review", totalPending)
	}

	return c.JSON(fiber.Map{
		"total_monitored":        totalMonitored,
		"total_monitored_change": totalMonitoredChange,
		"ongoing":                ongoingCount,
		"nearing_deadline":       nearingDeadlineStr,
		"completed":              completedCount,
		"completed_change":       completedChange,
		"pending_verification":   totalPending,
		"requires_review":        requiresReviewStr,
	})
}

// GetMonitoredActivities handles GET /api/admin/monitoring/activities
func GetMonitoredActivities(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	search := strings.TrimSpace(c.Query("search"))
	category := strings.TrimSpace(c.Query("category"))
	status := strings.TrimSpace(c.Query("status"))
	sortBy := strings.TrimSpace(c.Query("sort_by"))

	var activities []models.Activity
	query := config.DB.Model(&models.Activity{})

	if search != "" {
		lowerSearch := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(category) LIKE ?", lowerSearch, lowerSearch)
	}

	if category != "" {
		lowerCat := "%" + strings.ToLower(category) + "%"
		query = query.Where("LOWER(category) LIKE ? OR LOWER(name) LIKE ?", lowerCat, lowerCat)
	}

	if status != "" {
		switch strings.ToLower(status) {
		case "completed":
			query = query.Where("LOWER(status) IN (?, ?)", "completed", "closed")
		case "ongoing":
			query = query.Where("LOWER(status) IN (?, ?, ?)", "ongoing", "open", "closing soon")
		case "pending verification":
			query = query.Where("LOWER(status) IN (?, ?)", "pending verification", "pending")
		case "upcoming":
			query = query.Where("LOWER(status) = ?", "upcoming")
		default:
			query = query.Where("LOWER(status) LIKE ?", "%"+strings.ToLower(status)+"%")
		}
	}

	if sortBy != "" {
		switch sortBy {
		case "credits-desc":
			query = query.Order("credits DESC")
		case "credits-asc":
			query = query.Order("credits ASC")
		}
	}

	if err := query.Find(&activities).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to fetch activities")
	}

	// Fetch enrollment statistics for each activity
	type enrollmentStat struct {
		ActivityID uint
		Total      int
		Completed  int
	}

	var stats []enrollmentStat
	scopeCertOrEnrollmentQuery(config.DB.Model(&models.Enrollment{}), admin).
		Select("activity_id, COUNT(*) as total, SUM(CASE WHEN LOWER(status) = 'completed' THEN 1 ELSE 0 END) as completed").
		Group("activity_id").
		Scan(&stats)

	statMap := make(map[uint]enrollmentStat)
	for _, s := range stats {
		statMap[s.ActivityID] = s
	}

	// Also check certificate approvals per activity category/name for accurate completed metrics
	type certStat struct {
		ActivityName string
		Approved     int
	}
	var certStats []certStat
	scopeCertOrEnrollmentQuery(config.DB.Model(&models.Certificate{}), admin).
		Select("activity_name, COUNT(*) as approved").
		Where("LOWER(status) = ?", "approved").
		Group("activity_name").
		Scan(&certStats)

	certMap := make(map[string]int)
	for _, cs := range certStats {
		certMap[strings.ToLower(cs.ActivityName)] = cs.Approved
	}

	response := make([]MonitoredActivityResponse, 0, len(activities))
	for _, act := range activities {
		regCount := 0
		compCount := 0

		if s, ok := statMap[act.ID]; ok {
			regCount = s.Total
			compCount = s.Completed
		}

		// Add cert count if higher
		if approvedCerts, ok := certMap[strings.ToLower(act.Name)]; ok && approvedCerts > compCount {
			compCount = approvedCerts
			if compCount > regCount {
				regCount = compCount
			}
		}

		// Standardize date strings
		startStr := act.RegDeadline.Format("02 Jan 2006")
		if act.RegDeadline.IsZero() {
			startStr = act.CreatedAt.Format("02 Jan 2006")
		}
		endStr := act.ActivityDate.Format("02 Jan 2006")
		if act.ActivityDate.IsZero() {
			endStr = time.Now().AddDate(0, 1, 0).Format("02 Jan 2006")
		}

		// Normalize status capitalization for UI consistency
		dispStatus := act.Status
		switch strings.ToLower(act.Status) {
		case "completed", "closed":
			dispStatus = "Completed"
		case "ongoing", "open", "closing soon":
			dispStatus = "Ongoing"
		case "pending verification", "pending":
			dispStatus = "Pending Verification"
		case "upcoming":
			dispStatus = "Upcoming"
		}

		response = append(response, MonitoredActivityResponse{
			ID:          act.ID,
			Name:        act.Name,
			Category:    act.Category,
			Credits:     act.Credits,
			Registered:  regCount,
			Completed:   compCount,
			StartDate:   startStr,
			EndDate:     endStr,
			Status:      dispStatus,
			Description: act.Description,
			Coordinator: act.Coordinator,
		})
	}

	// Sort response based on sort_by parameter
	switch sortBy {
	case "credits-desc":
		sort.Slice(response, func(i, j int) bool { return response[i].Credits > response[j].Credits })
	case "credits-asc":
		sort.Slice(response, func(i, j int) bool { return response[i].Credits < response[j].Credits })
	case "registered-desc":
		sort.Slice(response, func(i, j int) bool { return response[i].Registered > response[j].Registered })
	case "completion-desc":
		sort.Slice(response, func(i, j int) bool {
			ratioI := 0.0
			if response[i].Registered > 0 {
				ratioI = float64(response[i].Completed) / float64(response[i].Registered)
			}
			ratioJ := 0.0
			if response[j].Registered > 0 {
				ratioJ = float64(response[j].Completed) / float64(response[j].Registered)
			}
			return ratioI > ratioJ
		})
	}

	return c.JSON(response)
}

// UpdateMonitoredActivity handles PUT /api/admin/monitoring/activities/:id
func UpdateMonitoredActivity(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Invalid activity ID")
	}

	var input UpdateMonitoredActivityInput
	if err := c.BodyParser(&input); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Invalid request body")
	}

	var activity models.Activity
	if err := config.DB.First(&activity, uint(id)).Error; err != nil {
		return errJSON(c, fiber.StatusNotFound, "Activity not found")
	}

	if strings.TrimSpace(input.Name) != "" {
		activity.Name = strings.TrimSpace(input.Name)
	}

	if strings.TrimSpace(input.Category) != "" {
		activity.Category = strings.TrimSpace(input.Category)
	}

	if input.Credits != nil {
		if *input.Credits < 0 {
			return errJSON(c, fiber.StatusBadRequest, "Credits cannot be negative")
		}
		activity.Credits = *input.Credits
	}

	if strings.TrimSpace(input.Status) != "" {
		activity.Status = strings.TrimSpace(input.Status)
	}

	if strings.TrimSpace(input.Coordinator) != "" {
		activity.Coordinator = strings.TrimSpace(input.Coordinator)
	}

	// Parse date fields if provided
	dateLayouts := []string{"02 Jan 2006", "2006-01-02", "02-01-2006", "Jan 02, 2006"}
	if strings.TrimSpace(input.StartDate) != "" {
		var parsedStart time.Time
		var parseErr error
		for _, layout := range dateLayouts {
			parsedStart, parseErr = time.Parse(layout, strings.TrimSpace(input.StartDate))
			if parseErr == nil {
				break
			}
		}
		if parseErr == nil {
			activity.RegDeadline = parsedStart
		}
	}

	if strings.TrimSpace(input.EndDate) != "" {
		var parsedEnd time.Time
		var parseErr error
		for _, layout := range dateLayouts {
			parsedEnd, parseErr = time.Parse(layout, strings.TrimSpace(input.EndDate))
			if parseErr == nil {
				break
			}
		}
		if parseErr == nil {
			activity.ActivityDate = parsedEnd
		}
	}

	if err := config.DB.Save(&activity).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to update activity")
	}

	return c.JSON(fiber.Map{
		"message":  "Activity updated successfully",
		"activity": activity,
	})
}

// GetMonitoringInsights handles GET /api/admin/monitoring/insights
func GetMonitoringInsights(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var activities []models.Activity
	if err := config.DB.Find(&activities).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to fetch activities")
	}

	mostParticipatedName := "N/A"
	mostParticipatedCount := 0

	highestCreditName := "N/A"
	highestCreditVal := 0

	highestCompletionName := "N/A"
	highestCompletionRate := 0

	for _, act := range activities {
		if act.Credits > highestCreditVal {
			highestCreditVal = act.Credits
			highestCreditName = act.Name
		}

		var regCount int64
		scopeCertOrEnrollmentQuery(config.DB.Model(&models.Enrollment{}), admin).
			Where("activity_id = ?", act.ID).
			Count(&regCount)

		var compCount int64
		scopeCertOrEnrollmentQuery(config.DB.Model(&models.Enrollment{}), admin).
			Where("activity_id = ? AND LOWER(status) = ?", act.ID, "completed").
			Count(&compCount)

		if int(regCount) > mostParticipatedCount {
			mostParticipatedCount = int(regCount)
			mostParticipatedName = act.Name
		}

		if regCount > 0 {
			rate := int(math.Round(float64(compCount) / float64(regCount) * 100))
			if rate > highestCompletionRate {
				highestCompletionRate = rate
				highestCompletionName = act.Name
			}
		}
	}

	var pendingCerts int64
	scopeCertOrEnrollmentQuery(config.DB.Model(&models.Certificate{}), admin).
		Where("LOWER(status) = ?", "pending").
		Count(&pendingCerts)

	var pendingActivities int64
	config.DB.Model(&models.Activity{}).
		Where("LOWER(status) = ?", "pending verification").
		Count(&pendingActivities)

	// Fetch students scoped to admin's assigned batch
	var students []models.Student
	scopeStudentQuery(config.DB.Preload("Certificates").Preload("Enrollments"), admin).Find(&students)

	belowCreditTarget := 0
	inactiveStudents := 0
	const targetCredits = 20

	for _, s := range students {
		earned := calculateStudentCredits(s)
		if earned < targetCredits {
			belowCreditTarget++
		}
		if len(s.Enrollments) == 0 && len(s.Certificates) == 0 {
			inactiveStudents++
		}
	}

	// Calculate attention list for students requiring followup
	attentionList := getAttentionListForStudents(students)
	followupStudentMap := make(map[string]bool)
	followupActivityMap := make(map[string]bool)
	for _, item := range attentionList {
		followupStudentMap[item.Enrollment] = true
		if item.Activity != "" {
			followupActivityMap[item.Activity] = true
		}
	}

	return c.JSON(fiber.Map{
		"most_participated": fiber.Map{
			"name":  mostParticipatedName,
			"count": mostParticipatedCount,
		},
		"highest_credit": fiber.Map{
			"name":    highestCreditName,
			"credits": highestCreditVal,
		},
		"highest_completion": fiber.Map{
			"name": highestCompletionName,
			"rate": highestCompletionRate,
		},
		"students_requiring_followup": fiber.Map{
			"count":          len(followupStudentMap),
			"activity_count": len(followupActivityMap),
		},
		"pending_certificates":      pendingCerts,
		"below_credit_target":       belowCreditTarget,
		"inactive_students":         inactiveStudents,
		"activities_pending_review": pendingActivities,
	})
}

// GetStudentsRequiringAttention handles GET /api/admin/monitoring/attention-students
func GetStudentsRequiringAttention(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var students []models.Student
	if err := scopeStudentQuery(config.DB.Preload("Certificates").Preload("Enrollments"), admin).Find(&students).Error; err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to fetch students requiring attention")
	}

	attentionList := getAttentionListForStudents(students)
	return c.JSON(attentionList)
}

func getAttentionListForStudents(students []models.Student) []AttentionStudentResponse {
	attentionList := make([]AttentionStudentResponse, 0)
	var itemID uint = 1

	for _, s := range students {
		// Check for pending certificates
		for _, cert := range s.Certificates {
			if strings.ToLower(cert.Status) == "pending" {
				days := int(time.Since(cert.CreatedAt).Hours() / 24)
				if days < 1 {
					days = 7
				}
				attentionList = append(attentionList, AttentionStudentResponse{
					ID:            itemID,
					Name:          s.Name,
					Initials:      getInitials(s.Name),
					Enrollment:    s.EnrollmentNo,
					Activity:      cert.ActivityName,
					Issue:         "Pending Verification",
					StatusColor:   "text-blue-600 font-bold",
					Days:          days,
					CourseName:    s.CourseName,
					CreditsEarned: calculateStudentCredits(s),
				})
				itemID++
			}
		}

		// Check for enrollments without certificates
		if len(s.Enrollments) > 0 && len(s.Certificates) == 0 {
			for _, enr := range s.Enrollments {
				days := int(time.Since(enr.CreatedAt).Hours() / 24)
				if days < 1 {
					days = 14
				}
				attentionList = append(attentionList, AttentionStudentResponse{
					ID:            itemID,
					Name:          s.Name,
					Initials:      getInitials(s.Name),
					Enrollment:    s.EnrollmentNo,
					Activity:      "Extracurricular Activity",
					Issue:         "Certificate Not Uploaded",
					StatusColor:   "text-rose-600 font-bold",
					Days:          days,
					CourseName:    s.CourseName,
					CreditsEarned: calculateStudentCredits(s),
				})
				itemID++
				break
			}
		}
	}
	return attentionList
}

// SendActivityMonitoringReminder handles POST /api/admin/monitoring/send-reminder
func SendActivityMonitoringReminder(c *fiber.Ctx) error {
	admin, err := getAuthenticatedAdmin(c)
	if err != nil {
		return err
	}

	var input ActivityReminderInput
	if err := c.BodyParser(&input); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "Cannot parse request body")
	}

	studentIdentifier := strings.TrimSpace(input.StudentEnrollment)
	if studentIdentifier == "" {
		studentIdentifier = strings.TrimSpace(input.StudentRollNo)
	}

	if studentIdentifier == "" {
		return errJSON(c, fiber.StatusBadRequest, "Student enrollment or roll number is required")
	}

	var student models.Student
	err = config.DB.Where("enrollment_no = ? OR roll_no = ?", studentIdentifier, studentIdentifier).First(&student).Error
	if err != nil {
		return errJSON(c, fiber.StatusNotFound, "Student not found with provided enrollment/roll number")
	}

	if !isStudentInAdminScope(&student, admin) {
		return errJSON(c, fiber.StatusForbidden, "Unauthorized: Student does not belong to your assigned batch")
	}

	msg := input.Message
	if msg == "" {
		activityName := input.ActivityName
		if activityName == "" {
			activityName = "Activity"
		}
		issue := input.Issue
		if issue == "" {
			issue = "Outstanding item requiring resolution"
		}
		msg = fmt.Sprintf("Reminder for %s: %s. Please update your submission in the portal.", activityName, issue)
	}

	note := models.AdminNote{
		StudentRollNo: student.RollNo,
		AdminID:       admin.AdminID,
		AuthorName:    admin.Name,
		Role:          admin.Role,
		Text:          models.AdminNoteReminderPrefix + msg,
	}

	// The admin note is the staff-side audit trail; the notification is what the
	// student actually sees in their bell. A reminder that reaches only the note
	// trail has not reached the student at all, so the two are written in one
	// transaction: if the notification cannot be stored the note is rolled back
	// too, the endpoint reports failure instead of claiming the reminder was
	// sent, and the admin can retry it.
	//
	// The notification carries the bare message: the "[ALERT REMINDER]" prefix
	// is an internal marker for the note trail, not something to show a student.
	if err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		return createNotificationTx(tx, student.RollNo, models.NotificationTitleActivityReminder, msg, models.NotificationTypeActivity)
	}); err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "Failed to send reminder alert")
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Reminder alert sent to %s (%s) for %s.", student.Name, student.EnrollmentNo, input.ActivityName),
		"note_id": note.ID,
	})
}

func getInitials(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "ST"
	}
	if len(parts) == 1 {
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(parts[0][:1] + parts[len(parts)-1][:1])
}

func calculateStudentCredits(s models.Student) int {
	total := 0
	for _, cert := range s.Certificates {
		if strings.ToLower(cert.Status) == "approved" {
			total += cert.Credits
		}
	}
	return total
}
