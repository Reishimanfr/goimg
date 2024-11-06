package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented",
	})
}
