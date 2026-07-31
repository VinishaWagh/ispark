package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/routes"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDBOnce sync.Once

// sharedTestDB is the database created once by SetupTestDB. It is retained
// separately from config.DB so SetupTestDB can restore it on every call: other
// test files intentionally swap config.DB for their own isolated databases, and
// without this a test that runs after one of those swaps would otherwise
// inherit a database missing the tables SetupTestDB migrates.
var sharedTestDB *gorm.DB

// SetupTestDB initializes an in-memory SQLite database for testing and points
// config.DB at it. The database is created and migrated once; every call
// restores config.DB to it and clears the tables for a clean slate.
func SetupTestDB(t *testing.T) {
	t.Setenv("TESTING", "true")
	testDBOnce.Do(func() {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			t.Fatalf("Failed to connect to in-memory SQLite database: %v", err)
		}

		// Auto-migrate all tables used in testing
		err = db.AutoMigrate(
			&models.Student{},
			&models.OTP{},
			&models.Admin{},
			&models.Activity{},
			&models.Certificate{},
			&models.Enrollment{},
			&models.Notification{},
			&models.AdminNote{},
			&models.Track{},
			&models.BatchOverride{},
			&models.GeneratedReport{},
			&models.ScheduledReport{},
			&models.ReportAuditLog{},
		)
		if err != nil {
			t.Fatalf("Failed to run database migrations: %v", err)
		}

		sharedTestDB = db
	})

	// Restore the shared database in case a previous test swapped config.DB out
	// for its own isolated instance.
	config.DB = sharedTestDB

	// Clear all tables to guarantee a clean slate. Unscoped is required because
	// most models use gorm.DeletedAt: a plain Delete only soft-deletes, leaving
	// the rows in place so that re-seeding a fixture with the same primary key
	// (e.g. Admin.AdminID) fails on the unique constraint.
	if config.DB != nil {
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Student{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.OTP{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Admin{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Activity{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Certificate{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Enrollment{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Notification{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.AdminNote{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Track{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.BatchOverride{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.GeneratedReport{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ScheduledReport{})
		config.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ReportAuditLog{})
	}
}

// Helper to solve the simple math captcha challenge
func solveCaptcha(question string) (string, error) {
	// Question format is: "X + Y = ?"
	re := regexp.MustCompile(`(\d+)\s*\+\s*(\d+)`)
	matches := re.FindStringSubmatch(question)
	if len(matches) < 3 {
		return "", fmt.Errorf("unable to parse captcha question: %s", question)
	}

	val1, err1 := strconv.Atoi(matches[1])
	val2, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("failed to convert numbers in captcha: %v, %v", err1, err2)
	}

	return strconv.Itoa(val1 + val2), nil
}

func TestAuthFlow(t *testing.T) {
	// Set test environment variables
	t.Setenv("JWT_SECRET", strings.Repeat("test-jwt-", 4))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("test-refresh-jwt-", 4))

	// Initialize test DB
	SetupTestDB(t)

	// Create and setup Fiber App
	app := fiber.New()
	routes.SetupRoutes(app)

	// Test registration details
	testStudentEmail := "test.student@example.com"
	testPassword := "SecurePassword123!"

	// 1. TEST REGISTRATION FLOW
	t.Run("RegisterStudent_Success", func(t *testing.T) {
		regPayload := map[string]interface{}{
			"name":             "John Doe",
			"roll_no":          "12345",
			"course_name":      "MCA",
			"semester":         2,
			"contact_no":       "9876543210",
			"email_id":         testStudentEmail,
			"enrollment_no":    "EN12345",
			"password":         testPassword,
			"confirm_password": testPassword,
		}
		body, _ := json.Marshal(regPayload)

		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute register request: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
		}

		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if respBody["email"] != testStudentEmail {
			t.Errorf("Expected registered email to be %s, got %v", testStudentEmail, respBody["email"])
		}

		var registered models.Student
		if err := config.DB.First(&registered, "email_id = ?", testStudentEmail).Error; err != nil {
			t.Fatalf("load registered student: %v", err)
		}
		if registered.CourseName != models.CourseMCA5Yr {
			t.Errorf("expected canonical course %q, got %q", models.CourseMCA5Yr, registered.CourseName)
		}
	})

	t.Run("RegisterStudent_DuplicateConflict", func(t *testing.T) {
		// Attempting to register again with same details
		regPayload := map[string]interface{}{
			"name":             "John Doe Duplicate",
			"roll_no":          "12345",
			"course_name":      "MCA",
			"semester":         2,
			"contact_no":       "9876543210",
			"email_id":         testStudentEmail,
			"enrollment_no":    "EN12345",
			"password":         testPassword,
			"confirm_password": testPassword,
		}
		body, _ := json.Marshal(regPayload)

		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute register request: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
		}

		// Since they are not verified, the old registration is cleared if it's the same student.
		// However, let's verify if registering a different field raises conflict or returns StatusCreated.
		// Wait, according to controllers/auth_controller.go, if they are not verified yet,
		// the unverified record is deleted and it re-registers (status 201 Created).
		// But let's check with a verified student.
		// Let's mark the student as verified in the DB to test the conflict response.
		var dbStudent models.Student
		config.DB.Where("email_id = ?", testStudentEmail).First(&dbStudent)
		dbStudent.IsVerified = true
		config.DB.Save(&dbStudent)

		// Now attempt duplicate registration
		resp2, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute duplicate register request: %v", err)
		}

		if resp2.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 Conflict for duplicate registration, got %d", resp2.StatusCode)
		}
	})

	// 2. TEST OTP VALIDATION
	t.Run("VerifyOTP_Success", func(t *testing.T) {
		// Reset database and re-register unverified student
		SetupTestDB(t)

		regPayload := map[string]interface{}{
			"name":             "Alice Smith",
			"roll_no":          "67890",
			"course_name":      models.CourseMTechCS,
			"semester":         4,
			"contact_no":       "9876543211",
			"email_id":         "alice@example.com",
			"enrollment_no":    "EN67890",
			"password":         testPassword,
			"confirm_password": testPassword,
		}
		body, _ := json.Marshal(regPayload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		if _, err := app.Test(req); err != nil {
			t.Fatalf("Failed to execute register request: %v", err)
		}

		// Get OTP code from SQLite
		var otp models.OTP
		err := config.DB.Where("email = ? AND purpose = ?", "alice@example.com", "register").First(&otp).Error
		if err != nil {
			t.Fatalf("Could not find generated OTP in database: %v", err)
		}

		// Verify OTP with correct code
		verifyPayload := map[string]string{
			"email": "alice@example.com",
			"code":  otp.Code,
		}
		vBody, _ := json.Marshal(verifyPayload)

		vReq := httptest.NewRequest("POST", "/api/auth/verify-otp", bytes.NewBuffer(vBody))
		vReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(vReq)
		if err != nil {
			t.Fatalf("Failed to execute verify OTP request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
		}

		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if respBody["access_token"] == nil {
			t.Error("Expected access token in response, got nil")
		}
	})

	t.Run("VerifyOTP_Failure_InvalidCode", func(t *testing.T) {
		verifyPayload := map[string]string{
			"email": "alice@example.com",
			"code":  "999999", // Wrong code
		}
		vBody, _ := json.Marshal(verifyPayload)

		vReq := httptest.NewRequest("POST", "/api/auth/verify-otp", bytes.NewBuffer(vBody))
		vReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(vReq)
		if err != nil {
			t.Fatalf("Failed to execute verify OTP request: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	// 3. TEST LOGIN FAILURE/SUCCESS & PROFILE ACCESS
	t.Run("LoginAndProfileFlow", func(t *testing.T) {
		// Fetch a captcha challenge first
		capReq := httptest.NewRequest("GET", "/api/auth/captcha", nil)
		capResp, err := app.Test(capReq)
		if err != nil {
			t.Fatalf("Failed to get captcha: %v", err)
		}
		if capResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected captcha status 200 OK, got %d", capResp.StatusCode)
		}

		var capBody map[string]interface{}
		if err := json.NewDecoder(capResp.Body).Decode(&capBody); err != nil {
			t.Fatalf("Failed to decode captcha response: %v", err)
		}
		captchaID := capBody["captcha_id"].(string)
		question := capBody["question"].(string)

		correctAnswer, err := solveCaptcha(question)
		if err != nil {
			t.Fatalf("Error solving captcha: %v", err)
		}

		// A. Login Failure - Wrong Captcha
		loginWrongCaptcha := map[string]string{
			"email_id":       "alice@example.com",
			"password":       testPassword,
			"captcha_id":     captchaID,
			"captcha_answer": "999", // Wrong answer
		}
		wBody, _ := json.Marshal(loginWrongCaptcha)
		wReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(wBody))
		wReq.Header.Set("Content-Type", "application/json")
		wResp, err := app.Test(wReq)
		if err != nil {
			t.Fatalf("Failed to execute login request: %v", err)
		}
		if wResp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for wrong captcha, got %d", wResp.StatusCode)
		}

		// B. Login Failure - Wrong Password
		loginWrongPass := map[string]string{
			"email_id":       "alice@example.com",
			"password":       "WrongPassword!",
			"captcha_id":     captchaID,
			"captcha_answer": correctAnswer,
		}
		wpBody, _ := json.Marshal(loginWrongPass)
		wpReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(wpBody))
		wpReq.Header.Set("Content-Type", "application/json")
		wpResp, err := app.Test(wpReq)
		if err != nil {
			t.Fatalf("Failed to execute login request: %v", err)
		}
		if wpResp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized for wrong credentials, got %d", wpResp.StatusCode)
		}

		// C. Login Success
		loginSuccessPayload := map[string]string{
			"email_id":       "alice@example.com",
			"password":       testPassword,
			"captcha_id":     captchaID,
			"captcha_answer": correctAnswer,
		}
		sBody, _ := json.Marshal(loginSuccessPayload)
		sReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(sBody))
		sReq.Header.Set("Content-Type", "application/json")
		sResp, err := app.Test(sReq)
		if err != nil {
			t.Fatalf("Failed to execute login request: %v", err)
		}
		if sResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for successful login, got %d", sResp.StatusCode)
		}

		var sBodyMap map[string]interface{}
		if err := json.NewDecoder(sResp.Body).Decode(&sBodyMap); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}
		accessToken := sBodyMap["access_token"].(string)

		// D. Profile Access - Unauthenticated
		profReqUnauth := httptest.NewRequest("GET", "/api/auth/profile", nil)
		profRespUnauth, err := app.Test(profReqUnauth)
		if err != nil {
			t.Fatalf("Failed to execute profile request: %v", err)
		}
		if profRespUnauth.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized for missing token, got %d", profRespUnauth.StatusCode)
		}

		// E. Profile Access - Authenticated
		profReqAuth := httptest.NewRequest("GET", "/api/auth/profile", nil)
		profReqAuth.Header.Set("Authorization", "Bearer "+accessToken)
		profRespAuth, err := app.Test(profReqAuth)
		if err != nil {
			t.Fatalf("Failed to execute profile request: %v", err)
		}
		if profRespAuth.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for authenticated profile access, got %d", profRespAuth.StatusCode)
		}

		var profBody map[string]interface{}
		if err := json.NewDecoder(profRespAuth.Body).Decode(&profBody); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		studentData := profBody["student"].(map[string]interface{})
		if studentData["email_id"] != "alice@example.com" {
			t.Errorf("Expected profile email to be alice@example.com, got %v", studentData["email_id"])
		}
	})
}

func TestRegisterRejectsInvalidAcademicDetails(t *testing.T) {
	SetupTestDB(t)
	app := fiber.New()
	routes.SetupRoutes(app)

	base := map[string]interface{}{
		"name":             "Academic Validation",
		"roll_no":          "ACA001",
		"course_name":      models.CourseMCA5Yr,
		"semester":         1,
		"contact_no":       "9876543299",
		"email_id":         "academic.validation@example.com",
		"enrollment_no":    "EN-ACA001",
		"password":         "SecurePassword123!",
		"confirm_password": "SecurePassword123!",
	}

	tests := []struct {
		name  string
		field string
		value interface{}
	}{
		{"unsupported course", "course_name", "Unknown Programme"},
		{"semester above range", "semester", 11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := make(map[string]interface{}, len(base))
			for key, value := range base {
				payload[key] = value
			}
			payload[test.field] = test.value

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("register: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestInactiveAndFirstLoginFlows(t *testing.T) {
	// Set test environment variables
	t.Setenv("JWT_SECRET", strings.Repeat("test-jwt-", 4))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("test-refresh-jwt-", 4))

	SetupTestDB(t)

	app := fiber.New()
	routes.SetupRoutes(app)

	// Password to use
	testPassword := "TestPass123!"
	hashed, _ := utils.HashPassword(testPassword)

	// Create an unverified/pending student
	student := models.Student{
		RollNo:       "STU001",
		Name:         "Pending Student",
		CourseName:   "MCA",
		Semester:     1,
		ContactNo:    "1234567890",
		EmailID:      "pending@student.com",
		EnrollmentNo: "EN-STU001",
		Password:     hashed,
		IsVerified:   false,
		Status:       "Pending",
	}
	if err := config.DB.Create(&student).Error; err != nil {
		t.Fatalf("Failed to create pending student: %v", err)
	}

	// 1. TEST FAILED FIRST LOGIN FLOW
	t.Run("Login_Unverified_FirstAttempt", func(t *testing.T) {
		// A. Fetch a captcha challenge first
		capReq := httptest.NewRequest("GET", "/api/auth/captcha", nil)
		capResp, err := app.Test(capReq)
		if err != nil {
			t.Fatalf("Failed to get captcha: %v", err)
		}
		var capBody map[string]interface{}
		if err := json.NewDecoder(capResp.Body).Decode(&capBody); err != nil {
			t.Fatalf("Failed to decode captcha response: %v", err)
		}
		captchaID := capBody["captcha_id"].(string)
		question := capBody["question"].(string)

		correctAnswer, err := solveCaptcha(question)
		if err != nil {
			t.Fatalf("Error solving captcha: %v", err)
		}

		// B. Try to login
		loginPayload := map[string]string{
			"email_id":       "pending@student.com",
			"password":       testPassword,
			"captcha_id":     captchaID,
			"captcha_answer": correctAnswer,
		}
		body, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed login attempt: %v", err)
		}

		// First login of unverified student must return 403 Forbidden with email
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403 Forbidden, got %d", resp.StatusCode)
		}

		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}
		if respBody["email"] != "pending@student.com" {
			t.Errorf("Expected response email to be pending@student.com, got %v", respBody["email"])
		}

		// C. Retrieve the created OTP
		var otp models.OTP
		err = config.DB.Where("email = ? AND purpose = ?", "pending@student.com", "register").First(&otp).Error
		if err != nil {
			t.Fatalf("Could not find registration reactivation OTP in database: %v", err)
		}

		// D. Verify the OTP
		verifyPayload := map[string]string{
			"email": "pending@student.com",
			"code":  otp.Code,
		}
		vBody, _ := json.Marshal(verifyPayload)
		vReq := httptest.NewRequest("POST", "/api/auth/verify-otp", bytes.NewBuffer(vBody))
		vReq.Header.Set("Content-Type", "application/json")

		vResp, err := app.Test(vReq)
		if err != nil {
			t.Fatalf("Failed verify OTP request: %v", err)
		}

		if vResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK after OTP verification, got %d", vResp.StatusCode)
		}

		var vRespBody map[string]interface{}
		if err := json.NewDecoder(vResp.Body).Decode(&vRespBody); err != nil {
			t.Fatalf("Failed to decode verify response: %v", err)
		}
		if vRespBody["access_token"] == nil {
			t.Error("Expected access token after successful OTP verification, got nil")
		}

		// Check that student status is now Active and IsVerified is true
		var updatedStudent models.Student
		config.DB.Where("roll_no = ?", "STU001").First(&updatedStudent)
		if !updatedStudent.IsVerified {
			t.Error("Expected student IsVerified to be true")
		}
		if updatedStudent.Status != "Active" {
			t.Errorf("Expected student Status to be Active, got %s", updatedStudent.Status)
		}
	})

	// 2. TEST STATUS AFTER SERVER REFETCH (INACTIVE STATUS)
	t.Run("Deactivate_Refetch_BlockLogin", func(t *testing.T) {
		// Create a superadmin token to access platform endpoints
		saToken, err := utils.GenerateAccessToken("SA001", "superadmin@example.com", "superadmin")
		if err != nil {
			t.Fatalf("Failed to generate superadmin token: %v", err)
		}

		// A. Update student status to Inactive via PUT /api/admin/platform/users/:id
		updatePayload := map[string]interface{}{
			"status": "Inactive",
		}
		uBody, _ := json.Marshal(updatePayload)
		uReq := httptest.NewRequest("PUT", "/api/admin/platform/users/STU001", bytes.NewBuffer(uBody))
		uReq.Header.Set("Content-Type", "application/json")
		uReq.Header.Set("Authorization", "Bearer "+saToken)

		uResp, err := app.Test(uReq)
		if err != nil {
			t.Fatalf("Failed to execute update user request: %v", err)
		}
		if uResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for update user, got %d", uResp.StatusCode)
		}

		// B. Refetch the user registry via GET /api/admin/platform/users
		rReq := httptest.NewRequest("GET", "/api/admin/platform/users", nil)
		rReq.Header.Set("Authorization", "Bearer "+saToken)

		rResp, err := app.Test(rReq)
		if err != nil {
			t.Fatalf("Failed to execute refetch users request: %v", err)
		}
		if rResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for refetch users, got %d", rResp.StatusCode)
		}

		var rBody map[string]interface{}
		if err := json.NewDecoder(rResp.Body).Decode(&rBody); err != nil {
			t.Fatalf("Failed to decode refetch users response: %v", err)
		}
		usersList := rBody["users"].([]interface{})

		// Find the student in the list and verify status is Inactive
		foundInactive := false
		for _, u := range usersList {
			userMap := u.(map[string]interface{})
			if userMap["id"] == "STU001" {
				if userMap["status"] != "Inactive" {
					t.Errorf("Expected refetched student status to be Inactive, got %v", userMap["status"])
				}
				foundInactive = true
				break
			}
		}
		if !foundInactive {
			t.Error("Did not find updated student in user registry")
		}

		// C. Attempt login as Inactive student
		capReq := httptest.NewRequest("GET", "/api/auth/captcha", nil)
		capResp, err := app.Test(capReq)
		if err != nil {
			t.Fatalf("Failed to get captcha: %v", err)
		}
		var capBody map[string]interface{}
		if err := json.NewDecoder(capResp.Body).Decode(&capBody); err != nil {
			t.Fatalf("Failed to decode captcha response: %v", err)
		}
		captchaID := capBody["captcha_id"].(string)
		question := capBody["question"].(string)
		correctAnswer, _ := solveCaptcha(question)

		loginPayload := map[string]string{
			"email_id":       "pending@student.com",
			"password":       testPassword,
			"captcha_id":     captchaID,
			"captcha_answer": correctAnswer,
		}
		body, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed login attempt: %v", err)
		}

		// Inactive student login must return 403 Forbidden with account inactive error
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403 Forbidden for inactive student login, got %d", resp.StatusCode)
		}

		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if respBody["error"] != "Your account is inactive. Please contact the administrator." {
			t.Errorf("Expected inactive error message, got: %v", respBody["error"])
		}
		// Confirm it does NOT return an email key to prevent frontend OTP transition
		if respBody["email"] != nil {
			t.Errorf("Expected email to be nil for inactive account login, got %v", respBody["email"])
		}
	})
}

func TestEmailValidationAndAdminStatus(t *testing.T) {
	// Set test environment variables
	t.Setenv("JWT_SECRET", strings.Repeat("test-jwt-", 4))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("test-refresh-jwt-", 4))

	SetupTestDB(t)

	app := fiber.New()
	routes.SetupRoutes(app)

	saToken, err := utils.GenerateAccessToken("SA001", "superadmin@example.com", "superadmin")
	if err != nil {
		t.Fatalf("Failed to generate superadmin token: %v", err)
	}

	// 1. Invalid Email format validation tests
	t.Run("CreateUser_InvalidEmailFormat", func(t *testing.T) {
		payloads := []map[string]interface{}{
			{"name": "Student A", "id": "STU101", "email": "x@.", "role": "Student", "dept": "CS", "semester": 1},
			{"name": "Student B", "id": "STU102", "email": "invalidemail", "role": "Student", "dept": "CS", "semester": 1},
			{"name": "Admin A", "id": "ADM101", "email": "invalidemail@", "role": "Admin", "dept": "CS"},
		}

		for _, payload := range payloads {
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/admin/platform/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+saToken)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400 Bad Request for invalid email %s, got %d", payload["email"], resp.StatusCode)
			}
		}
	})

	// 2. Same-table case variant checks (e.g. Student email case duplicate check)
	t.Run("CreateUser_SameTableCaseVariantEmailDuplicate", func(t *testing.T) {
		// First create student using Case.Dupe@example.com
		payload1 := map[string]interface{}{
			"name": "Student Dupe 1", "id": "STU103", "email": "Case.Dupe@example.com", "role": "Student", "dept": "CS", "semester": 1,
		}
		body1, _ := json.Marshal(payload1)
		req1 := httptest.NewRequest("POST", "/api/admin/platform/users", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+saToken)
		resp1, err := app.Test(req1)
		if err != nil {
			t.Fatalf("Failed to create first student: %v", err)
		}
		if resp1.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201 Created for first student, got %d", resp1.StatusCode)
		}

		// Now attempt to create student using case.dupe@example.com
		payload2 := map[string]interface{}{
			"name": "Student Dupe 2", "id": "STU104", "email": "case.dupe@example.com", "role": "Student", "dept": "CS", "semester": 1,
		}
		body2, _ := json.Marshal(payload2)
		req2 := httptest.NewRequest("POST", "/api/admin/platform/users", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+saToken)
		resp2, err := app.Test(req2)
		if err != nil {
			t.Fatalf("Failed to attempt duplicate student creation: %v", err)
		}
		if resp2.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 Conflict for same-table duplicate case variant email, got %d", resp2.StatusCode)
		}
	})

	// 3. Cross-table case variant checks (Admin vs Student)
	t.Run("CreateUser_CrossTableCaseVariantEmailDuplicate", func(t *testing.T) {
		// Attempt to create an Admin with the duplicate email (e.g. CASE.DUPE@example.com)
		payload := map[string]interface{}{
			"name": "Admin Dupe", "id": "ADM102", "email": "CASE.DUPE@example.com", "role": "Admin", "dept": "CS",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/admin/platform/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+saToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to attempt cross-table duplicate admin creation: %v", err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 Conflict for cross-table duplicate case variant email, got %d", resp.StatusCode)
		}
	})

	// 4. Admin status update and enforcement tests
	t.Run("AdminStatus_UpdateAndEnforce", func(t *testing.T) {
		// Create a test Admin using CreatePlatformUser
		adminID := "ADM201"
		adminEmail := "admin.test@example.com"
		adminPayload := map[string]interface{}{
			"name": "Test Admin", "id": adminID, "email": adminEmail, "role": "Admin", "dept": "CS",
		}
		body1, _ := json.Marshal(adminPayload)
		req1 := httptest.NewRequest("POST", "/api/admin/platform/users", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+saToken)
		resp1, err := app.Test(req1)
		if err != nil {
			t.Fatalf("Failed to create admin: %v", err)
		}
		if resp1.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201 Created for admin creation, got %d", resp1.StatusCode)
		}

		var createdBody map[string]interface{}
		if err := json.NewDecoder(resp1.Body).Decode(&createdBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		tempPassword := createdBody["temporary_password"].(string)

		// Verify initial status is Active via GET
		reqGet1 := httptest.NewRequest("GET", "/api/admin/platform/users", nil)
		reqGet1.Header.Set("Authorization", "Bearer "+saToken)
		respGet1, err := app.Test(reqGet1)
		if err != nil {
			t.Fatalf("Failed to get platform users: %v", err)
		}
		var getBody1 map[string]interface{}
		if err := json.NewDecoder(respGet1.Body).Decode(&getBody1); err != nil {
			t.Fatalf("Failed to decode platform users: %v", err)
		}
		users := getBody1["users"].([]interface{})
		foundAdmin := false
		for _, u := range users {
			uMap := u.(map[string]interface{})
			if uMap["id"] == adminID {
				if uMap["status"] != "Active" {
					t.Errorf("Expected admin status to initially be Active, got %v", uMap["status"])
				}
				foundAdmin = true
				break
			}
		}
		if !foundAdmin {
			t.Fatalf("Did not find newly created admin in user registry")
		}

		// Try updating Admin status to an invalid value (should fail with 400)
		invalidAdminPayload := map[string]interface{}{
			"status": "Bananas",
		}
		bodyInvalidAdmin, _ := json.Marshal(invalidAdminPayload)
		reqInvalidAdmin := httptest.NewRequest("PUT", "/api/admin/platform/users/"+adminID, bytes.NewBuffer(bodyInvalidAdmin))
		reqInvalidAdmin.Header.Set("Content-Type", "application/json")
		reqInvalidAdmin.Header.Set("Authorization", "Bearer "+saToken)
		respInvalidAdmin, err := app.Test(reqInvalidAdmin)
		if err != nil {
			t.Fatalf("Failed to execute invalid admin update request: %v", err)
		}
		if respInvalidAdmin.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for invalid admin status update, got %d", respInvalidAdmin.StatusCode)
		}

		// Try updating Student status to an invalid value (should fail with 400)
		invalidStudentPayload := map[string]interface{}{
			"status": "Bananas",
		}
		bodyInvalidStudent, _ := json.Marshal(invalidStudentPayload)
		reqInvalidStudent := httptest.NewRequest("PUT", "/api/admin/platform/users/STU103", bytes.NewBuffer(bodyInvalidStudent))
		reqInvalidStudent.Header.Set("Content-Type", "application/json")
		reqInvalidStudent.Header.Set("Authorization", "Bearer "+saToken)
		respInvalidStudent, err := app.Test(reqInvalidStudent)
		if err != nil {
			t.Fatalf("Failed to execute invalid student update request: %v", err)
		}
		if respInvalidStudent.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for invalid student status update, got %d", respInvalidStudent.StatusCode)
		}

		// Update Admin status to Inactive via PUT
		updatePayload := map[string]interface{}{
			"status": "Inactive",
		}
		bodyUpdate, _ := json.Marshal(updatePayload)
		reqUpdate := httptest.NewRequest("PUT", "/api/admin/platform/users/"+adminID, bytes.NewBuffer(bodyUpdate))
		reqUpdate.Header.Set("Content-Type", "application/json")
		reqUpdate.Header.Set("Authorization", "Bearer "+saToken)
		respUpdate, err := app.Test(reqUpdate)
		if err != nil {
			t.Fatalf("Failed to update admin: %v", err)
		}
		if respUpdate.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 OK for updating admin status, got %d", respUpdate.StatusCode)
		}

		// Refetch admin and verify status is Inactive
		respGet2, err := app.Test(reqGet1)
		if err != nil {
			t.Fatalf("Failed to get platform users after update: %v", err)
		}
		var getBody2 map[string]interface{}
		if err := json.NewDecoder(respGet2.Body).Decode(&getBody2); err != nil {
			t.Fatalf("Failed to decode platform users: %v", err)
		}
		users2 := getBody2["users"].([]interface{})
		for _, u := range users2 {
			uMap := u.(map[string]interface{})
			if uMap["id"] == adminID {
				if uMap["status"] != "Inactive" {
					t.Errorf("Expected updated admin status to be Inactive, got %v", uMap["status"])
				}
				break
			}
		}

		// Try logging in with updated admin account (should be blocked)
		loginPayload := map[string]interface{}{
			"admin_id": adminID,
			"password": tempPassword,
		}
		bodyLogin, _ := json.Marshal(loginPayload)
		reqLogin := httptest.NewRequest("POST", "/api/admin/auth/login", bytes.NewBuffer(bodyLogin))
		reqLogin.Header.Set("Content-Type", "application/json")
		respLogin, err := app.Test(reqLogin)
		if err != nil {
			t.Fatalf("Failed to execute login request: %v", err)
		}

		// Inactive admin login must return 403 Forbidden with account inactive error
		if respLogin.StatusCode != http.StatusForbidden {
			t.Errorf("Expected login status 403 Forbidden for inactive admin, got %d", respLogin.StatusCode)
		}
		var loginRespBody map[string]interface{}
		if err := json.NewDecoder(respLogin.Body).Decode(&loginRespBody); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}
		if loginRespBody["error"] != "Your account is inactive. Please contact the administrator." {
			t.Errorf("Expected inactive admin error message, got: %v", loginRespBody["error"])
		}
	})
}
