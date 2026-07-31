package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/controllers"
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/routes"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
)

func setupActivityMonitoringApp(t *testing.T) (*fiber.App, string, uint) {
	t.Helper()

	SetupTestDB(t)

	// AutoMigrate all models
	err := config.DB.AutoMigrate(
		&models.Admin{},
		&models.Student{},
		&models.Activity{},
		&models.Enrollment{},
		&models.Certificate{},
		&models.AdminNote{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test DB: %v", err)
	}

	// Clear test tables safely
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.AdminNote{})
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Certificate{})
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Enrollment{})
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Activity{})
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Student{})
	config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Admin{})

	// Create test admin
	hashedPwd, _ := utils.HashPassword("AdminPass123!")
	admin := models.Admin{
		AdminID:  "ADM001",
		Name:     "Test Super Admin",
		Email:    "admin@test.com",
		Password: hashedPwd,
		Role:     "superadmin",
		Status:   "Active",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to create test admin: %v", err)
	}

	token, err := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Seed test student
	student := models.Student{
		RollNo:       "21CS0071",
		Name:         "Dev Mehta",
		CourseName:   "MCA 5yrs Integrated",
		Semester:     6,
		ContactNo:    "9876543210",
		EmailID:      "dev@test.com",
		EnrollmentNo: "21CS0071",
		IsVerified:   true,
		Status:       "Approved",
	}
	config.DB.Create(&student)

	// Seed test activities
	act1 := models.Activity{
		Name:         "Leadership Workshop",
		Category:     "Soft Skills",
		Credits:      10,
		Mode:         "Offline",
		Status:       "Completed",
		RegDeadline:  time.Now().AddDate(0, 0, -10),
		ActivityDate: time.Now().AddDate(0, 0, -5),
		Coordinator:  "Dr. Rajesh Kumar",
	}
	config.DB.Create(&act1)

	act2 := models.Activity{
		Name:         "Research Project",
		Category:     "Academic",
		Credits:      20,
		Mode:         "Hybrid",
		Status:       "Ongoing",
		RegDeadline:  time.Now(),
		ActivityDate: time.Now().AddDate(0, 1, 0),
		Coordinator:  "Dr. Anita Sharma",
	}
	config.DB.Create(&act2)

	// Seed enrollment
	config.DB.Create(&models.Enrollment{
		StudentRollNo: student.RollNo,
		ActivityID:    act1.ID,
		Status:        "Completed",
	})

	// Seed certificate
	config.DB.Create(&models.Certificate{
		StudentRollNo:    student.RollNo,
		ActivityName:     act1.Name,
		ActivityCategory: act1.Category,
		Credits:          10,
		Status:           "Pending",
	})

	app := fiber.New()

	// Auth Middleware helper for tests
	authMW := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token header"})
		}
		claims, err := utils.ValidateAccessToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
		c.Locals("roll_no", claims.RollNo)
		c.Locals("user_id", claims.RollNo)
		c.Locals("admin_id", claims.RollNo)
		c.Locals("role", claims.Role)
		return c.Next()
	}

	adminGroup := app.Group("/api/admin", authMW)
	adminGroup.Get("/monitoring/stats", controllers.GetActivityMonitoringStats)
	adminGroup.Get("/monitoring/activities", controllers.GetMonitoredActivities)
	adminGroup.Put("/monitoring/activities/:id", controllers.UpdateMonitoredActivity)
	adminGroup.Get("/monitoring/insights", controllers.GetMonitoringInsights)
	adminGroup.Get("/monitoring/attention-students", controllers.GetStudentsRequiringAttention)
	adminGroup.Post("/monitoring/send-reminder", controllers.SendActivityMonitoringReminder)

	return app, token, act1.ID
}

func TestGetActivityMonitoringStats(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if body["total_monitored"] == nil {
		t.Errorf("Expected total_monitored in response")
	}
}

func TestGetMonitoredActivities(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/activities?search=Leadership&sort_by=credits-desc", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var activities []controllers.MonitoredActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(activities) == 0 {
		t.Errorf("Expected at least one activity returned")
	}
	if activities[0].Name != "Leadership Workshop" {
		t.Errorf("Expected Leadership Workshop, got %s", activities[0].Name)
	}
}

func TestUpdateMonitoredActivity(t *testing.T) {
	app, token, actID := setupActivityMonitoringApp(t)

	newCredits := 15
	payload := controllers.UpdateMonitoredActivityInput{
		Name:        "Updated Leadership Workshop",
		Category:    "Soft Skills",
		Credits:     &newCredits,
		Status:      "Ongoing",
		StartDate:   "01 Jan 2025",
		EndDate:     "15 Jan 2025",
		Coordinator: "Dr. New Coordinator",
	}

	jsonBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/monitoring/activities/%d", actID), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test non-existent activity ID
	req404 := httptest.NewRequest("PUT", "/api/admin/monitoring/activities/9999", bytes.NewBuffer(jsonBytes))
	req404.Header.Set("Content-Type", "application/json")
	req404.Header.Set("Authorization", "Bearer "+token)
	resp404, _ := app.Test(req404)
	if resp404.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent ID, got %d", resp404.StatusCode)
	}
}

func TestGetMonitoringInsights(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/insights", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetStudentsRequiringAttention(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/attention-students", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var attentionList []controllers.AttentionStudentResponse
	if err := json.NewDecoder(resp.Body).Decode(&attentionList); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(attentionList) == 0 {
		t.Errorf("Expected attention students returned")
	}
}

func TestSendActivityMonitoringReminder(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	payload := controllers.ActivityReminderInput{
		StudentEnrollment: "21CS0071",
		ActivityName:      "Technical Symposium",
		Issue:             "Certificate Not Uploaded",
		Message:           "Please upload your certificate as soon as possible.",
	}

	jsonBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/admin/monitoring/send-reminder", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test missing student identifier
	badPayload := controllers.ActivityReminderInput{
		ActivityName: "Some Activity",
	}
	badBytes, _ := json.Marshal(badPayload)
	reqBad := httptest.NewRequest("POST", "/api/admin/monitoring/send-reminder", bytes.NewBuffer(badBytes))
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.Header.Set("Authorization", "Bearer "+token)

	respBad, _ := app.Test(reqBad)
	if respBad.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing student roll/enrollment, got %d", respBad.StatusCode)
	}
}

func TestAssignedBatchIsolation_AttentionStudentsAndReminder(t *testing.T) {
	// Seed the fixtures, then discard the hand-wired app in favour of the real
	// route table. This test asserts that a reminder reaches the student, so it
	// has to exercise the same handlers and middleware the application serves:
	// wiring the routes by hand here previously let it pass against a handler
	// the portal no longer calls.
	setupActivityMonitoringApp(t)

	app := fiber.New()
	routes.SetupRoutes(app)

	// Create scoped Admin A for IT2K24
	hashedPwd, _ := utils.HashPassword("AdminPass123!")
	admin24 := models.Admin{
		AdminID:       "ADM24",
		Name:          "Mentor 24",
		Email:         "m24@test.com",
		Password:      hashedPwd,
		Role:          "admin",
		AssignedBatch: "IT2K24",
		Status:        "Active",
	}
	config.DB.Create(&admin24)
	token24, _ := utils.GenerateAccessToken(admin24.AdminID, admin24.Email, admin24.Role)

	// Create scoped Admin B for IT2K25
	admin25 := models.Admin{
		AdminID:       "ADM25",
		Name:          "Mentor 25",
		Email:         "m25@test.com",
		Password:      hashedPwd,
		Role:          "admin",
		AssignedBatch: "IT2K25",
		Status:        "Active",
	}
	config.DB.Create(&admin25)

	// Seed student in IT2K24
	student24 := models.Student{
		RollNo:       "IT2K24001",
		Name:         "Aarav Sharma",
		CourseName:   "MCA 5yrs Integrated",
		Semester:     4,
		ContactNo:    "9876543211",
		EmailID:      "aarav24@test.com",
		EnrollmentNo: "EN-IT2K24001",
		IsVerified:   true,
		Status:       "Approved",
	}
	config.DB.Create(&student24)

	// Seed student in IT2K25
	student25 := models.Student{
		RollNo:       "IT2K25001",
		Name:         "Bhavya Patel",
		CourseName:   "MCA 5yrs Integrated",
		Semester:     2,
		ContactNo:    "9876543212",
		EmailID:      "bhavya25@test.com",
		EnrollmentNo: "EN-IT2K25001",
		IsVerified:   true,
		Status:       "Approved",
	}
	config.DB.Create(&student25)

	// Seed pending certificates for both students
	config.DB.Create(&models.Certificate{
		StudentRollNo:    student24.RollNo,
		ActivityName:     "Python Workshop",
		ActivityCategory: "Technical",
		Credits:          10,
		Status:           "Pending",
	})
	config.DB.Create(&models.Certificate{
		StudentRollNo:    student25.RollNo,
		ActivityName:     "Cloud Summit",
		ActivityCategory: "Technical",
		Credits:          10,
		Status:           "Pending",
	})

	// Admin 24 queries attention students
	reqAtt := httptest.NewRequest("GET", "/api/admin/monitoring/attention-students", nil)
	reqAtt.Header.Set("Authorization", "Bearer "+token24)

	respAtt, err := app.Test(reqAtt)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	if respAtt.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", respAtt.StatusCode)
	}

	var attList []controllers.AttentionStudentResponse
	if err := json.NewDecoder(respAtt.Body).Decode(&attList); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify only IT2K24 student is present for Admin 24
	for _, item := range attList {
		if item.Enrollment == "EN-IT2K25001" {
			t.Errorf("Assigned batch isolation broken: Admin 24 received IT2K25 student EN-IT2K25001")
		}
	}

	// Admin 24 attempts to send reminder to IT2K25 student -> must be rejected (403 Forbidden)
	payloadCross := controllers.ActivityReminderInput{
		StudentEnrollment: "EN-IT2K25001",
		ActivityName:      "Cloud Summit",
		Issue:             "Pending Verification",
		Message:           "Please complete your submission.",
	}
	jsonCross, _ := json.Marshal(payloadCross)
	reqCross := httptest.NewRequest("POST", "/api/admin/monitoring/send-reminder", bytes.NewBuffer(jsonCross))
	reqCross.Header.Set("Content-Type", "application/json")
	reqCross.Header.Set("Authorization", "Bearer "+token24)

	respCross, _ := app.Test(reqCross)
	if respCross.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for cross-batch reminder, got %d", respCross.StatusCode)
	}

	// Admin 24 sends reminder to IT2K24 student -> must succeed (200 OK)
	payloadValid := controllers.ActivityReminderInput{
		StudentEnrollment: "EN-IT2K24001",
		ActivityName:      "Python Workshop",
		Issue:             "Pending Verification",
	}
	jsonValid, _ := json.Marshal(payloadValid)
	reqValid := httptest.NewRequest("POST", "/api/admin/monitoring/send-reminder", bytes.NewBuffer(jsonValid))
	reqValid.Header.Set("Content-Type", "application/json")
	reqValid.Header.Set("Authorization", "Bearer "+token24)

	respValid, _ := app.Test(reqValid)
	if respValid.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for valid batch reminder, got %d", respValid.StatusCode)
	}

	// Verify the reminder through the notification endpoint the portal's bell
	// actually reads, in that endpoint's response format.
	tokenStudent24, _ := utils.GenerateAccessToken(student24.RollNo, student24.EmailID, "student")
	reqNotif := httptest.NewRequest("GET", "/api/student/notifications", nil)
	reqNotif.Header.Set("Authorization", "Bearer "+tokenStudent24)

	respNotif, err := app.Test(reqNotif)
	if err != nil {
		t.Fatalf("Failed to execute student notifications request: %v", err)
	}
	if respNotif.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for student notifications, got %d", respNotif.StatusCode)
	}

	var notifBody struct {
		Notifications []models.Notification `json:"notifications"`
		UnreadCount   int                   `json:"unread_count"`
		Total         int                   `json:"total"`
	}
	if err := json.NewDecoder(respNotif.Body).Decode(&notifBody); err != nil {
		t.Fatalf("Failed to decode student notifications response: %v", err)
	}

	if len(notifBody.Notifications) != 1 {
		t.Fatalf("Expected exactly 1 reminder notification for student 24, got %d", len(notifBody.Notifications))
	}
	reminder := notifBody.Notifications[0]
	if !strings.Contains(reminder.Message, "Python Workshop") {
		t.Errorf("Expected reminder notification to contain activity name, got: %s", reminder.Message)
	}
	if reminder.Type != models.NotificationTypeActivity {
		t.Errorf("Expected reminder notification type %q, got %q", models.NotificationTypeActivity, reminder.Type)
	}
	if reminder.IsRead {
		t.Errorf("Expected a freshly sent reminder to be unread")
	}
	if notifBody.UnreadCount != 1 {
		t.Errorf("Expected unread_count 1 for the new reminder, got %d", notifBody.UnreadCount)
	}
	// The internal note-trail marker must not leak into what the student reads.
	if strings.Contains(reminder.Message, "[ALERT REMINDER]") {
		t.Errorf("Expected the internal marker to be stripped from the student message, got: %s", reminder.Message)
	}

	// The cross-batch reminder rejected earlier must not have notified anyone.
	tokenStudent25, _ := utils.GenerateAccessToken(student25.RollNo, student25.EmailID, "student")
	reqNotif25 := httptest.NewRequest("GET", "/api/student/notifications", nil)
	reqNotif25.Header.Set("Authorization", "Bearer "+tokenStudent25)

	respNotif25, err := app.Test(reqNotif25)
	if err != nil {
		t.Fatalf("Failed to execute student 25 notifications request: %v", err)
	}
	var notifBody25 struct {
		Notifications []models.Notification `json:"notifications"`
	}
	if err := json.NewDecoder(respNotif25.Body).Decode(&notifBody25); err != nil {
		t.Fatalf("Failed to decode student 25 notifications response: %v", err)
	}
	if len(notifBody25.Notifications) != 0 {
		t.Errorf("Expected the rejected cross-batch reminder to notify nobody, got %d notification(s)", len(notifBody25.Notifications))
	}
}

func TestDatabaseDerivedInsights(t *testing.T) {
	app, token, _ := setupActivityMonitoringApp(t)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/insights", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode insights response body: %v", err)
	}

	// Verify students_requiring_followup count matches exact pending certificates / attention count (1), NOT 1 + 3 = 4
	followup, ok := body["students_requiring_followup"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected students_requiring_followup object in response")
	}
	followupCount := int(followup["count"].(float64))
	if followupCount != 1 {
		t.Errorf("Expected database-derived followup count 1, got %d (check for hardcoded offsets)", followupCount)
	}

	belowTarget := int(body["below_credit_target"].(float64))
	if belowTarget < 0 {
		t.Errorf("Invalid below_credit_target value: %d", belowTarget)
	}
}
