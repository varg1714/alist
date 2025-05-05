package javdb

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	SpiderServer string `json:"spider_server"`
	Cookie       string `json:"cookie" required:"true"`
	driver.RootID
	OrderBy               string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection        string `json:"order_direction" type:"select" options:"ASC,DESC"`
	Mocked                bool   `json:"mocked"`
	MockedLink            string `json:"mocked_link" `
	MockedByMatchUa       string `json:"mocked_by_match_ua"`
	QuickCache            bool   `json:"quick_cache" required:"true"`
	CloudPlayDriverType   string `json:"cloud_play_driver_type" required:"true" default:"PikPak" type:"select" options:"PikPak,115 Cloud"`
	BackPlayDriverType    string `json:"back_play_driver_type" required:"false" default:"PikPak" type:"select" options:"PikPak,115 Cloud"`
	FallbackPlay          bool   `json:"fallback_play" required:"false"`
	CloudPlayDownloadPath string `json:"cloud_play_download_path" required:"false" help:"If empty then use global setting."`
	Filter                string `json:"filter" required:"false" help:"Multi values must separated by commas."`
	SubtitleScanTime      uint64 `json:"subtitle_scan_time" required:"true" type:"number" `
	RefreshNfo            bool   `json:"refresh_nfo"`
	EmbyServers           string `json:"emby_servers" required:"false" type:"text"`
}

var config = driver.Config{
	Name:        "Javdb",
	LocalSort:   false,
	OnlyProxy:   false,
	NoUpload:    false,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Javdb{}
	})
}
