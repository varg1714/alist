package miss_av

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	json "github.com/json-iterator/go"
	"net/http"
	"time"
)

type MIssAV struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *MIssAV) Config() driver.Config {
	return config
}

func (d *MIssAV) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *MIssAV) Init(ctx context.Context) error {
	return nil
}

func (d *MIssAV) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *MIssAV) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	results := make([]model.Obj, 0)
	m := make(map[string]string)

	err := json.Unmarshal([]byte(d.Categories), &m)
	if err != nil {
		return results, err
	}

	dirName := dir.GetName()
	if d.RootID.GetRootId() == dirName {
		// 1. 顶级目录
		for category, _ := range m {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     category,
					IsFolder: true,
					ID:       category,
					Size:     622857143,
					Modified: time.Now(),
				},
			})
		}
		return results, nil
	} else {

		url, exist := m[dirName]
		if !exist {
			return results, nil
		}

		return d.getFilms(func(index int) string {
			return fmt.Sprintf(url, index)
		})

	}

}

func (d *MIssAV) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	link, err := d.getLink(file)
	if err != nil {
		return nil, err
	}

	//log.Infof("res:%s,url:%s\n", res, videoUrl)
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://madou.club/"},
		},
		URL: link,
	}, nil

}

var _ driver.Driver = (*MIssAV)(nil)
