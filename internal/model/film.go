package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Film struct {
	ID        uint        `gorm:"primarykey"`
	Url       string      `json:"url" gorm:"index"`
	Name      string      `json:"name"`
	Image     string      `json:"image"`
	Source    string      `json:"source"`
	Actor     string      `json:"actor"`
	ActorId   string      `json:"actor_id"`
	Date      time.Time   `json:"date"`
	CreatedAt time.Time   `json:"created_at"`
	Actors    StringArray `json:"actors" gorm:"type:json;serializer:json"`
	Title     string      `json:"title"`
}

type MagnetCache struct {
	ID         uint              `gorm:"primarykey"`
	Magnet     string            `json:"magnet" gorm:"index"`
	FileId     string            `json:"file_id"`
	Name       string            `json:"name" gorm:"index"`
	Code       string            `json:"code" gorm:"index"`
	DriverType string            `json:"driver_type"`
	Option     map[string]string `json:"option" gorm:"type:json;serializer:json"`
	Subtitle   bool              `json:"subtitle"`
	ScanAt     time.Time         `json:"scan_at"`
	ScanCount  uint              `json:"scan_count"`
}

type MissedFilm struct {
	gorm.Model
	Code string `json:"code" gorm:"index"`
}

type Actor struct {
	gorm.Model
	Dir  string `json:"dir" gorm:"index"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type VirtualFile struct {
	ID              uint          `gorm:"primarykey"`
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
	// 替换类型： 0：顺序重命名；1：正则重命名；
	Type          int    `json:"type"`
	Start         int    `json:"start"`
	End           int    `json:"end"`
	StartNum      int    `json:"startNum"`
	OldNameRegexp string `json:"oldNameRegexp"`
	SourceName    string `json:"sourceName"`
}

type Replacement struct {
	ID        uint   `gorm:"primarykey"`
	StorageId uint   `json:"storage_id"`
	DirName   string `json:"dir_name"`
	OldName   string `json:"old_name"`
	NewName   string `json:"new_name"`
	// 0: 重命名 1: 删除
	Type int `json:"type"`
}

type StringArray []string

func (o *StringArray) Scan(src any) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	return json.Unmarshal(bytes, o)
}
func (o StringArray) Value() (driver.Value, error) {
	return json.Marshal(o)
}
