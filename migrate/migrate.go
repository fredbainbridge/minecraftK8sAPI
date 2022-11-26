package main

import (
	"main/initializers"
	"main/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	initializers.DB.AutoMigrate(&models.World{})
	initializers.DB.AutoMigrate(&models.Tag{})
	initializers.DB.AutoMigrate(&models.Volume{})
}
