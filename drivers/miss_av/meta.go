package miss_av

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	PikPakPath           string `json:"pik_pak_path" required:"true"`
	PikPakCacheDirectory string `json:"pik_pak_cache_directory" required:"true"`
	SpiderServer         string `json:"spider_server" required:"true"`
	Cookie               string `json:"cookie" required:"true"`
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
