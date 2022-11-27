package controllers

import (
	"main/initializers"
	"main/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WorldsDelete(c *gin.Context) {
	id := c.Param("worldId")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.Status(400)
		return
	}
	var world = models.World{}
	var result *gorm.DB
	result = initializers.DB.Model(&models.World{}).Preload("Volumes").First(&world, idInt)
	if result.Error != nil {
		c.Status(400)
		return
	}
	if len(world.Volumes) == 1 {
		nukeAll(world.K8sName, world.Volumes[0].Claim, world.Volumes[0].LocalPath)
	}
	result = initializers.DB.Delete(&world)
	if result.Error != nil {
		c.Status(400)
		return
	}

}
