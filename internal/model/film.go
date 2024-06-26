package model

import (
	"gorm.io/gorm"
	"time"
)

type Film struct {
	Url       string    `json:"url" gorm:"index"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Source    string    `json:"source"`
	Actor     string    `json:"actor"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type MagnetCache struct {
	Magnet string `json:"magnet" gorm:"index"`
	FileId string `json:"file_id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
}

type Actor struct {
	gorm.Model
	Dir  string `json:"dir" gorm:"index"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type VirtualFile struct {
	StorageId       uint          `json:"storage_id"`
	Name            string        `json:"name"`
	ShareID         string        `json:"shareId"`
	SharePwd        string        `json:"sharePwd"`
	ParentDir       string        `json:"parentDir"`
	AppendSubFolder bool          `json:"appendSubFolder"`
	ExcludeUnMatch  bool          `json:"excludeUnMatch"`
	SourceName      string        `json:"sourceName"`
	MinFileSize     int64         `json:"minFileSize"`
	Modified        time.Time     `json:"modified"`
	Replace         []ReplaceItem `json:"replace" gorm:"type:json;serializer:json"`
}

type ReplaceItem struct {
	Start      int    `json:"start"`
	End        int    `json:"end"`
	StartNum   int    `json:"startNum"`
	SourceName string `json:"sourceName"`
}

type Replacement struct {
	StorageId uint   `json:"storage_id"`
	DirName   string `json:"dir_name"`
	OldName   string `json:"old_name"`
	NewName   string `json:"new_name"`
}
