package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
)

// GetLeaderboard returns the leaderboard sorted by total credits
func GetLeaderboard(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	type LeaderboardEntry struct {
		RollNo     string `json:"roll_no"`
		Name       string `json:"name"`
		CourseName string `json:"course_name"`
		Semester   int    `json:"semester"`
		Points     int    `json:"points"`
		IsSelf     bool   `json:"is_self"`
	}

	var entries []LeaderboardEntry

	err := config.DB.Raw(`
		SELECT 
			s.roll_no, 
			s.name, 
			s.course_name, 
			s.semester, 
			COALESCE(SUM(c.credits), 0) as points
		FROM students s
		LEFT JOIN certificates c ON c.student_roll_no = s.roll_no AND c.status = 'Approved'
		GROUP BY s.roll_no, s.name, s.course_name, s.semester
		ORDER BY points DESC, s.name ASC
	`).Scan(&entries).Error

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
