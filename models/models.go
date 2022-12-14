package models

import "gorm.io/gorm"

type World struct {
	gorm.Model
	Name    string
	K8sName string
	Port    int
	Tags    []Tag
	Volumes []Volume
}

type Tag struct {
	gorm.Model
	WorldId uint
	Key     string
	Value   string
}

type Volume struct {
	gorm.Model
	WorldId   uint
	HostPath  string
	LocalPath string
	Storage   string
	Claim     string
}
