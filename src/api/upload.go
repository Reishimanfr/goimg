package api

import (
	"bash06/goimg/src/database"
	"bash06/goimg/src/files"
	"bash06/goimg/src/flags"
	"bash06/goimg/src/util"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileUploadOptions struct {
	ExpiresAt   int64
	CreatedAt   int64
	MaxFileSize int64
	OwnerId     string
}

func prepGuestUpload(c *gin.Context, requestId string) *FileUploadOptions {
	if !*flags.AllowGuestUploads {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "Guest file uploads are disabled on this server. Please login to upload files. (No token provided)",
			"status":     "failed",
			"request_id": requestId,
		})
		return nil
	}

	maxFileSize := *flags.MaxFileSize << 20

	if *flags.MaxGuestFileSize > 0 {
		maxFileSize = *flags.MaxGuestFileSize << 20
	}

	expiresAt := int64(0)

	if *flags.EnableGuestFileDeletion {
		expiresAt = time.Now().Add(time.Second * time.Duration(*flags.GuestFileDeletionTime)).Unix()
	}

	return &FileUploadOptions{
		CreatedAt:   time.Now().Unix(),
		OwnerId:     "",
		ExpiresAt:   expiresAt,
		MaxFileSize: int64(maxFileSize),
	}
}

func prepUserUpload(c *gin.Context, h *ApiHandler, requestId string) *FileUploadOptions {
	tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	userId := ""

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(*flags.HMACSecretKey), nil
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "Invalid token provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId = claims["user_id"].(string)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "Invalid token provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return nil
	}

	var user *database.UserRecord

	if err := h.Db.Where("uuid = ?", userId).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":      "User with this ID doesn't exist. Please go to the registration page and make a new account",
				"status":     "failed",
				"request_id": requestId,
			})
			return nil
		}

		h.Log.Error("Error while looking up user by ID", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while saving your file",
			"status":     "failed",
			"request_id": requestId,
		})
		return nil
	}

	maxFileSize := *flags.MaxFileSize << 20

	// TODO: allow users to make files expire after some time

	return &FileUploadOptions{
		CreatedAt:   time.Now().Unix(),
		MaxFileSize: int64(maxFileSize),
		OwnerId:     userId,
	}
}

// Uploads a new, singular file to the instance.
// Depending on how you configured the flags unregistered users have different permissions (or can't use the service)
func (h *ApiHandler) upload(c *gin.Context) {
	// This is useful in case someone encounters a 500 error. The user would report the request ID which would make reading the specific logs easier to see what went wrong
	requestId := util.RandomString(10)
	tokenString := c.GetHeader("Authorization")

	var uploadOptions *FileUploadOptions

	if tokenString == "" {
		uploadOptions = prepGuestUpload(c, requestId)
	} else {
		uploadOptions = prepUserUpload(c, h, requestId)
	}

	// That means the request was aborted at some point
	if uploadOptions == nil {
		return
	}

	bodySizeUnsafe, _ := strconv.ParseInt(c.Request.Header.Get("Content-Length"), 10, 64)

	// If someone were to claim a 5MiB file is actually 200000MiB I can't be bothered to check if it's true so I might as well use the god forsaken Content-Length header
	// Why would you lie about your file being larger than it actually is
	if bodySizeUnsafe > uploadOptions.MaxFileSize {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":      "Entity too large",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	// ...We check again if the file that's supposedly smaller than the limit actually enforces the limit. If not we ofc return a 413
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, uploadOptions.MaxFileSize)
	fileHeader, err := c.FormFile("file")

	if err != nil {
		if err.Error() == "http: request body too large" {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":      "Entity too large",
				"status":     "failed",
				"request_id": requestId,
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while parsing the provided file",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if fileHeader == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "No files to upload provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	// We can't really trust the mime type since someone could tamper with it...
	// ...just like with the content-length header
	// Imagine how easy life would be if we could trust these things
	file, _ := fileHeader.Open()
	defer file.Close()

	b := make([]byte, 512)

	file.Read(b)

	// The std library detects mp4 files as application/octet-stream....
	// If I'm wrong please make a PR with the corrected code so I don't need another dep
	mime := mimetype.Detect(b).String()

	// Reset the file pointer
	file.Seek(0, 0)

	if !slices.Contains(flags.AllowedFileTypes, mime) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      mime + " files are not allowed",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	fileInfo := &files.FileInfo{
		OwnerId:   uploadOptions.OwnerId,
		Header:    *fileHeader,
		MimeType:  mime,
		ExpiresAt: uploadOptions.ExpiresAt,
	}

	err = h.FileManager.Save(fileInfo)
	if err != nil {
		h.Log.Error("File manager failed to save file", zap.String("RequestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}
}
