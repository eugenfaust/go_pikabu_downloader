package main

import (
	"gorm.io/gorm"
)
type Story struct {
	gorm.Model
	PikabuID int `gorm:"unique"`
	Title  string
	Url    string `gorm:"unique"`
	Images []Image
	Videos []Video
}

type Image struct {
	gorm.Model
	StoryID uint
	Url, FileId     string
}

type Video struct {
	gorm.Model
	StoryID  uint
	Url, FileId      string
	Duration int
}

var db *gorm.DB