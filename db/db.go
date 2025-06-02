package db

import (
	"log"
	"os"
	"sdmht-server/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetupDB() {
	dsn, exists := os.LookupEnv("GORM_DSN")
	var dialector gorm.Dialector
	if exists {
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open("db.sqlite3?_pragma=foreign_keys(1)")
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	models.AutoMigrate(db)
	DB = db
}
