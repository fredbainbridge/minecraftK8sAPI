package controllers

import (
	"errors"
	"main/initializers"
	"main/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WorldsRead(c *gin.Context) {
	id := c.Param("worldId")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.Status(400)
		return
	}
	var world = models.World{}
	err = initializers.DB.Model(&models.World{}).Preload("Volumes").Preload("Tags").First(&world, idInt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(404)
		c.Error(err)
		return
	}

	c.JSON(200, world)
}

func WorldsReadList(c *gin.Context) {
	var worlds []models.World
	result := initializers.DB.Model(&models.World{}).Preload("Volumes").Preload("Tags").Find(&worlds)
	if result.Error != nil {
		c.Status(400)
		return
	}
	c.JSON(200, worlds)
}
