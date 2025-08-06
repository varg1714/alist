package quark_share

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	driver.RootPath
	Cookie          string `json:"cookie" required:"true"`
	QuarkDriverPath string `json:"QuarkDriverPath"`
	TransferPath    string `json:"transferPath"`
}

var config = driver.Config{
	Name:              "QuarkShare",
	LocalSort:         true,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          true,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &QuarkShare{
			conf: Conf{
				ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
				referer: "https://pan.quark.cn",
				api:     "https://drive-pc.quark.cn",
				pr:      "ucpro",
			},
		}
	})
}
