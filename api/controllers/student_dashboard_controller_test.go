package controllers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/controllers"
	"github.com/iips-oss/ispark/api/models"
)

func TestStudentDashboardFlow(t *testing.T) {
	t.Setenv("TESTING", "true")
	SetupTestDB(t)

	app := fiber.New()

	// Dummy middleware to simulate authenticated student
	studentAuth := func(rollNo string) fiber.Handler {
		return func(c *fiber.Ctx) error {
			c.Locals("roll_no", rollNo)
			return c.Next()
		}
	}

	studentRoll := "2024-MCA-01"

	// Seed test student
	student := models.Student{
		RollNo:        studentRoll,
		Name:          "Test Student",
		CourseName:    models.CourseMCA5Yr,
		Semester:      2,
		EmailID:       "student@test.com",
		EnrollmentNo:  "EN202401",
		IsVerified:    true,
		Status:        "Active",
	}
	if err := config.DB.Create(&student).Error; err != nil {
		t.Fatalf("Failed to seed student: %v", err)
	}

	// Register routes
	api := app.Group("/api/student", studentAuth(studentRoll))
	api.Get("/notifications", controllers.GetStudentNotifications)
	api.Get("/certificates", controllers.GetCertificates)
	api.Post("/certificates", controllers.UploadCertificate)
	api.Get("/certificates/:id/file", controllers.DownloadCertificate)
	api.Get("/dashboard/stats", controllers.GetDashboardStats)
	api.Get("/activities", controllers.GetActivities)
	api.Post("/activities/:id/enroll", controllers.EnrollActivity)
	api.Get("/enrollments", controllers.GetEnrollments)

	t.Run("GetStudentNotifications_Success", func(t *testing.T) {
		// Seed admin note/notification
		note := models.AdminNote{
			StudentRollNo: studentRoll,
			AuthorName:    "Admin John",
			Role:          "admin",
			Text:          "[ALERT REMINDER] Complete activity submission",
		}
		config.DB.Create(&note)

		req := httptest.NewRequest("GET", "/api/student/notifications", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var notes []controllers.StudentNotificationResponse
		json.NewDecoder(resp.Body).Decode(&notes)
		if len(notes) == 0 || notes[0].Text != "Complete activity submission" {
			t.Errorf("Unexpected notification response: %v", notes)
		}
	})

	t.Run("GetDashboardStats_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/student/dashboard/stats", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadCertificate_ValidationErrors", func(t *testing.T) {
		// Missing required fields
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/student/certificates", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for missing fields, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadCertificate_ValidPNG", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("activity_name", "Hackathon 2026")
		_ = writer.WriteField("activity_category", "Technical")
		_ = writer.WriteField("activity_date", "2026-03-15")
		_ = writer.WriteField("participation_type", "Winner")
		_ = writer.WriteField("event_level", "National")

		// Create dummy PNG content (PNG magic header: \x89PNG\r\n\x1a\n)
		pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
		part, _ := writer.CreateFormFile("certificate_file", "cert.png")
		part.Write(pngHeader)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/student/certificates", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201 Created for valid certificate upload, got %d", resp.StatusCode)
		}

		// Verify certificate created in DB
		var cert models.Certificate
		if err := config.DB.Where("student_roll_no = ? AND activity_name = ?", studentRoll, "Hackathon 2026").First(&cert).Error; err != nil {
			t.Errorf("Certificate not saved in DB: %v", err)
		}
	})

	t.Run("DownloadCertificate_OwnershipCheck", func(t *testing.T) {
		// Attempting to download non-existent certificate
		req := httptest.NewRequest("GET", "/api/student/certificates/9999/file", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found for non-existent certificate, got %d", resp.StatusCode)
		}
	})

	t.Run("ActivityEnrollmentFlow", func(t *testing.T) {
		// Seed activity
		act := models.Activity{
			Name:         "AI Seminar",
			Category:     "TECHNICAL",
			Type:         "Workshop",
			Mode:         "Offline",
			ActivityDate: time.Now().AddDate(0, 0, 5),
			Venue:        "Auditorium",
			Status:       "Open",
		}
		config.DB.Create(&act)

		// Get activities
		reqGetAct := httptest.NewRequest("GET", "/api/student/activities", nil)
		respGetAct, err := app.Test(reqGetAct)
		if err != nil || respGetAct.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for activities, got %d", respGetAct.StatusCode)
		}

		// Get enrollments
		reqGet := httptest.NewRequest("GET", "/api/student/enrollments", nil)
		respGet, err := app.Test(reqGet)
		if err != nil {
			t.Fatalf("Failed to execute get enrollments: %v", err)
		}
		if respGet.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for enrollments, got %d", respGet.StatusCode)
		}
	})
}
