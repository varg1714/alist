package fc2

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	SpiderServer string `json:"spider_server" required:"true"`
	driver.RootID
	OrderBy               string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection        string `json:"order_direction" type:"select" options:"ASC,DESC"`
	Mocked                bool   `json:"mocked"`
	MockedLink            string `json:"mocked_link" `
	CloudPlayDriverType   string `json:"cloud_play_driver_type" required:"true" default:"PikPak" type:"select" options:"PikPak,115 Cloud"`
	CloudPlayDownloadPath string `json:"cloud_play_download_path" required:"false" help:"If empty then use global setting."`
	ReleaseScanTime       uint64 `json:"ReleaseScanTime" required:"true" type:"number" `
	ScanTimeLimit         uint64 `json:"ScanTimeLimit" required:"true" type:"number" `
	RefreshNfo            bool   `json:"refresh_nfo"`
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
