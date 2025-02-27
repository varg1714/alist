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
	OpenAiUrl             string `json:"open_ai_url" required:"true"`
	OpenAiApiKey          string `json:"open_ai_api_key" required:"true"`
	TranslatePromote      string `json:"translate_promote" required:"true"`
	QuickCache            bool   `json:"quick_cache" required:"true"`
	CloudPlayDriverType   string `json:"cloud_play_driver_type" required:"true"`
	CloudPlayDownloadPath string `json:"cloud_play_download_path" required:"true"`
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
