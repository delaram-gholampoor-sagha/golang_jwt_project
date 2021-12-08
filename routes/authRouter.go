package routes

import (
	controller "github.com/Delaram-Gholampoor-Sagha/golang_jwt_project/controllers"
	"github.com/Delaram-Gholampoor-Sagha/golang_jwt_project/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())

}
