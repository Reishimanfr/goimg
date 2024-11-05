package routes

import (
	"bash06/goimg/src/files"
	"bash06/goimg/src/util"
	"bash06/goimg/src/worker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	Router   *gin.Engine
	Db       *gorm.DB
	Worker   *worker.Worker
	Log      *zap.Logger
	Argon    *util.Argon2idHash
	FManager *files.FileManager
}

func InitHandler(r *gin.Engine, db *gorm.DB, w *worker.Worker, l *zap.Logger) {
	h := &Handler{
		Router:   r,
		Db:       db,
		Worker:   w,
		Log:      l,
		Argon:    util.NewArgon2idHash(1, 32, 64*1024, 32, 256),
		FManager: files.New(files.ManagerModeOnDisk, db),
	}

	public := r.Group("/")
	{
		public.POST("/upload", h.upload)
	}
}
