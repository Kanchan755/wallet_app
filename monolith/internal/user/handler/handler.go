package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error()))
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"data":   nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})

}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userId := c.GetString("user_id")
	user, err := h.svc.GetProfile(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userId := c.GetString("user_id")
	err := h.svc.DeleteAccount(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"status":  "success",
		"message": "account deleted successfully",
	})
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userId := c.GetString("user_id")

	//get the file from request mutltipart
	file, err := c.FormFile("avatar")
	if err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_REQUEST_BODY", "failed to upload file"))
		return
	}
	defer file.Close()
	// validate file format
	if file.Size > 5*1024*1024 {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_REQUEST_BODY", "file size must be less than 5MB"))
		return
	}
	//validate file content
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_REQUEST_BODY", "invalid file format"))
		return
	}

	//folder for saving
	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, os.ModePerm)

	//rename file based on user id
	filename := fmt.Sprintf("%s%s", userId, ext)
	savePath := filepath.Join(uploadDir, filename)

	//save the file
	if err = c.SaveUploadedFile(file, savePath); err != nil {
		c.Error(customError.NewAppError(http.StatusInternalServerError, "FAILED_TO_SAVE_FILE", err.Error()))
		return
	}

	//update url in database
	avatarURL := "uploads/" + filename
	if err := h.svc.UpdateAvatar(c.Request.Context(), userId, avatarURL); err != nil {
		c.Error(customError.NewAppError(http.StatusInternalServerError, "FAILED_TO_UPDATE_AVATAR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   avatarURL,
	})

}
