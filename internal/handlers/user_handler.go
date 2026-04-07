package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/service"
	"github.com/parxyws/cozybox/internal/tools/helper"
	"github.com/parxyws/cozybox/internal/tools/validator"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	result, err := h.userService.GetProfile(ctx)
	if err != nil {
		helper.Error(c, http.StatusNotFound, "User not found", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile retrieved", result)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.userService.UpdateProfile(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	helper.Success(c, http.StatusOK, "Profile updated", result)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.userService.UpdatePassword(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update password", err)
		return
	}

	helper.Success(c, http.StatusOK, "Password updated", result)
}

func (h *UserHandler) UpdateEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.UpdateEmailRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.userService.UpdateEmail(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to initiate email update", err)
		return
	}

	helper.Success(c, http.StatusOK, "Verification email sent", result)
}

func (h *UserHandler) CommitUpdateEmail(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.CommitUpdateEmailRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.userService.CommitUpdateEmail(ctx, &req)
	if err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to update email", err)
		return
	}

	helper.Success(c, http.StatusOK, "Email updated", result)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	ctx, cancel := helper.GetContext(c)
	defer cancel()

	var req dto.DeleteAccountRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := validator.Validate.StructCtx(ctx, req); err != nil {
		helper.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.userService.DeleteAccount(ctx, &req); err != nil {
		helper.Error(c, http.StatusInternalServerError, "Failed to delete account", err)
		return
	}

	helper.Success(c, http.StatusOK, "Account deleted", nil)
}
