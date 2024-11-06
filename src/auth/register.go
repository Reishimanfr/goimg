package auth

import (
	"bash06/goimg/src/flags"
	"bash06/goimg/src/util"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRegisterBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language"`
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (h *AuthHandler) register(c *gin.Context) {
	requestId := util.RandomString(10)

	if !*flags.EnableUserRegister {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":      "Registration for new users has been disabled on this server",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	var newUser *UserRegisterBody

	if err := c.MustBindWith(&newUser, binding.JSON); err != nil {
		h.Log.Error("Error while binding JSON data to struct", zap.String("RequestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while parsing the provided data",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if strings.Trim(newUser.Email, " ") == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "No email provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if strings.Trim(newUser.Password, "") == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "No password provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	if !validateEmail(newUser.Email) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "An invalid email address was provided",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	var exists bool
	if err := h.Db.Model(&UserAuth{}).Select("count(*) > 0").Where("email = ?", newUser.Email).Find(&exists).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			h.Log.Error("Failed to check if user with email exists", zap.String("RequestId", requestId), zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":      "Something went wrong while processing your request",
				"status":     "failed",
				"request_id": requestId,
			})
			return
		}
	}

	if exists {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error":      "An account with this email already exists",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	hs, err := h.Argon.GenerateHash([]byte(newUser.Password), nil)
	if err != nil {
		h.Log.Error("Failed to hash new user password", zap.String("RequestId", requestId), zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Something went wrong while processing your request",
			"status":     "failed",
			"request_id": requestId,
		})
		return
	}

	newUserRecord := &UserAuth{
		Email:        newUser.Email,
		PasswordHash: string(hs.Hash),
		PasswordSalt: string(hs.Salt),
		IsVerified:   true,
		Role:         "user", // TODO
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
