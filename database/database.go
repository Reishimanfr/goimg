package database

import (
	"fmt"
	"reishimanfr/goimg/flags"
	"reishimanfr/goimg/token"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// It's faster do to a database lookup than to do io operations
type FileRecord struct {
	Id        uint   `gorm:"primaryKey,autoIncrement,unique"`
	Filename  string `gorm:"unique"`
	CreatedAt int64
	Mimetype  string `gorm:"not null"`
	Size      int64
}

type Token struct {
	Id    uint   `gorm:"primaryKey,autoIncrement,unique"`
	Token string `gorm:"unique"`
}

func Init() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database.db"))
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&FileRecord{}, &Token{})

	// Count existing tokens and if n == 0 generate one and print it in the console
	var tokenCount int64

	if err := db.Table("tokens").Where("1 = 1").Count(&tokenCount).Error; err != nil {
		return nil, err
	}

	if tokenCount == 0 {
		opaqToken, err := token.Generate(*flags.TokenSizeBits)
		if err != nil {
			return nil, err
		}

		if err := db.Table("tokens").Create(&Token{Token: opaqToken}).Error; err != nil {
			return nil, err
		}

		fmt.Printf(`

		No API access token found. I've generated one for you. You should NEVER share this token with anyone
		as it allows others to upload and delete images from the server. Your token will be revealed in 15 seconds...
		
		`)

		time.Sleep(time.Second * 15)
		fmt.Printf("\n		API access token: %v\n\n\n", opaqToken)
	}

	return db, err
}
