package routes

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reishimanfr/goimg/database"
	"reishimanfr/goimg/flags"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (h *Handler) serve(c *gin.Context) {
	key := c.Param("key")

	if key == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "No key provided",
		})
		return
	}

	// Browser request for favicon
	if key == "favicon.ico" {
		return
	}

	var fileRecord *database.FileRecord

	if err := h.Db.Table("file_records").Where("filename = ?", key).First(&fileRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}

		h.Logger.Error("Error while querying database", zap.String("Key", key), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Something went wrong while processing your request",
		})
		return
	}

	filePath := filepath.Join(flags.BasePth, "files", fileRecord.Filename)

	file, err := os.Open(filePath)
	if err != nil {
		h.Logger.Error("Error while reading source file", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Something went wrong while reading the file",
		})
		return
	}

	sizeStr := strconv.FormatInt(fileRecord.Size, 10)

	c.Writer.Header().Set("Content-Type", fileRecord.Mimetype)
	c.Writer.Header().Set("Content-Length", sizeStr)

	// if strings.HasPrefix(fileRecord.Mimetype, "video/") {
	buffer := make([]byte, 64*1024)

	if _, err := io.CopyBuffer(c.Writer, file, buffer); err != nil {
		h.Logger.Error("Error while streaming video", zap.String("Filename", fileRecord.Filename), zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	// }
	// } else {
	// 	fileBytes := make([]byte, fileRecord.Size)
	// 	_, err := file.Read(fileBytes)
	// 	if err != nil && err != io.EOF {
	// 		h.Logger.Error("Error while reading file into buffer", zap.String("Filename", fileRecord.Filename), zap.Error(err))
	// 		c.Status(http.StatusInternalServerError)
	// 		return
	// 	}

	// 	c.Data(http.StatusOK, fileRecord.Mimetype, fileBytes)
	// }
}
