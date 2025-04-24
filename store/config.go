package store

import (
	"log/slog"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/playsthisgame/binq/types"
)

func Setup() (*gorm.DB, error) {
	// create .store if not exists
	path := ".store"
	_ = os.MkdirAll(path, os.ModePerm)

	// init db
	db, err := gorm.Open(sqlite.Open(".store/binq.db"), &gorm.Config{})
	if err != nil {
		slog.Error("Error initializing sqlite", "error", err)
		return nil, err
	}

	// get raw SQL DB connection
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Error getting raw DB", "error", err)
		return nil, err
	}

	// set WAL journal mode
	_, err = sqlDB.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("Error setting WAL mode", "error", err)
	}

	// automigrate db
	db.AutoMigrate(&types.Message{})
	db.AutoMigrate(&types.Queue{})

	return db, nil
}

// TODO: you should add a cleanup for records that have been in the queue for too long, make it configurable
// Schedule a cleanup of deleted records every day at midnight
func ScheduleCleanup() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		if now.After(next) {
			// If current time is past today's midnight, schedule for tomorrow
			next = next.Add(24 * time.Hour)
		}

		duration := next.Sub(now)
		time.Sleep(duration)

		performCleanup()
	}
}

// performCleanup deletes expired records
func performCleanup() {
	db, err := gorm.Open(sqlite.Open(".store/binq.db"), &gorm.Config{})
	if err != nil {
		slog.Error("Error opening DB for cleanup", "error", err)
		return
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	slog.Info("Starting cleanup", "time", time.Now())

	// Delete records older than 1 day
	result := db.Unscoped().
		Where("deleted_at < ?", time.Now().AddDate(0, 0, -1)).
		Delete(&types.Message{})

	if result.Error != nil {
		slog.Error("Error during cleanup", "Error", result.Error)
		return
	}

	slog.Info("Deleted expired records", "count", result.RowsAffected)
	slog.Info("Cleanup finished", "time", time.Now())
}
