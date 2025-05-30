package fc2

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	SpiderServer      string `json:"spider_server" required:"true"`
	SpiderMaxWaitTime uint64 `json:"spider_max_wait_time" required:"true" type:"number" `
	driver.RootID
	OrderBy               string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection        string `json:"order_direction" type:"select" options:"ASC,DESC"`
	Mocked                bool   `json:"mocked"`
	MockedLink            string `json:"mocked_link" `
	MockedByMatchUa       string `json:"mocked_by_match_ua"`
	CloudPlayDriverType   string `json:"cloud_play_driver_type" required:"true" default:"PikPak" type:"select" options:"PikPak,115 Cloud"`
	CloudPlayDownloadPath string `json:"cloud_play_download_path" required:"false" help:"If empty then use global setting."`
	ReleaseScanTime       uint64 `json:"ReleaseScanTime" required:"true" type:"number" `
	ScanTimeLimit         uint64 `json:"ScanTimeLimit" required:"true" type:"number" `
	RefreshNfo            bool   `json:"refresh_nfo"`
	ScraperApi            string `json:"scraper_api" required:"false"`
	MissAvMaxPage         int    `json:"miss_av_max_page" required:"true" type:"number" `
	EmbyServers           string `json:"emby_servers" required:"false" type:"text"`
}

var config = driver.Config{
	Name:        "FC2",
	LocalSort:   false,
	OnlyProxy:   false,
	NoUpload:    false,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &FC2{}
	})
}
