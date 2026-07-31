package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/controllers"
	"github.com/iips-oss/ispark/api/middleware"
)

// SetupRoutes configures the endpoints for the API
func SetupRoutes(app *fiber.App) {
	// Base API group
	api := app.Group("/api")

	// Auth group
	auth := api.Group("/auth")

	// Public Auth routes
	auth.Get("/captcha", controllers.GetCaptcha)
	auth.Post("/register", controllers.Register)
	auth.Post("/verify-otp", controllers.VerifyOTP)
	auth.Post("/login", controllers.Login)
	auth.Post("/forgot-password", controllers.ForgotPassword)
	auth.Post("/reset-password", controllers.ResetPassword)
	auth.Post("/refresh", controllers.RefreshToken)

	// Protected routes (Require login)
	auth.Use(middleware.AuthRequired())
	auth.Post("/logout", controllers.Logout)
	auth.Get("/profile", controllers.GetProfile)

	// Student Dashboard routes (Require login)
	student := api.Group("/student")
	student.Use(middleware.AuthRequired())
	student.Get("/certificates", controllers.GetCertificates)
	student.Post("/certificates", controllers.UploadCertificate)
	student.Get("/certificates/:id/file", controllers.DownloadCertificate)
	student.Get("/leaderboard/champions", controllers.GetCategoryChampions)
	student.Get("/leaderboard", controllers.GetLeaderboard)
	student.Get("/activities", controllers.GetActivities)
	student.Put("/profile", controllers.UpdateProfile)
	student.Post("/change-password", controllers.ChangePassword)
	student.Post("/activities/:id/enroll", controllers.EnrollActivity)
	student.Get("/enrollments", controllers.GetEnrollments)
	student.Get("/dashboard/stats", controllers.GetDashboardStats)
	student.Get("/marksheet", controllers.GetMarksheet)

	// Notifications. GetNotifications serves the notifications table; it
	// supersedes the earlier admin-note-backed GetStudentNotifications, which
	// used to be registered on this same path and shadowed the handlers below.
	student.Get("/notifications", controllers.GetNotifications)
	student.Get("/notifications/unread-count", controllers.GetUnreadNotificationCount)
	student.Put("/notifications/read-all", controllers.MarkAllNotificationsRead)
	student.Put("/notifications/:id/read", controllers.MarkNotificationRead)

	// Admin
	api.Post("/admin/auth/login", controllers.AdminLogin)

	admin := api.Group("/admin")

	// Must be logged in AND hold an administrative role. A super admin is not
	// batch-scoped, so it sees every student.
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.RoleRequired("admin", "superadmin"))

	// Must change the password
	admin.Post("/change-password", controllers.AdminChangePassword)

	// Admin Portal Dashboard
	admin.Get("/dashboard/stats", controllers.GetAdminDashboardStats)
	admin.Get("/dashboard/recent-activities", controllers.GetRecentActivities)

	admin.Get("/students", controllers.GetAllStudents)
	admin.Get("/students/:roll", controllers.GetStudentDetail)
	admin.Get("/students/:roll/observations", controllers.GetMentorObservations)
	admin.Post("/students/:roll/observations", controllers.AddMentorObservation)
	admin.Put("/students/:roll/observations/:id", controllers.EditMentorObservation)
	admin.Post("/students/:roll/notice", controllers.SendStudentNotice)
	// Admin profile
	admin.Get("/profile", controllers.GetAdminProfile)
	admin.Put("/profile", controllers.UpdateAdminProfile)
	// Certificates
	admin.Get("/certificates", controllers.GetCertificatesQueue)
	admin.Get("/certificates/queue", controllers.GetCertificatesQueue)
	admin.Get("/certificates/:id/download", controllers.AdminDownloadCertificate)
	admin.Get("/certificates/:id/file", controllers.AdminDownloadCertificate)
	admin.Post("/certificates/:id/approve", controllers.ApproveCertificate)
	admin.Post("/certificates/:id/reject", controllers.RejectCertificate)
	// Batch Analytics
	admin.Get("/batch-analytics/export", controllers.ExportBatchReport)
	admin.Get("/batch-analytics/reports/export", controllers.ExportBatchReport)
	admin.Get("/batch-analytics", controllers.GetBatchAnalyticsOverview)
	admin.Get("/batch-analytics/:batch", controllers.GetBatchDetail)
	admin.Put("/batch-analytics/:batch", controllers.UpdateBatchAnalytics)

	// Activity Monitoring API
	admin.Get("/monitoring/stats", controllers.GetActivityMonitoringStats)
	admin.Get("/monitoring/activities", controllers.GetMonitoredActivities)
	admin.Put("/monitoring/activities/:id", controllers.UpdateMonitoredActivity)
	admin.Get("/monitoring/insights", controllers.GetMonitoringInsights)
	admin.Get("/monitoring/attention-students", controllers.GetStudentsRequiringAttention)
	admin.Post("/monitoring/send-reminder", controllers.SendActivityMonitoringReminder)

	// Platform-wide routes, super admin only
	platform := admin.Group("/platform", middleware.RoleRequired("superadmin"))
	platform.Get("/stats", controllers.GetPlatformStats)
	platform.Get("/users", controllers.GetPlatformUsers)
	platform.Post("/users", controllers.CreatePlatformUser)
	platform.Put("/users/:id", controllers.UpdatePlatformUser)
	platform.Delete("/users/:id", controllers.DeletePlatformUser)
	platform.Get("/activities", controllers.GetPlatformActivities)
	platform.Post("/activities", controllers.CreatePlatformActivity)
	platform.Put("/activities/:id", controllers.UpdatePlatformActivity)
	platform.Delete("/activities/:id", controllers.DeletePlatformActivity)

	// System settings
	platform.Get("/settings", controllers.GetPlatformSettings)
	platform.Put("/settings", controllers.UpdatePlatformSettings)
	platform.Put("/settings/:key", controllers.UpdatePlatformSetting)

	// Reports center
	platform.Get("/reports/summary", controllers.GetReportsSummary)
	platform.Get("/reports/templates", controllers.GetReportTemplates)
	platform.Get("/reports/export/counts", controllers.GetExportCounts)
	platform.Get("/reports/export", controllers.ExportData)
	platform.Get("/reports/audit", controllers.GetReportAuditLog)
	platform.Get("/reports/institutional", controllers.GetInstitutionalOverview)
	platform.Get("/reports/filters", controllers.GetReportFilters)

	// Reports center: scheduled reports
	platform.Get("/reports/scheduled", controllers.GetScheduledReports)
	platform.Post("/reports/scheduled", controllers.CreateScheduledReport)
	platform.Put("/reports/scheduled/:id", controllers.UpdateScheduledReport)
	platform.Delete("/reports/scheduled/:id", controllers.DeleteScheduledReport)

	// Reports center: generated reports
	platform.Get("/reports", controllers.GetGeneratedReports)
	platform.Post("/reports/generate", controllers.GenerateReport)
	platform.Get("/reports/:id", controllers.GetReportDetail)
	platform.Get("/reports/:id/download", controllers.DownloadReport)
	platform.Delete("/reports/:id", controllers.DeleteReport)

	// Announcement management
	platform.Get("/announcements/stats", controllers.GetAnnouncementStats)
	platform.Get("/announcements", controllers.GetAnnouncements)
	platform.Post("/announcements", controllers.CreateAnnouncement)
	platform.Get("/announcements/:id", controllers.GetAnnouncement)
	platform.Put("/announcements/:id", controllers.UpdateAnnouncement)
	platform.Delete("/announcements/:id", controllers.DeleteAnnouncement)
	platform.Post("/announcements/:id/publish", controllers.PublishAnnouncement)

}
