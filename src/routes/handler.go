package routes

import (
	"bash06/goimg/src/worker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	R   *gin.Engine
	Db  *gorm.DB
	W   *worker.Worker
	Log *zap.Logger
}

func InitHandler(r *gin.Engine, db *gorm.DB, w *worker.Worker, l *zap.Logger) {
	h := &Handler{
		R:   r,
		Db:  db,
		W:   w,
		Log: l,
	}

	public := r.Group("/")
	{
		public.POST("/upload", h.upload)
	}
}
