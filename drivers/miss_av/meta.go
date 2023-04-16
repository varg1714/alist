package miss_av

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Categories      string `json:"tags" required:"true"`
	Actors          string `json:"actors" required:"true"`
	SpiderServer    string `json:"spider_server" required:"true"`
	PlayServer      string `json:"play_server" required:"true"`
	PlayProxyServer string `json:"play_proxy_server" required:"true"`
	driver.RootID
	OrderBy        string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection string `json:"order_direction" type:"select" options:"ASC,DESC"`
}

var config = driver.Config{
	Name:        "MissAV",
	LocalSort:   false,
	OnlyProxy:   false,
	NoUpload:    true,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &MIssAV{}
	})
}
