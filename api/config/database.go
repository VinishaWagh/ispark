package config

import (
	"fmt"
	"log"
	"os"

	"github.com/iips-oss/ispark/api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	log.Printf("Connecting to database at %s:%s...", dbHost, dbPort)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// Auto Migration
	log.Println("Running AutoMigration...")
	err = DB.AutoMigrate(&models.Student{}, &models.OTP{}, &models.Admin{}, &models.Activity{}, &models.Certificate{}, &models.Enrollment{}, &models.SystemSetting{}, &models.AdminNote{}, &models.Track{}, &models.Announcement{}, &models.GeneratedReport{}, &models.ScheduledReport{}, &models.ReportAuditLog{}, &models.Notification{}, &models.BatchOverride{})

	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migration completed.")

	// Safe data migration: backfill coordinator_id from matching admin Name if it is empty/null
	var admins []models.Admin
	if err := DB.Find(&admins).Error; err == nil {
		for _, admin := range admins {
			if err := DB.Model(&models.Activity{}).
				Where("coordinator = ? AND (coordinator_id = ? OR coordinator_id IS NULL)", admin.Name, "").
				Update("coordinator_id", admin.AdminID).Error; err != nil {
				log.Printf("Warning: Failed to backfill coordinator_id for admin %s: %v", admin.AdminID, err)
			}
		}
	}

	// Safe data migration: backfill track_id for existing activities if null or 0
	var unassignedCount int64
	if err := DB.Model(&models.Activity{}).Where("track_id IS NULL OR track_id = 0").Count(&unassignedCount).Error; err == nil && unassignedCount > 0 {
		var skillTrack, personalityTrack models.Track
		DB.Where("LOWER(name) = ?", "skill building").FirstOrCreate(&skillTrack, models.Track{
			Name: "Skill Building", Description: "Technical and vocational activities that develop practical competencies.", Status: "Active",
		})
		DB.Where("LOWER(name) = ?", "personality development").FirstOrCreate(&personalityTrack, models.Track{
			Name: "Personality Development", Description: "Activities focused on personal growth, communication, and leadership skills.", Status: "Active",
		})

		if skillTrack.ID != 0 {
			DB.Model(&models.Activity{}).
				Where("(track_id IS NULL OR track_id = 0) AND UPPER(category) IN ?", []string{"TECHNICAL", "RESEARCH", "SPORTS", "CULTURAL"}).
				Update("track_id", skillTrack.ID)
		}
		if personalityTrack.ID != 0 {
			DB.Model(&models.Activity{}).
				Where("track_id IS NULL OR track_id = 0").
				Update("track_id", personalityTrack.ID)
		}
	}

	// Safe data migration: existing verified students predate the status column
	// (whose default is "Pending"), so mark them Active. This keeps the super
	// admin registry from reporting already-active accounts as Pending.
	DB.Model(&models.Student{}).
		Where("is_verified = ? AND status = ?", true, "Pending").
		Update("status", "Active")
}
