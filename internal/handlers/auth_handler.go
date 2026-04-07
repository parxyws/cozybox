package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/service"
	"github.com/parxyws/cozybox/internal/tools/helper"
	"github.com/parxyws/cozybox/internal/tools/validator"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (a *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.RegisterUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.userService.Register(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	helper.Success(c, http.StatusCreated, "User registered successfully", result)
}

func (a *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.VerifyEmailRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.userService.VerifyEmail(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to verify email", err)
		return
	}

	helper.Success(c, http.StatusOK, "User email verified successfully", result)
}

func (a *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.userService.Login(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Login successful", result)
}

func (a *AuthHandler) RefreshToken(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := a.userService.RefreshToken(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusUnauthorized, "Token refresh failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (a *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	sessionID, exists := c.Get("session_id")
	if !exists {
		helper.Error(c, http.StatusBadRequest, "Session not found", nil)
		return
	}

	if err := a.userService.Logout(ctx, sessionID.(string)); err != nil {
		helper.Error(c, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	helper.Success(c, http.StatusOK, "Logged out successfully", nil)
}

func (a *AuthHandler) ForgotPassword(c *gin.Context) {}

func (a *AuthHandler) ResetPassword(c *gin.Context) {}
