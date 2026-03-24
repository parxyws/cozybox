package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/handlers"
)

func UserRoute(route *gin.RouterGroup, handler handlers.UserHandler) {
	user := route.Group("/user")
	{
		user.POST("/update-email")
	}
}
