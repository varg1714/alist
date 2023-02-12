package jable_tv

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

type MaDouClub struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *MaDouClub) Config() driver.Config {
	return config
}

func (d *MaDouClub) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *MaDouClub) Init(ctx context.Context) error {
	return nil
}

func (d *MaDouClub) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *MaDouClub) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	results := make([]model.Obj, 0)
	//log.Infof("我自己的内容:dir:%s,args:%s,RefreshToken:%s,ShareId:%s\n", dir.GetName(), args, d.RefreshToken, d.ShareId)

	virtualDirectoryMap := make(map[string][]string, 0)

	dirName := dir.GetName()

	tags := strings.Split(d.Addition.Tags, ",")
	categories := strings.Split(d.Addition.Categories, ",")
	searchers := strings.Split(d.Addition.Searchers, ",")

	// 1. 构建虚拟文件夹
	virtualDirectoryMap[d.RootID.GetRootId()] = []string{"标签", "分类", "搜索关注"}
	virtualDirectoryMap["标签"] = tags
	virtualDirectoryMap["分类"] = categories
	virtualDirectoryMap["搜索关注"] = searchers

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

	if utils.Contains(tags, dirName) {
		// 标签列表
		return getFilms(func(index int) string {
			return fmt.Sprintf("https://madou.club/tag/%s/page/%s", dirName, strconv.Itoa(index))

		})
	} else if utils.Contains(categories, dirName) {
		// 分类列表
		return getFilms(func(index int) string {
			return fmt.Sprintf("https://madou.club/category/%s/page/%s", dirName, strconv.Itoa(index))

		})
	} else if utils.Contains(searchers, dirName) {
		// 搜索列表
		return getFilms(func(index int) string {
			return fmt.Sprintf("https://madou.club/page/%s?s=%s", strconv.Itoa(index), dirName)
		})
	}

	return results, nil
}

func (d *MaDouClub) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	link, err := getLink(file)
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

var _ driver.Driver = (*MaDouClub)(nil)
