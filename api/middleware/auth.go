package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/vinishawagh/ispark/api/utils"
)

// AuthRequired is a middleware that enforces authentication via JWT
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string

		// 1. Try to get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// 2. Try to get token from cookie if header was empty (convenient for frontends)
		if tokenStr == "" {
			tokenStr = c.Cookies("access_token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication token is missing",
			})
		}

		// Validate token
		claims, err := utils.ValidateAccessToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store claims in context locals
		c.Locals("roll_no", claims.RollNo)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// RoleRequired is a middleware helper that enforces user roles (e.g., student, admin, superadmin)
func RoleRequired(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal := c.Locals("role")
		if roleVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		userRole := roleVal.(string)
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: you do not have permission to access this resource",
			})
		}

		return c.Next()
	}
}
