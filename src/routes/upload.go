package routes

import (
	"bash06/goimg/src/database"
	"bash06/goimg/src/flags"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	themLettersNShit = "abcdefghijklmnoprstuwxyzABCDEFGHIJKLMNOPRSTUWXYZ1234567890"
)

func randomString(n int) string {
	if n < 1 {
		return ""
	}

	name := ""

	for range n {
		name += string(themLettersNShit[rand.Intn(len(themLettersNShit))])
	}

	return name
}

// Uploads a new, singular file to the instance.
// Depending on how you configured the flags unregistered users have different permissions (or can't use the service)
func (h *Handler) upload(c *gin.Context) {
	// This is useful in case someone encounters a 500 error. The user would report the request ID which would make reading the specific logs easier to see what went wrong
	requestId := randomString(10)

	tokenString := c.GetHeader("Authorization")

	if tokenString == "" && !*flags.AllowGuestUploads {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "Guest file uploads are disabled on this server. Please login to upload files. (No token provided)",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	userId := ""

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
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
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId = claims["user_id"].(string)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":      "Invalid token provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	var user *database.UserRecord

	if err := h.Db.Where("user_id = ?", userId).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":      "User with this ID doesn't exist. Please go to the registration page and make a new account",
				"status":     "failed",
				"request_id": requestId,
			})
			return
		}

		h.Log.Error("Error while looking up user by ID", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while saving your file",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	maxFileSize := *flags.MaxFileSize << 20

	if tokenString == "" && *flags.MaxGuestFileSize > 0 {
		maxFileSize = *flags.MaxGuestFileSize << 20
	}

	bodySizeUnsafe, _ := strconv.Atoi(c.Request.Header.Get("Content-Length"))

	h.Log.Debug("New file is being uploaded...", zap.String("requestId", requestId), zap.Int("Unsafe body size", bodySizeUnsafe))

	// If someone were to claim a 5MiB file is actually 200000MiB I can't be bothered to check if it's true so I might as well use the god forsaken Content-Length header
	// Why would you lie about your file being larger than it actually is
	if bodySizeUnsafe > int(maxFileSize) {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":      "Entity too large",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	// ...We check again if the file that's supposedly smaller than the limit actually enforces the limit. If not we ofc return a 413
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxFileSize))
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
	// If I'm straight up wrong please make a PR with the corrected code so I don't need another dep
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

	fileRecord := &database.FileRecord{
		CreatedAt: time.Now().Unix(),
	}

	// People who don't have an account have their files deleted after X hours (configured with flags)
	// This can be disabled with flags, but it's on by default
	if tokenString == "" {
		fileRecord.OwnerId = ""

		if *flags.EnableGuestFileDeletion {
			duration := time.Duration(*flags.GuestFileDeletionTime) * time.Hour
			future := time.Now().Add(duration)

			fileRecord.ExpiresAt = future.Unix()
		}
	} else {
		// TODO: look up the person that uploaded the file n attach stuff
	}

	if err := h.Db.Create(fileRecord).Error; err != nil {
		h.Log.Error("Failed to create database record", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while saving your file",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	randomFilename := randomString(10)

	s := &s3.PutObjectInput{
		ACL:         aws.String("public-read"),
		Body:        file,
		Bucket:      aws.String(*flags.OvhContainerName),
		Key:         aws.String(randomFilename),
		ContentType: aws.String(mime),
	}

	if _, err := h.W.Svc.PutObject(s); err != nil {
		h.Log.Error("Failed to upload files to S3", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while saving your file",
			"status":     "failed",
			"request_id": requestId,
		})

		return
	}

	endpointWithoutHttps := strings.TrimPrefix(*flags.OvhEndpoint, "https://")

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("https://%s.%s%s", *flags.OvhContainerName, endpointWithoutHttps, randomFilename),
	})
}
