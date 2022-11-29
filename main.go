package main

import (
	"main/controllers"
	"main/initializers"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SetKubeConfig()
}
func main() {
	r := gin.Default()

	r.POST("/worlds", controllers.WorldsCreate)
	r.GET("/worlds/:worldId", controllers.WorldsRead)
	r.GET("/worlds", controllers.WorldsReadList)
	r.DELETE("/worlds/:worldId", controllers.WorldsDelete)

	r.Run()
}
