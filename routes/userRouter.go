package routes

import (
	controller "golang_jwt_project/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleWare.Authenticate)
	incomingRoutes.GET("/users", controller.GetUsers())
}
