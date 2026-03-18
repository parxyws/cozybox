package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/service"
)

type AuthRoute struct {
	userService *service.UserService
}

func NewAuthRoute(userService *service.UserService) *AuthRoute {
	return &AuthRoute{userService: userService}
}

func (a *AuthRoute) Register(c *gin.Context) {

}
