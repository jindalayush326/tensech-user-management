package routes

import (
	"github.com/gin-gonic/gin"

	"tensechassignment/internal/auth"
	"tensechassignment/internal/controller"
	"tensechassignment/internal/middleware"
)

func Setup(
	router *gin.Engine,
	validator *auth.Validator,
	userController *controller.UserController,
	csvController *controller.CSVController,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	api := router.Group("/api/v1")
	api.Use(middleware.Auth(validator))

	users := api.Group("/users")

	users.GET("", userController.ListUsers)
	users.GET("/:id", userController.GetUser)

	users.POST(
		"",
		middleware.RequireAdmin(),
		userController.CreateUser,
	)

	users.POST(
		"/bulk-upload",
		middleware.RequireAdmin(),
		csvController.UploadCSV,
	)
}