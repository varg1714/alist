package pornhub

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	driver.RootID
	OrderBy           string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection    string `json:"order_direction" type:"select" options:"ASC,DESC"`
	Mocked            bool   `json:"mocked"`
	MockedLink        string `json:"mocked_link" `
	MockedByMatchUa   string `json:"mocked_by_match_ua"`
	SpiderServer      string `json:"spider_server" required:"true"`
	SpiderMaxWaitTime uint64 `json:"spider_max_wait_time" required:"true" type:"number" `
	ServerUrl         string `json:"server_url" required:"true"`
}

var config = driver.Config{
	Name:        "Pornhub",
	LocalSort:   false,
	OnlyProxy:   false,
	NoUpload:    false,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Pornhub{}
	})
}
