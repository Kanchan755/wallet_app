package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/user/model"
	"github.com/kanchan755/wallet_app/monolith/internal/user/service"
)

type UserHandler struct {
	// Define any dependencies or services needed for handling user-related requests
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

// Implement methods to handle user-related requests, such as registration, profile retrieval, etc.

func (h *UserHandler) RegisterUser(c *gin.Context) {
	// Implement logic to handle user registration request
	// For example, you can parse the request body, call the service layer, and return a response
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		//Register the error input to gin context, so that it can be handled by the error handler middleware
		c.Error(customError.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		c.Error(customError.NewAppError(http.StatusInternalServerError, "Failed to register user", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// Implement logic to handle user profile retrieval request
	id := c.Param("id")
	user, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.Error(customError.NewAppError(http.StatusInternalServerError, "Failed to retrieve user profile", err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	// Implement logic to handle user profile update request
	id := c.Param("id")
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
