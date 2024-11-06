package api

import (
	"bash06/goimg/src/files"
	"bash06/goimg/src/worker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ApiHandler struct {
	Router      *gin.Engine
	Db          *gorm.DB
	Worker      *worker.Worker
	Log         *zap.Logger
	FileManager *files.FileManager
}

func InitHandler(r *gin.Engine, db *gorm.DB, w *worker.Worker, l *zap.Logger) {
	h := &ApiHandler{
		Router:      r,
		Db:          db,
		Worker:      w,
		Log:         l,
		FileManager: files.New(files.ManagerMode.OnDisk, db, l),
	}

	public := r.Group("/api/v1")
	{
		public.POST("/upload", h.upload)
		public.GET("/:key", h.serve)
	}
}
