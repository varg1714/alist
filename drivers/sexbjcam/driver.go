package sexbj_cam

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	"gorm.io/gorm/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SexBjCam struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *SexBjCam) Config() driver.Config {
	return config
}

func (d *SexBjCam) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *SexBjCam) Init(ctx context.Context) error {
	return nil
}

func (d *SexBjCam) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *SexBjCam) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	virtualDirectoryMap := make(map[string][]string, 0)

	dirName := dir.GetName()

	actors := strings.Split(d.Addition.Actors, ",")

	// 1. 构建虚拟文件夹
	virtualDirectoryMap[d.RootID.GetRootId()] = []string{"关注演员"}
	virtualDirectoryMap["关注演员"] = actors

	// 2. 构建页面路径
	directory, exist := virtualDirectoryMap[dirName]
	if exist {
		for _, category := range directory {
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
	}

	if utils.Contains(actors, dirName) {
		// 演员列表
		return d.getFilms(func(index int) string {
			return fmt.Sprintf("https://sexbjcam.com/actor/%s/page/%s/", dirName, strconv.Itoa(index))

		})
	}
	return results, nil
}

func (d *SexBjCam) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

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

var _ driver.Driver = (*SexBjCam)(nil)
