package database

import (
	"bufio"
	"fmt"
	"os"
	"reishimanfr/goimg/flags"
	"reishimanfr/goimg/token"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// It's faster do to a database lookup than to do io operations
type FileRecord struct {
	Id        uint   `gorm:"primaryKey,autoIncrement,unique"`
	Filename  string `gorm:"unique,index"`
	CreatedAt int64
	Mimetype  string `gorm:"not null"`
	Size      int64
	Private   bool
}

type Token struct {
	Id      uint   `gorm:"primaryKey,autoIncrement,unique"`
	Token   string `gorm:"unique,index"`
	UserTag string `gorm:"unique"`
}

type User struct {
	Id           uint `gorm:"primaryKey,autoIncrement,unique"`
	PasswordHash string
	PasswordSalt string
	UserTag      string `gorm:"unique,index"`
	Role         string // "admin", "user"
}

func Init() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database.db"))
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&FileRecord{}, &Token{}, &User{})

	// Count existing tokens and if n == 0 generate one and print it in the console
	var tokenCount int64

	if err := db.Table("tokens").Where("1 = 1").Count(&tokenCount).Error; err != nil {
		return nil, err
	}

	if tokenCount == 0 {
		opaqToken, err := token.GenerateOpaque(*flags.TokenSizeBits)
		if err != nil {
			return nil, err
		}

		if err := db.Table("tokens").Create(&Token{Token: opaqToken}).Error; err != nil {
			return nil, err
		}

		fmt.Printf(`
		
		No API access token found, one has been generated for you and will be reveled here.
		You should NEVER share this token with anyone as it allows others to upload and delete
		images from the server. Press ENTER to continue.
		
		`)

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		fmt.Printf("\n		API access token: %v", opaqToken)
	}

	var adminCount int64

	if err := db.Table("users").Where("role = admin").Count(&adminCount).Error; err != nil {
		return nil, err
	}

	// This will only initially check if an administrator account exists. The database should NEVER be manually
	// changed and there will be NO CHECKS FOR ANY INVALID USERS EDITED MANUALLY.
	if adminCount == 0 {
		fmt.Printf("\nNo administrator accounts found, attempting to create a base account...")

		argon := token.NewArgon2idHash(1, 32, 64*1024, 32, 256)

		hs, err := argon.GenerateHash([]byte("admin"), nil)
		if err != nil {
			return nil, err
		}

		if err := db.Table("users").Create(&User{
			UserTag:      "admin",
			Role:         "admin",
			PasswordHash: string(hs.Hash),
			PasswordSalt: string(hs.Salt),
		}).Error; err != nil {
			return nil, err
		}

		opaqToken, err := token.GenerateOpaque(*flags.TokenSizeBits)
		if err != nil {
			return nil, err
		}

		// Possibly unsafe? should check if a user called "admin" exists and if so change the name of the
		// initial account so we don't have a random person get admin permissions
		// TODO: revisit this
		if err := db.Table("tokens").Create(&Token{
			Token:   opaqToken,
			UserTag: "admin",
		}).Error; err != nil {

		}

		fmt.Printf("\nBase account created. Credentials will be shown after you press ENTER.")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		fmt.Printf("\nLogin: admin, Password: admin, Access token: %v", "")
	}

	return db, err
}
