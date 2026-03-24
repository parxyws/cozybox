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
	defer func() {
		cancel()
	}()

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

	helper.Success(c, http.StatusOK, "User registered successfully", result)
}

func (a *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer func() {
		cancel()
	}()

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

func (a *AuthHandler) Login(c *gin.Context) {}

func (a *AuthHandler) ForgotPassword(c *gin.Context) {}

func (a *AuthHandler) ResetPassword(c *gin.Context) {}
