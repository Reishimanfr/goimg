package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type FileRecord struct {
	Id        uint   `gorm:"primaryKey,unique,autoIncrement"`
	CreatedAt int64  `json:"created_at"`
	OwnerId   string `json:"owner"`
	ExpiresAt int64  `json:"expires_at"`
	MimeType  string `json:"mime_type"` // Could be useful for making thumbnails
}

type UserRecord struct {
	Id           string `gorm:"primaryKey,unique,autoIncrement"`
	Email        string `gorm:"unique"`
	PasswordHash string
	PasswordSalt string
	IsVerified   bool
	Role         string // User, admin, etc
	CreatedAt    int64
	LastLogin    int64
	Language     string // Used to display stuff in different languages
}

func New() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

	db.AutoMigrate(&FileRecord{}, &UserRecord{})

	return db, err
}
