package api

import (
	"bash06/goimg/src/database"
	"errors"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Serves a file based on the provided key. If the file is set to be private this will do nothing
func (h *ApiHandler) serve(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Abort()
		return
	}

	var recordLocation string
	// TODO: maybe return forbidden if the file is private? idk
	// TODO: disallow guest uploads to be private since they don't belong to anyone
	if err := h.Db.Model(&database.FileRecord{}).Where("filename = ? AND private != 1", key).Select("location").First(&recordLocation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Abort()
			return
		}

		h.Log.Error("Failed to query database", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Something went wrong while processing your request",
		})
		return
	}

	if _, err := os.Stat(recordLocation); errors.Is(err, fs.ErrNotExist) {

	}

	http.ServeFile(c.Writer, c.Request, recordLocation)
}
