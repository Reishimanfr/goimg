package routes

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reishimanfr/goimg/database"
	"reishimanfr/goimg/flags"
	"slices"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	"gorm.io/gorm"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStr(n int) string {
	if n < 1 {
		return ""
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, n)

	for i := range n {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func (h *Handler) upload(c *gin.Context) {
	requestId := randStr(10)
	token := c.GetHeader("Authorization")

	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "No authorization token provided",
			"request_id": requestId,
		})
		return
	}

	var tokenRecord *database.Token

	if err := h.Db.Table("tokens").Where("token = ?", token).First(&tokenRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid token provided",
				"request_id": requestId,
			})
			return
		}

		h.Logger.Error("Error while looking for opaque token", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.MaxFileSize)

	err := c.Request.ParseMultipartForm(8 << 20)
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":      "The provided file is too big",
				"request_id": requestId,
			})
			return
		}

		h.Logger.Error("Error while parsing multipart form", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.Logger.Error("Error while reading multipart form file", zap.String("RequestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	srcFile, err := file.Open()
	if err != nil {
		h.Logger.Error("Error while opening source file", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	defer srcFile.Close()

	b := make([]byte, 512)
	srcFile.Read(b)
	srcFile.Seek(0, 0)

	mime := mimetype.Detect(b).String()

	if !slices.Contains(flags.AllowedFileTypes, mime) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      mime + " files are not allowed",
			"request_id": requestId,
		})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := randStr(10) + ext

	tx := h.Db.Begin()

	if err := tx.Table("file_records").Create(&database.FileRecord{
		Filename:  filename,
		CreatedAt: time.Now().Unix(),
		Size:      file.Size,
		Mimetype:  mime,
	}).Error; err != nil {
		tx.Rollback()

		h.Logger.Error("Error while creating database record", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.Logger.Error("Error while committing transaction", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	newFilePth := filepath.Join(flags.BasePth, "files", filename)

	dest, err := os.Create(newFilePth)
	if err != nil {
		h.Logger.Error("Error while creating destination file", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	defer dest.Close()

	writer := bufio.NewWriter(dest)
	defer writer.Flush()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		h.Logger.Error("Error while copying file contents", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"request_id": requestId,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key": filename,
	})
}
