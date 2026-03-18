package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
}
