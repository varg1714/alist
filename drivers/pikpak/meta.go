package pikpak

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Username                string `json:"username" required:"true"`
	Password                string `json:"password" required:"true"`
	FileNameBlackChars      string `json:"file_name_black_chars"`
	LinkIndex               uint   `json:"link_index" type:"number" default:"0"`
	Platform                string `json:"platform" required:"true" default:"web" type:"select" options:"android,web,pc"`
	RefreshToken            string `json:"refresh_token" required:"true" default:""`
	RefreshTokenMethod      string `json:"refresh_token_method" required:"true" type:"select" options:"oauth2,http"`
	CaptchaToken            string `json:"captcha_token" default:""`
	DeviceID                string `json:"device_id"  required:"false" default:""`
	DisableMediaLink        bool   `json:"disable_media_link" default:"true"`
}

var config = driver.Config{
	Name:        "PikPak",
	LocalSort:   true,
	DefaultRoot: "",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &PikPak{}
	})
}
