package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) heartbeat(c *gin.Context) { c.Status(http.StatusOK) }
