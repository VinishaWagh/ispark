package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vinishawagh/ispark/api/controllers"
	"github.com/vinishawagh/ispark/api/middleware"
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
}
