package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type FileRecord struct {
	Id           uint   `gorm:"primaryKey,unique,autoIncrement"`
	CreatedAt    int64  `json:"created_at"`
	OwnerId      string `json:"owner_id"`
	ExpiresAt    int64  `json:"expires_at"`
	MimeType     string `json:"mime_type"` // Could be useful for making thumbnails
	Location     string `json:"location"`  // Either the path or URL to the file
	Filename     string `json:"filename"`
	UploadMethod string `json:"upload_method"`
	Private      bool   `json:"private"`
}

type UserRecord struct {
	Id        string `gorm:"primaryKey,unique,autoIncrement"`
	UUID      string `gorm:"unique"`
	CreatedAt int64
	LastLogin int64
	Language  string // Used to display stuff in different languages
}

func NewAPIDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	db.AutoMigrate(&FileRecord{}, &UserRecord{})
	return db, err
}
