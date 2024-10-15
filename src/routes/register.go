package routes

import "github.com/gin-gonic/gin"

// Registering can be configured nicely with flags.
// You can completely disable registering or require users to be verified before being able to actually use your instance
// By default users are able to register new accounts and have to be verified manually from the admin dashboard
// This function will return a JWT token that never expires unless the user decides to regenerate the token
// Admin users are always allowed to create new accounts, although they need access to the user's email
func (h *Handler) register(c *gin.Context) {

}
