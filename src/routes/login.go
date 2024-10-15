package routes

import "github.com/gin-gonic/gin"

// Allows existing users to login
// This function returns a JWT token that never expires that'll be saved into the browser's local storage
// The JWT token will also be available to do stuff like shareX integrations or uploading files via HTTP POST requests in whatever way the user wants to make such requests
func (h *Handler) login(c *gin.Context) {

}
