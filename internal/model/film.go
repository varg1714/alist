package model

type Film struct {
	Url    string `json:"url" gorm:"index"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Source string `json:"source"`
	Actor  string `json:"actor"`
}

type MagnetCache struct {
	Magnet string `json:"magnet" gorm:"index"`
	FileId string `json:"file_id"`
}
