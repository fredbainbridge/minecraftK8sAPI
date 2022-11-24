package models

import "gorm.io/gorm"

type Node struct {
	gorm.Model
	Name string
}
