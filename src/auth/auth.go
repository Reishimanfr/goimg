package auth

import (
	"bash06/goimg/src/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AuthHandler struct {
	Router *gin.Engine
	Db     *gorm.DB
	Log    *zap.Logger
	Argon  *util.Argon2idHash
}

type UserAuth struct {
	Id           uint   `gorm:"primaryKey,unique,autoIncrement"`
	Email        string `gorm:"email"`
	PasswordHash string
	PasswordSalt string
	IsVerified   bool
	Role         string
}

func NewDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
	db.AutoMigrate(&UserAuth{})
	return db, err
}

func InitHandler(r *gin.Engine, db *gorm.DB, l *zap.Logger) {
	h := &AuthHandler{
		Router: r,
		Db:     db,
		Log:    l,
		Argon:  util.NewArgon2idHash(1, 32, 64*1024, 32, 256),
	}

	public := r.Group("/api/v1")
	{
		public.POST("/login", h.login)
		public.POST("/register", nil)
		public.POST("/update", nil)
		public.HEAD("/heartbeat", h.heartbeat)
	}
}
