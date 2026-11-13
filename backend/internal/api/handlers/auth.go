package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/lambra/internal/api/middleware"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/service"
	"github.com/yourusername/lambra/pkg/response"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login authenticates a user and returns a JWT
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	result, err := h.service.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, result, "Login successful")
}

// Me returns the authenticated user from the token
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "Not authenticated")
		return
	}

	response.Success(c, gin.H{
		"uuid":     user.UUID,
		"username": user.Username,
		"role":     user.Role,
	}, "User retrieved successfully")
}
