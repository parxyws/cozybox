package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/handlers"
)

// UserRoute registers user profile management routes (all require authentication).
func UserRoute(route *gin.RouterGroup, handler *handlers.UserHandler) {
	user := route.Group("/user")
	{
		user.GET("/profile", handler.GetProfile)
		user.PUT("/profile", handler.UpdateProfile)
		user.PUT("/password", handler.UpdatePassword)
		user.POST("/update-email", handler.UpdateEmail)
		user.POST("/commit-update-email", handler.CommitUpdateEmail)
		user.DELETE("/account", handler.DeleteAccount)
	}
}
