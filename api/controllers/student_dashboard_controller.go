package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iips-oss/ispark/api/config"
	"github.com/iips-oss/ispark/api/models"
)

// GetMarksheet gets marksheet distribution and credit summary
func GetMarksheet(c *fiber.Ctx) error {
	rollNo := c.Locals("roll_no").(string)

	// Fetch student details
	var student models.Student
	if err := config.DB.Where("roll_no = ?", rollNo).First(&student).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Student not found",
		})
	}

	// Fetch all approved certificates for credit categories
	var approvedCerts []models.Certificate
	if err := config.DB.Where("student_roll_no = ? AND status = 'Approved'", rollNo).Find(&approvedCerts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch approved credits",
		})
	}

	// Group activities/credits by category
	categoriesMap := map[string]*struct {
		Category     string `json:"category"`
		Activities   int    `json:"activities"`
		Credits      int    `json:"credits"`
		Contribution string `json:"contribution"`
	}{
		"TECHNICAL":       {"Technical Skills", 0, 0, "0%"},
		"PUBLIC SPEAKING": {"Public Speaking", 0, 0, "0%"},
		"RESEARCH":        {"Research", 0, 0, "0%"},
		"SOCIAL SERVICE":  {"Social Service", 0, 0, "0%"},
		"SPORTS":          {"Sports", 0, 0, "0%"},
		"LEADERSHIP":      {"Leadership", 0, 0, "0%"},
		"CULTURAL":        {"Cultural", 0, 0, "0%"},
		"LITERARY":        {"Literary", 0, 0, "0%"},
	}

	totalCredits := 0
	totalActivities := 0

	for _, cert := range approvedCerts {
		catKey := cert.ActivityCategory
		matchKey := ""
		for k := range categoriesMap {
			if k == catKey || k == cert.ActivityCategory {
				matchKey = k
				break
			}
		}

		if matchKey == "" {
			matchKey = catKey
			categoriesMap[matchKey] = &struct {
				Category     string `json:"category"`
				Activities   int    `json:"activities"`
				Credits      int    `json:"credits"`
				Contribution string `json:"contribution"`
			}{Category: catKey, Activities: 0, Credits: 0, Contribution: "0%"}
		}

		categoriesMap[matchKey].Activities++
		categoriesMap[matchKey].Credits += cert.Credits
		totalCredits += cert.Credits
		totalActivities++
	}

	// Convert categories to array and compute percentages
	var categoriesList []interface{}
	for _, data := range categoriesMap {
		if data.Activities > 0 {
			pct := 0.0
			if totalCredits > 0 {
				pct = float64(data.Credits) / float64(totalCredits) * 100
			}
			data.Contribution = fmt.Sprintf("%.0f%%", pct)
			categoriesList = append(categoriesList, data)
		}
	}

	if len(categoriesList) == 0 {
		categoriesList = []interface{}{}
	}

	// Semester contribution summary
	semesterMap := make(map[int]*struct {
		Semester   string `json:"semester"`
		Credits    int    `json:"credits"`
		Activities int    `json:"activities"`
		Cumulative int    `json:"cumulative"`
	})

	for sem := 1; sem <= student.Semester; sem++ {
		roman := getRomanSemester(sem)
		semesterMap[sem] = &struct {
			Semester   string `json:"semester"`
			Credits    int    `json:"credits"`
			Activities int    `json:"activities"`
			Cumulative int    `json:"cumulative"`
		}{Semester: fmt.Sprintf("Semester %s", roman), Credits: 0, Activities: 0, Cumulative: 0}
	}

	for i, cert := range approvedCerts {
		semVal := (i % student.Semester) + 1
		semesterMap[semVal].Activities++
		semesterMap[semVal].Credits += cert.Credits
	}

	var semesterSummary []interface{}
	cumulative := 0
	for sem := 1; sem <= student.Semester; sem++ {
		semData := semesterMap[sem]
		cumulative += semData.Credits
		semData.Cumulative = cumulative
		semesterSummary = append(semesterSummary, semData)
	}

	return c.JSON(fiber.Map{
		"student_info": fiber.Map{
			"name":          student.Name,
			"roll_no":       student.RollNo,
			"enrollment_no": student.EnrollmentNo,
			"semester":      getOrdinalSemester(student.Semester),
			"course":        student.CourseName,
			"department":    "DAVV Department",
			"batch":         fmt.Sprintf("%d – %d", time.Now().Year()-student.Semester/2-1, time.Now().Year()+3),
			"institute":     "IIPS, DAVV Indore",
		},
		"credit_categories": categoriesList,
		"total_activities":  totalActivities,
		"total_credits":     totalCredits,
		"semester_summary":  semesterSummary,
	})
}

// Helpers
func getRomanSemester(sem int) string {
	romans := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}
	if sem >= 1 && sem <= 10 {
		return romans[sem-1]
	}
	return fmt.Sprintf("%d", sem)
}

func getOrdinalSemester(sem int) string {
	ordinals := []string{"First", "Second", "Third", "Fourth", "Fifth", "Sixth", "Seventh", "Eighth", "Ninth", "Tenth"}
	roman := getRomanSemester(sem)
	if sem >= 1 && sem <= 10 {
		return fmt.Sprintf("%s (%s)", roman, ordinals[sem-1])
	}
	return fmt.Sprintf("%s", roman)
}
