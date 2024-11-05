package routes

import (
	"bash06/goimg/src/database"
	"bash06/goimg/src/flags"
	"net/http"
	"net/mail"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type UserRegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language"`
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Registering can be configured nicely with flags.
// You can completely disable registering or require users to be verified before being able to actually use your instance
// By default users are able to register new accounts and have to be verified manually from the admin dashboard
// This function will return a JWT token that never expires unless the user decides to regenerate the token
// Admin users are always allowed to create new accounts, although they need access to the user's email
func (h *Handler) register(c *gin.Context) {
	requestId := randomString(10)

	if !*flags.EnableUserRegister {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":      "Registration for new users has been disabled on this server",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	NewUserData := new(UserRegisterInput)

	if err := c.MustBindWith(NewUserData, binding.JSON); err != nil {
		h.Log.Error("Error while binding JSON data to struct", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while parsing the provided data",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if NewUserData.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "No email provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if !validateEmail(NewUserData.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "An invalid email address was provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if NewUserData.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "No password provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	existsEmail := new(database.UserRecord)

	if err := h.Db.Where("email = ?", NewUserData.Email).First(&existsEmail).Error; err != nil {
		h.Log.Error("Error while checking if email is already used", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while creating an account for you",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if existsEmail != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error":      "An account with this email already exists. Login instead?",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	hs, err := h.Argon.GenerateHash([]byte(NewUserData.Password), nil)
	if err != nil {
		h.Log.Error("Error while hashing user password", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while creating your account",
			"status":     "failure",
			"request_id": requestId,
		})
		return
	}

	newUserRecord := &database.UserRecord{
		Email:        NewUserData.Email,
		Role:         "user",
		LastLogin:    0,
		CreatedAt:    time.Now().Unix(),
		Language:     "en",
		IsVerified:   true,
		PasswordHash: string(hs.Hash),
		PasswordSalt: string(hs.Salt),
	}

	if *flags.VerifyNewUsersManually {
		newUserRecord.IsVerified = false
	}

	if err := h.Db.Create(newUserRecord).Error; err != nil {
		h.Log.Error("Failed to create new user record in database", zap.String("requestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while creating your account",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if newUserRecord.IsVerified {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Account created! You're ready to use the service",
			"status":     "success",
			"request_id": requestId,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Account created, but you're not verified yet. Please wait until an administrator verifies your account",
			"status":     "success",
			"request_id": requestId,
		})
	}
}
