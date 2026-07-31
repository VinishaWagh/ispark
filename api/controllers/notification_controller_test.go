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
	"github.com/iips-oss/ispark/api/models"
	"github.com/iips-oss/ispark/api/routes"
	"github.com/iips-oss/ispark/api/utils"
	"gorm.io/gorm"
)

// setupNotificationApp boots the in-memory DB and a fully-routed Fiber app so
// notification tests exercise the real middleware + routing, not just handlers.
func setupNotificationApp(t *testing.T) *fiber.App {
	t.Helper()
	t.Setenv("JWT_SECRET", strings.Repeat("test-jwt-", 4))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("test-refresh-jwt-", 4))

	SetupTestDB(t)

	app := fiber.New()
	routes.SetupRoutes(app)
	return app
}

// seedNotificationStudent inserts a verified student and returns it.
func seedNotificationStudent(t *testing.T, roll, email string) models.Student {
	t.Helper()
	hashed, _ := utils.HashPassword("Password123")
	student := models.Student{
		RollNo:       roll,
		Name:         "Test Student",
		CourseName:   models.CourseMCA5Yr,
		Semester:     3,
		ContactNo:    "9999999999",
		EmailID:      email,
		EnrollmentNo: "EN-" + roll,
		Password:     hashed,
		IsVerified:   true,
		Status:       "Active",
	}
	if err := config.DB.Create(&student).Error; err != nil {
		t.Fatalf("Failed to seed student %s: %v", roll, err)
	}
	return student
}

// studentAuthToken mints an access token for a student roll number.
func studentAuthToken(t *testing.T, roll, email string) string {
	t.Helper()
	token, err := utils.GenerateAccessToken(roll, email, "student")
	if err != nil {
		t.Fatalf("Failed to generate student token: %v", err)
	}
	return token
}

// doAuthJSON issues an authenticated request and decodes the JSON response.
func doAuthJSON(t *testing.T, app *fiber.App, method, path, token string, body any) (*http.Response, map[string]any) {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}

	decoded := map[string]any{}
	if res.Body != nil {
		_ = json.NewDecoder(res.Body).Decode(&decoded)
	}
	return res, decoded
}

// seedNotification inserts a notification for a student with an explicit read state.
func seedNotification(t *testing.T, roll, title, message, nType string, isRead bool) models.Notification {
	t.Helper()
	n := models.Notification{
		StudentRollNo: roll,
		Title:         title,
		Message:       message,
		Type:          nType,
		IsRead:        isRead,
	}
	if err := config.DB.Create(&n).Error; err != nil {
		t.Fatalf("Failed to seed notification: %v", err)
	}
	return n
}

// seedNotificationAdmin inserts a superadmin and returns an access token for it.
func seedNotificationAdmin(t *testing.T, adminID, email string) string {
	t.Helper()
	admin := models.Admin{
		AdminID:  adminID,
		Name:     "Super Admin",
		Email:    email,
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	token, err := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}
	return token
}

// resetAnnouncements migrates the announcements table and empties it, so a test
// only ever sees the announcements it seeded itself.
func resetAnnouncements(t *testing.T) {
	t.Helper()
	if err := config.DB.AutoMigrate(&models.Announcement{}); err != nil {
		t.Fatalf("Failed to migrate announcements: %v", err)
	}
	if err := config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Announcement{}).Error; err != nil {
		t.Fatalf("Failed to clear announcements: %v", err)
	}
}

// breakNotificationStorage simulates a transient outage of notification storage
// by dropping the table, so every notification insert fails. The returned
// function restores it; it also runs on cleanup if a test fails before calling
// it, so the shared test database is never left broken for the next test.
func breakNotificationStorage(t *testing.T) func() {
	t.Helper()
	if err := config.DB.Migrator().DropTable(&models.Notification{}); err != nil {
		t.Fatalf("Failed to drop notifications table: %v", err)
	}

	restored := false
	restore := func() {
		if restored {
			return
		}
		restored = true
		if err := config.DB.AutoMigrate(&models.Notification{}); err != nil {
			t.Fatalf("Failed to restore notifications table: %v", err)
		}
	}
	t.Cleanup(restore)
	return restore
}

// countNotificationsTitled reports how many notifications of a title a student holds.
func countNotificationsTitled(t *testing.T, roll, title string) int64 {
	t.Helper()
	var count int64
	if err := config.DB.Model(&models.Notification{}).
		Where("student_roll_no = ? AND title = ?", roll, title).
		Count(&count).Error; err != nil {
		t.Fatalf("Failed to count notifications: %v", err)
	}
	return count
}

func TestGetNotifications(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22001", "mca001@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	seedNotification(t, student.RollNo, "First", "Oldest message", models.NotificationTypeGeneral, true)
	seedNotification(t, student.RollNo, "Second", "Unread message", models.NotificationTypeCredits, false)
	newest := seedNotification(t, student.RollNo, "Third", "Newest unread", models.NotificationTypeCertificate, false)

	t.Run("Unauthenticated", func(t *testing.T) {
		res, _ := doAuthJSON(t, app, http.MethodGet, "/api/student/notifications", "", nil)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", res.StatusCode)
		}
	})

	t.Run("ReturnsOwnNotificationsNewestFirst", func(t *testing.T) {
		res, body := doAuthJSON(t, app, http.MethodGet, "/api/student/notifications", token, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d (%v)", res.StatusCode, body)
		}
		list, ok := body["notifications"].([]any)
		if !ok || len(list) != 3 {
			t.Fatalf("expected 3 notifications, got %v", body["notifications"])
		}
		if unread := body["unread_count"].(float64); unread != 2 {
			t.Fatalf("expected unread_count 2, got %v", unread)
		}
		if total := body["total"].(float64); total != 3 {
			t.Fatalf("expected total 3, got %v", total)
		}
		first := list[0].(map[string]any)
		if uint(first["id"].(float64)) != newest.ID {
			t.Fatalf("expected newest notification first, got id %v", first["id"])
		}
	})
}

func TestGetUnreadNotificationCount(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22002", "mca002@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	seedNotification(t, student.RollNo, "A", "read", models.NotificationTypeGeneral, true)
	seedNotification(t, student.RollNo, "B", "unread", models.NotificationTypeGeneral, false)
	seedNotification(t, student.RollNo, "C", "unread", models.NotificationTypeGeneral, false)

	res, body := doAuthJSON(t, app, http.MethodGet, "/api/student/notifications/unread-count", token, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if unread := body["unread_count"].(float64); unread != 2 {
		t.Fatalf("expected unread_count 2, got %v", unread)
	}
}

func TestMarkNotificationRead(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22003", "mca003@isparc.dev")
	other := seedNotificationStudent(t, "MCA2K22004", "mca004@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	own := seedNotification(t, student.RollNo, "Mine", "unread", models.NotificationTypeGeneral, false)
	foreign := seedNotification(t, other.RollNo, "Theirs", "unread", models.NotificationTypeGeneral, false)

	t.Run("MarksOwnNotification", func(t *testing.T) {
		res, _ := doAuthJSON(t, app, http.MethodPut, fmt.Sprintf("/api/student/notifications/%d/read", own.ID), token, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var updated models.Notification
		config.DB.First(&updated, own.ID)
		if !updated.IsRead {
			t.Fatalf("expected notification to be marked read")
		}
	})

	t.Run("CannotMarkAnotherStudentsNotification", func(t *testing.T) {
		res, _ := doAuthJSON(t, app, http.MethodPut, fmt.Sprintf("/api/student/notifications/%d/read", foreign.ID), token, nil)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for foreign notification, got %d", res.StatusCode)
		}
		var still models.Notification
		config.DB.First(&still, foreign.ID)
		if still.IsRead {
			t.Fatalf("another student's notification must not be modified")
		}
	})
}

func TestMarkAllNotificationsRead(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22005", "mca005@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	seedNotification(t, student.RollNo, "A", "unread", models.NotificationTypeGeneral, false)
	seedNotification(t, student.RollNo, "B", "unread", models.NotificationTypeGeneral, false)
	seedNotification(t, student.RollNo, "C", "already read", models.NotificationTypeGeneral, true)

	res, body := doAuthJSON(t, app, http.MethodPut, "/api/student/notifications/read-all", token, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if updated := body["updated_count"].(float64); updated != 2 {
		t.Fatalf("expected updated_count 2, got %v", updated)
	}

	var remaining int64
	config.DB.Model(&models.Notification{}).Where("student_roll_no = ? AND is_read = ?", student.RollNo, false).Count(&remaining)
	if remaining != 0 {
		t.Fatalf("expected 0 unread notifications after read-all, got %d", remaining)
	}
}

func TestEnrollActivityCreatesNotification(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22006", "mca006@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	activity := models.Activity{Name: "Annual Hackathon", Category: "TECHNICAL", Credits: 10, Mode: "Offline"}
	if err := config.DB.Create(&activity).Error; err != nil {
		t.Fatalf("Failed to seed activity: %v", err)
	}

	res, _ := doAuthJSON(t, app, http.MethodPost, fmt.Sprintf("/api/student/activities/%d/enroll", activity.ID), token, nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}

	var n models.Notification
	if err := config.DB.Where("student_roll_no = ? AND type = ?", student.RollNo, models.NotificationTypeEnrollment).First(&n).Error; err != nil {
		t.Fatalf("expected an enrollment notification to be created: %v", err)
	}
	if !strings.Contains(n.Message, activity.Name) {
		t.Fatalf("expected notification to reference the activity, got %q", n.Message)
	}
}

func TestCertificateApprovalCreatesNotification(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22007", "mca007@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin1",
		Name:     "Super Admin",
		Email:    "super1@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	cert := models.Certificate{
		StudentRollNo:     student.RollNo,
		ActivityName:      "Robotics Workshop",
		ActivityCategory:  "TECHNICAL",
		ActivityDate:      time.Now(),
		ParticipationType: "Participant",
		Credits:           15,
		Status:            "Pending",
	}
	if err := config.DB.Create(&cert).Error; err != nil {
		t.Fatalf("Failed to seed certificate: %v", err)
	}

	res, _ := doAuthJSON(t, app, http.MethodPost, fmt.Sprintf("/api/admin/certificates/%d/approve", cert.ID), adminToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var n models.Notification
	if err := config.DB.Where("student_roll_no = ? AND type = ?", student.RollNo, models.NotificationTypeCertificate).First(&n).Error; err != nil {
		t.Fatalf("expected a certificate notification to be created: %v", err)
	}
	if !strings.Contains(n.Title, "Approved") {
		t.Fatalf("expected an approval notification, got title %q", n.Title)
	}
}

func TestSendNoticeCreatesNotification(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22008", "mca008@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin2",
		Name:     "Super Admin Two",
		Email:    "super2@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	res, _ := doAuthJSON(t, app, http.MethodPost, "/api/admin/students/"+student.RollNo+"/notice", adminToken, map[string]any{
		"message": "Please submit your pending certificate.",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var n models.Notification
	if err := config.DB.Where("student_roll_no = ? AND type = ?", student.RollNo, models.NotificationTypeNotice).First(&n).Error; err != nil {
		t.Fatalf("expected a notice notification to be created: %v", err)
	}
	if !strings.Contains(n.Message, "pending certificate") {
		t.Fatalf("expected notice message persisted, got %q", n.Message)
	}
}

func TestPublishAnnouncementCreatesNotificationsForStudents(t *testing.T) {
	app := setupNotificationApp(t)
	if err := config.DB.AutoMigrate(&models.Announcement{}); err != nil {
		t.Fatalf("Failed to migrate announcements: %v", err)
	}

	studentA := seedNotificationStudent(t, "MCA2K22009", "mca009@isparc.dev")
	studentB := seedNotificationStudent(t, "MCA2K22010", "mca010@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin3",
		Name:     "Super Admin Three",
		Email:    "super3@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	announcement := models.Announcement{
		Title:       "Semester Break Schedule",
		Description: "The institute reopens on August 1st.",
		Category:    "General",
		Audience:    "Students",
		Priority:    "High",
		PublishDate: time.Now().AddDate(0, 0, -1),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		Status:      "scheduled",
	}
	if err := config.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("Failed to seed announcement: %v", err)
	}

	res, _ := doAuthJSON(t, app, http.MethodPost, fmt.Sprintf("/api/admin/platform/announcements/%d/publish", announcement.ID), adminToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	for _, roll := range []string{studentA.RollNo, studentB.RollNo} {
		var count int64
		config.DB.Model(&models.Notification{}).Where("student_roll_no = ? AND type = ?", roll, models.NotificationTypeAnnouncement).Count(&count)
		if count != 1 {
			t.Fatalf("expected 1 announcement notification for %s, got %d", roll, count)
		}
	}
}

// Publishing is idempotent: the notifications belong to the announcement going
// live, not to the publish request. Re-publishing an already-active
// announcement must not deliver the same message to students a second time.
func TestPublishAnnouncementTwiceNotifiesStudentsOnce(t *testing.T) {
	app := setupNotificationApp(t)
	if err := config.DB.AutoMigrate(&models.Announcement{}); err != nil {
		t.Fatalf("Failed to migrate announcements: %v", err)
	}

	studentA := seedNotificationStudent(t, "MCA2K22011", "mca011@isparc.dev")
	studentB := seedNotificationStudent(t, "MCA2K22012", "mca012@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin4",
		Name:     "Super Admin Four",
		Email:    "super4@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	announcement := models.Announcement{
		Title:       "Updated Credit Policy Guidelines",
		Description: "Credit caps change from the next semester.",
		Category:    "Academic",
		Audience:    "All Users",
		Priority:    "High",
		PublishDate: time.Now().AddDate(0, 0, -1),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		Status:      "draft",
	}
	if err := config.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("Failed to seed announcement: %v", err)
	}

	path := fmt.Sprintf("/api/admin/platform/announcements/%d/publish", announcement.ID)
	for attempt := 1; attempt <= 2; attempt++ {
		res, body := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("publish attempt %d: expected 200, got %d (%v)", attempt, res.StatusCode, body)
		}
	}

	for _, roll := range []string{studentA.RollNo, studentB.RollNo} {
		var count int64
		config.DB.Model(&models.Notification{}).
			Where("student_roll_no = ? AND type = ? AND title = ?", roll, models.NotificationTypeAnnouncement, announcement.Title).
			Count(&count)
		if count != 1 {
			t.Fatalf("expected exactly 1 %q notification for %s after publishing twice, got %d", announcement.Title, roll, count)
		}
	}
}

// Approving an already-approved certificate is the same underlying decision, so
// a retried request must not produce a second notification.
func TestApproveCertificateTwiceNotifiesStudentOnce(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22013", "mca013@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin5",
		Name:     "Super Admin Five",
		Email:    "super5@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	cert := models.Certificate{
		StudentRollNo:     student.RollNo,
		ActivityName:      "Robotics Workshop",
		ActivityCategory:  "TECHNICAL",
		ActivityDate:      time.Now(),
		ParticipationType: "Participant",
		Credits:           15,
		Status:            "Pending",
	}
	if err := config.DB.Create(&cert).Error; err != nil {
		t.Fatalf("Failed to seed certificate: %v", err)
	}

	path := fmt.Sprintf("/api/admin/certificates/%d/approve", cert.ID)
	for attempt := 1; attempt <= 2; attempt++ {
		res, body := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("approve attempt %d: expected 200, got %d (%v)", attempt, res.StatusCode, body)
		}
	}

	var count int64
	config.DB.Model(&models.Notification{}).
		Where("student_roll_no = ? AND type = ?", student.RollNo, models.NotificationTypeCertificate).
		Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 certificate notification after approving twice, got %d", count)
	}
}

// A genuine change of decision is still a new event and must notify the student.
func TestRejectAfterApproveNotifiesStudentAgain(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22014", "mca014@isparc.dev")

	admin := models.Admin{
		AdminID:  "superadmin6",
		Name:     "Super Admin Six",
		Email:    "super6@isparc.dev",
		Password: "x",
		Role:     "superadmin",
	}
	if err := config.DB.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to seed admin: %v", err)
	}
	adminToken, _ := utils.GenerateAccessToken(admin.AdminID, admin.Email, admin.Role)

	cert := models.Certificate{
		StudentRollNo:     student.RollNo,
		ActivityName:      "Robotics Workshop",
		ActivityCategory:  "TECHNICAL",
		ActivityDate:      time.Now(),
		ParticipationType: "Participant",
		Credits:           15,
		Status:            "Pending",
	}
	if err := config.DB.Create(&cert).Error; err != nil {
		t.Fatalf("Failed to seed certificate: %v", err)
	}

	res, _ := doAuthJSON(t, app, http.MethodPost, fmt.Sprintf("/api/admin/certificates/%d/approve", cert.ID), adminToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("approve: expected 200, got %d", res.StatusCode)
	}
	res, _ = doAuthJSON(t, app, http.MethodPost, fmt.Sprintf("/api/admin/certificates/%d/reject", cert.ID), adminToken, map[string]any{
		"reason": "Certificate is illegible.",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("reject: expected 200, got %d", res.StatusCode)
	}

	var titles []string
	config.DB.Model(&models.Notification{}).
		Where("student_roll_no = ? AND type = ?", student.RollNo, models.NotificationTypeCertificate).
		Order("id asc").
		Pluck("title", &titles)
	if len(titles) != 2 || !strings.Contains(titles[0], "Approved") || !strings.Contains(titles[1], "Rejected") {
		t.Fatalf("expected an approval then a rejection notification, got %v", titles)
	}
}

// Creating an announcement directly as "active" is a publication, not just a
// row insert: the super-admin form offers the active status, so that path must
// deliver the announcement to students exactly as an explicit publish does.
func TestCreateActiveAnnouncementNotifiesStudents(t *testing.T) {
	app := setupNotificationApp(t)
	resetAnnouncements(t)

	studentA := seedNotificationStudent(t, "MCA2K22015", "mca015@isparc.dev")
	studentB := seedNotificationStudent(t, "MCA2K22016", "mca016@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin7", "super7@isparc.dev")

	title := "Library Timings Extended"
	res, body := doAuthJSON(t, app, http.MethodPost, "/api/admin/platform/announcements", adminToken, map[string]any{
		"title":        title,
		"description":  "The reading room now closes at 10 PM.",
		"category":     "General",
		"audience":     "Students",
		"priority":     "High",
		"publish_date": announcementDate(0),
		"expiry_date":  announcementDate(30),
		"status":       "active",
	})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%v)", res.StatusCode, body)
	}

	for _, roll := range []string{studentA.RollNo, studentB.RollNo} {
		if got := countNotificationsTitled(t, roll, title); got != 1 {
			t.Fatalf("expected 1 announcement notification for %s after an active create, got %d", roll, got)
		}
	}
}

// Editing a draft to "active" publishes it through the same form, so it must
// notify students too — and a later edit of the now-live announcement must not
// deliver the message a second time.
func TestUpdateAnnouncementToActiveNotifiesStudentsOnce(t *testing.T) {
	app := setupNotificationApp(t)
	resetAnnouncements(t)

	student := seedNotificationStudent(t, "MCA2K22017", "mca017@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin8", "super8@isparc.dev")

	title := "Placement Drive Registration"
	announcement := models.Announcement{
		Title:       title,
		Description: "Register before Friday.",
		Category:    "Academic",
		Audience:    "All Users",
		Priority:    "High",
		PublishDate: time.Now().AddDate(0, 0, -1),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		Status:      "draft",
	}
	if err := config.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("Failed to seed announcement: %v", err)
	}

	path := fmt.Sprintf("/api/admin/platform/announcements/%d", announcement.ID)
	res, body := doAuthJSON(t, app, http.MethodPut, path, adminToken, map[string]any{"status": "active"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d (%v)", res.StatusCode, body)
	}
	if got := countNotificationsTitled(t, student.RollNo, title); got != 1 {
		t.Fatalf("expected 1 announcement notification after updating to active, got %d", got)
	}

	// Editing the live announcement again is not a new publication.
	res, body = doAuthJSON(t, app, http.MethodPut, path, adminToken, map[string]any{
		"description": "Register before Friday. Bring two copies of your resume.",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on the follow-up edit, got %d (%v)", res.StatusCode, body)
	}
	if got := countNotificationsTitled(t, student.RollNo, title); got != 1 {
		t.Fatalf("expected still 1 announcement notification after editing a live announcement, got %d", got)
	}
}

// A scheduled announcement reaching its publish date goes live during a status
// refresh. That activation is a publication and must notify students, and doing
// it twice (two refreshes) must not deliver the message twice.
func TestScheduledAnnouncementActivationNotifiesStudents(t *testing.T) {
	app := setupNotificationApp(t)
	resetAnnouncements(t)

	student := seedNotificationStudent(t, "MCA2K22018", "mca018@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin9", "super9@isparc.dev")

	title := "Mid-Semester Exam Timetable"
	announcement := models.Announcement{
		Title:       title,
		Description: "The timetable is now available on the portal.",
		Category:    "Academic",
		Audience:    "Students",
		Priority:    "High",
		PublishDate: time.Now().AddDate(0, 0, -1),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		Status:      "scheduled",
	}
	if err := config.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("Failed to seed announcement: %v", err)
	}

	// Listing announcements refreshes their statuses, which is what flips a due
	// scheduled announcement to active.
	for attempt := 1; attempt <= 2; attempt++ {
		res, body := doAuthJSON(t, app, http.MethodGet, "/api/admin/platform/announcements", adminToken, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("refresh %d: expected 200, got %d (%v)", attempt, res.StatusCode, body)
		}
	}

	var stored models.Announcement
	if err := config.DB.First(&stored, announcement.ID).Error; err != nil {
		t.Fatalf("Failed to reload announcement: %v", err)
	}
	if stored.Status != "active" {
		t.Fatalf("expected the due announcement to become active, got %q", stored.Status)
	}
	if got := countNotificationsTitled(t, student.RollNo, title); got != 1 {
		t.Fatalf("expected exactly 1 announcement notification after scheduled activation, got %d", got)
	}
}

// If notifications cannot be written, publishing must not half-succeed: the
// announcement stays unpublished, the endpoint reports failure, and the same
// request delivers the announcement once storage recovers.
func TestPublishAnnouncementRetriesAfterNotificationFailure(t *testing.T) {
	app := setupNotificationApp(t)
	resetAnnouncements(t)

	student := seedNotificationStudent(t, "MCA2K22019", "mca019@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin10", "super10@isparc.dev")

	title := "Campus Wi-Fi Maintenance"
	announcement := models.Announcement{
		Title:       title,
		Description: "Wi-Fi will be down on Sunday morning.",
		Category:    "General",
		Audience:    "Students",
		Priority:    "Medium",
		PublishDate: time.Now().AddDate(0, 0, -1),
		ExpiryDate:  time.Now().AddDate(0, 0, 30),
		Status:      "draft",
	}
	if err := config.DB.Create(&announcement).Error; err != nil {
		t.Fatalf("Failed to seed announcement: %v", err)
	}

	path := fmt.Sprintf("/api/admin/platform/announcements/%d/publish", announcement.ID)

	restoreNotifications := breakNotificationStorage(t)
	res, _ := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 while notification storage is down, got %d", res.StatusCode)
	}

	var stored models.Announcement
	if err := config.DB.First(&stored, announcement.ID).Error; err != nil {
		t.Fatalf("Failed to reload announcement: %v", err)
	}
	if stored.Status != "draft" {
		t.Fatalf("expected the failed publish to roll back to draft, got %q", stored.Status)
	}
	if stored.NotifiedAt != nil {
		t.Fatalf("expected notified_at to stay unclaimed after a failed publish, got %v", stored.NotifiedAt)
	}

	restoreNotifications()

	res, body := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected the retry to succeed with 200, got %d (%v)", res.StatusCode, body)
	}
	if got := countNotificationsTitled(t, student.RollNo, title); got != 1 {
		t.Fatalf("expected exactly 1 announcement notification after the retry, got %d", got)
	}

	if err := config.DB.First(&stored, announcement.ID).Error; err != nil {
		t.Fatalf("Failed to reload announcement: %v", err)
	}
	if stored.Status != "active" || stored.NotifiedAt == nil {
		t.Fatalf("expected the retry to publish and claim delivery, got status %q notified_at %v", stored.Status, stored.NotifiedAt)
	}
}

// The same guarantee for certificate decisions: an approval that cannot notify
// the student is not recorded, and retrying it both approves and notifies.
func TestApproveCertificateRetriesAfterNotificationFailure(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22020", "mca020@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin11", "super11@isparc.dev")

	cert := models.Certificate{
		StudentRollNo:     student.RollNo,
		ActivityName:      "Robotics Workshop",
		ActivityCategory:  "TECHNICAL",
		ActivityDate:      time.Now(),
		ParticipationType: "Participant",
		Credits:           15,
		Status:            "Pending",
	}
	if err := config.DB.Create(&cert).Error; err != nil {
		t.Fatalf("Failed to seed certificate: %v", err)
	}

	path := fmt.Sprintf("/api/admin/certificates/%d/approve", cert.ID)

	restoreNotifications := breakNotificationStorage(t)
	res, _ := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 while notification storage is down, got %d", res.StatusCode)
	}

	var stored models.Certificate
	if err := config.DB.First(&stored, cert.ID).Error; err != nil {
		t.Fatalf("Failed to reload certificate: %v", err)
	}
	if stored.Status != "Pending" {
		t.Fatalf("expected the failed approval to roll back to Pending, got %q", stored.Status)
	}

	restoreNotifications()

	res, body := doAuthJSON(t, app, http.MethodPost, path, adminToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected the retry to succeed with 200, got %d (%v)", res.StatusCode, body)
	}

	if err := config.DB.First(&stored, cert.ID).Error; err != nil {
		t.Fatalf("Failed to reload certificate: %v", err)
	}
	if stored.Status != "Approved" {
		t.Fatalf("expected the retry to approve the certificate, got %q", stored.Status)
	}
	if got := countNotificationsTitled(t, student.RollNo, "Certificate Approved"); got != 1 {
		t.Fatalf("expected exactly 1 approval notification after the retry, got %d", got)
	}
}

// The same guarantee for enrolments, reproducing the reported loss: enrolling
// while notification storage is down must not commit the enrolment, because the
// retry would otherwise be rejected as a duplicate and the student would never
// receive the notification.
func TestEnrollActivityRetriesAfterNotificationFailure(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22021", "mca021@isparc.dev")
	token := studentAuthToken(t, student.RollNo, student.EmailID)

	activity := models.Activity{Name: "Robotics Lab", Category: "TECHNICAL", Credits: 10, Mode: "Offline"}
	if err := config.DB.Create(&activity).Error; err != nil {
		t.Fatalf("Failed to seed activity: %v", err)
	}

	path := fmt.Sprintf("/api/student/activities/%d/enroll", activity.ID)

	restoreNotifications := breakNotificationStorage(t)
	res, _ := doAuthJSON(t, app, http.MethodPost, path, token, nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 while notification storage is down, got %d", res.StatusCode)
	}

	// The enrolment must not survive the failed notification, otherwise the
	// retry below returns 409 and the notification is lost for good.
	var count int64
	if err := config.DB.Model(&models.Enrollment{}).
		Where("student_roll_no = ? AND activity_id = ?", student.RollNo, activity.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("Failed to count enrollments: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected the failed enrolment to roll back, found %d enrollment(s)", count)
	}

	restoreNotifications()

	res, body := doAuthJSON(t, app, http.MethodPost, path, token, nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected the retry to succeed with 201, got %d (%v)", res.StatusCode, body)
	}

	if err := config.DB.Model(&models.Enrollment{}).
		Where("student_roll_no = ? AND activity_id = ?", student.RollNo, activity.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("Failed to count enrollments: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 enrollment after the retry, got %d", count)
	}
	if got := countNotificationsTitled(t, student.RollNo, "Activity Enrollment"); got != 1 {
		t.Fatalf("expected exactly 1 enrollment notification after the retry, got %d", got)
	}
}

// An admin notice must not report success when its notification was not
// persisted, and retrying once storage recovers must deliver exactly one.
func TestSendNoticeRetriesAfterNotificationFailure(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22022", "mca022@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin12", "super12@isparc.dev")

	path := "/api/admin/students/" + student.RollNo + "/notice"
	payload := map[string]any{"message": "Please submit your pending certificate."}

	restoreNotifications := breakNotificationStorage(t)
	res, _ := doAuthJSON(t, app, http.MethodPost, path, adminToken, payload)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 while notification storage is down, got %d", res.StatusCode)
	}

	restoreNotifications()

	res, body := doAuthJSON(t, app, http.MethodPost, path, adminToken, payload)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected the retry to succeed with 200, got %d (%v)", res.StatusCode, body)
	}
	if got := countNotificationsTitled(t, student.RollNo, "New Notice"); got != 1 {
		t.Fatalf("expected exactly 1 notice notification after the retry, got %d", got)
	}
}

// The endpoint the portal's bell reads must be the single place a student sees
// every kind of event, whatever wrote it. This covers the two sources that used
// to live apart: an activity-monitoring reminder (which also writes an admin
// note) and a platform event notification, both surfacing through one
// GET /api/student/notifications response.
func TestReminderAndEventNotificationsShareOneEndpoint(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22023", "mca023@isparc.dev")
	studentToken := studentAuthToken(t, student.RollNo, student.EmailID)
	adminToken := seedNotificationAdmin(t, "superadmin13", "super13@isparc.dev")

	// CreatePlatformActivity resolves the activity's track, so it has to exist.
	track := models.Track{Name: "Skill Building", Description: "Technical activities.", Status: "Active"}
	if err := config.DB.Create(&track).Error; err != nil {
		t.Fatalf("Failed to seed track: %v", err)
	}

	// Source 1: an activity-monitoring reminder.
	res, body := doAuthJSON(t, app, http.MethodPost, "/api/admin/monitoring/send-reminder", adminToken, map[string]any{
		"student_enrollment": student.EnrollmentNo,
		"activity_name":      "Python Workshop",
		"issue":              "Pending Verification",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from send-reminder, got %d (%v)", res.StatusCode, body)
	}

	// Source 2: a platform activity opening for registration.
	res, body = doAuthJSON(t, app, http.MethodPost, "/api/admin/platform/activities", adminToken, map[string]any{
		"name":     "Annual Hackathon",
		"track":    track.Name,
		"category": "TECHNICAL",
		"status":   "Active",
		"mode":     "Offline",
	})
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		t.Fatalf("expected the activity to be created, got %d (%v)", res.StatusCode, body)
	}

	// Both must be visible through the one endpoint the portal actually calls.
	res, body = doAuthJSON(t, app, http.MethodGet, "/api/student/notifications", studentToken, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from the notifications endpoint, got %d (%v)", res.StatusCode, body)
	}

	list, ok := body["notifications"].([]any)
	if !ok {
		t.Fatalf("expected a notifications array in the response, got %v", body["notifications"])
	}

	titles := map[string]string{}
	for _, item := range list {
		n, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected notification objects, got %T", item)
		}
		titles[n["title"].(string)] = n["message"].(string)
	}

	reminder, sawReminder := titles[models.NotificationTitleActivityReminder]
	if !sawReminder {
		t.Fatalf("expected the monitoring reminder in the notification list, got titles %v", titles)
	}
	if !strings.Contains(reminder, "Python Workshop") {
		t.Errorf("expected the reminder to name the activity, got %q", reminder)
	}

	newActivity, sawActivity := titles[models.NotificationTitleNewActivity]
	if !sawActivity {
		t.Fatalf("expected the new-activity notification in the list, got titles %v", titles)
	}
	if !strings.Contains(newActivity, "Annual Hackathon") {
		t.Errorf("expected the event notification to name the activity, got %q", newActivity)
	}

	if unread := body["unread_count"].(float64); unread != 2 {
		t.Errorf("expected both events to count as unread, got unread_count %v", unread)
	}

	// The reminder still leaves its staff-side audit note, and that note's
	// internal marker must not be what the student reads.
	var noteCount int64
	if err := config.DB.Model(&models.AdminNote{}).
		Where("student_roll_no = ? AND text LIKE ?", student.RollNo, models.AdminNoteReminderPrefix+"%").
		Count(&noteCount).Error; err != nil {
		t.Fatalf("Failed to count admin notes: %v", err)
	}
	if noteCount != 1 {
		t.Errorf("expected the reminder to still record 1 admin note, got %d", noteCount)
	}
	if strings.Contains(reminder, models.AdminNoteReminderPrefix) {
		t.Errorf("expected the note marker to be stripped from the student message, got %q", reminder)
	}
}

// A platform activity that is not yet open for registration notifies nobody.
func TestClosedPlatformActivityNotifiesNobody(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22024", "mca024@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin14", "super14@isparc.dev")

	track := models.Track{Name: "Skill Building", Description: "Technical activities.", Status: "Active"}
	if err := config.DB.Create(&track).Error; err != nil {
		t.Fatalf("Failed to seed track: %v", err)
	}

	res, body := doAuthJSON(t, app, http.MethodPost, "/api/admin/platform/activities", adminToken, map[string]any{
		"name":     "Unannounced Workshop",
		"track":    track.Name,
		"category": "TECHNICAL",
		"status":   "Inactive",
		"mode":     "Offline",
	})
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		t.Fatalf("expected the activity to be created, got %d (%v)", res.StatusCode, body)
	}

	if got := countNotificationsTitled(t, student.RollNo, models.NotificationTitleNewActivity); got != 0 {
		t.Fatalf("expected a closed activity to notify nobody, got %d notification(s)", got)
	}
}

// An activity whose notification fan-out fails is not created, so the super
// admin can retry it rather than leaving a live activity nobody was told about.
func TestPlatformActivityRetriesAfterNotificationFailure(t *testing.T) {
	app := setupNotificationApp(t)
	student := seedNotificationStudent(t, "MCA2K22025", "mca025@isparc.dev")
	adminToken := seedNotificationAdmin(t, "superadmin15", "super15@isparc.dev")

	track := models.Track{Name: "Skill Building", Description: "Technical activities.", Status: "Active"}
	if err := config.DB.Create(&track).Error; err != nil {
		t.Fatalf("Failed to seed track: %v", err)
	}

	payload := map[string]any{
		"name":     "Robotics Bootcamp",
		"track":    track.Name,
		"category": "TECHNICAL",
		"status":   "Active",
		"mode":     "Offline",
	}

	restoreNotifications := breakNotificationStorage(t)
	res, _ := doAuthJSON(t, app, http.MethodPost, "/api/admin/platform/activities", adminToken, payload)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 while notification storage is down, got %d", res.StatusCode)
	}

	var activityCount int64
	if err := config.DB.Model(&models.Activity{}).Where("name = ?", "Robotics Bootcamp").Count(&activityCount).Error; err != nil {
		t.Fatalf("Failed to count activities: %v", err)
	}
	if activityCount != 0 {
		t.Fatalf("expected the failed creation to roll back, found %d activity/activities", activityCount)
	}

	restoreNotifications()

	res, body := doAuthJSON(t, app, http.MethodPost, "/api/admin/platform/activities", adminToken, payload)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		t.Fatalf("expected the retry to create the activity, got %d (%v)", res.StatusCode, body)
	}
	if got := countNotificationsTitled(t, student.RollNo, models.NotificationTitleNewActivity); got != 1 {
		t.Fatalf("expected exactly 1 new-activity notification after the retry, got %d", got)
	}
}
