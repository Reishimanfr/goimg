package routes

import (
	"reishimanfr/goimg/flags"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
)

type Handler struct {
	Router      *gin.Engine
	Db          *gorm.DB
	Logger      *zap.Logger
	MaxFileSize int64
}

func (h *Handler) Init() {
	h.MaxFileSize = *flags.MaxFileSize << 20

	memoryStore := persist.NewMemoryStore(time.Minute * 2)

	h.Router.GET("/:key", h.serve, cache.CacheByRequestURI(memoryStore, time.Minute*2))
	h.Router.POST("/upload", h.upload)
}
