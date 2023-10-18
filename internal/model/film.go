package model

import "time"

type Film struct {
	Url       string    `json:"url" gorm:"index"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Source    string    `json:"source"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

type MagnetCache struct {
	Magnet string `json:"magnet" gorm:"index"`
	FileId string `json:"file_id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
}

type Actor struct {
	Dir  string `json:"dir" gorm:"index"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type VirtualFile struct {
	StorageId       string `json:"storage_id"`
	Name            string `json:"name"`
	ShareId         string `json:"share_id"`
	ParentDir       string `json:"parent_dir"`
	AppendSubFolder int    `json:"append_sub_folder"`
	ExcludeUnMatch  bool   `json:"exclude_un_match"`
	Start           int    `json:"start" gorm:"default -1"`
	End             int    `json:"end" gorm:"default -1"`
	SourceName      string `json:"source_name"`
	StartNum        int    `json:"start_num"`
	Type            int    `json:"type"`
}
