package _91_md

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	Searchers      string `json:"searchers" required:"true"`
	Categories     string `json:"categories" required:"true"`
	TransferServer string `json:"transfer_server" required:"true"`
	SpiderServer   string `json:"spider_server" required:"true"`
	PlayServer     string `json:"play_server" required:"true"`
	driver.RootID
	OrderBy        string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection string `json:"order_direction" type:"select" options:"ASC,DESC"`
}

var config = driver.Config{
	Name:        "91MD",
	LocalSort:   false,
	OnlyProxy:   false,
	NoUpload:    true,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &_91MD{}
	})
}
