package jable_tv

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type JableTV struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *JableTV) Config() driver.Config {
	return config
}

func (d *JableTV) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *JableTV) Init(ctx context.Context) error {
	return nil
}

func (d *JableTV) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *JableTV) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	results := make([]model.Obj, 0)
	//log.Infof("我自己的内容:dir:%s,args:%s,RefreshToken:%s,ShareId:%s\n", dir.GetName(), args, d.RefreshToken, d.ShareId)

	actorsMap := make(map[string]string)
	categoriesMap := make(map[string]string)

	err := json.Unmarshal([]byte(d.Actors), &actorsMap)
	if err != nil {
		return results, err
	}
	err = json.Unmarshal([]byte(d.Categories), &categoriesMap)
	if err != nil {
		return results, err
	}

	dirName := dir.GetName()

	if d.RootID.GetRootId() == dirName {
		// 1. 顶级目录
		// 1.1 返回演员目录
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     "关注演员",
				IsFolder: true,
				ID:       "关注演员",
				Size:     622857143,
				Modified: time.Now(),
			},
		})

		// 1.2 添加系统目录
		for category := range categoriesMap {
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
	} else if dirName == "关注演员" {
		for actor := range actorsMap {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     actor,
					IsFolder: true,
					ID:       actor,
					Size:     622857143,
					Modified: time.Now(),
				},
			})
		}
		return results, nil
	} else {

		actor, exist := actorsMap[dirName]
		if exist {
			return getActorFilms(actor, results)
		}

		category, exist := categoriesMap[dirName]
		if exist {
			return getFilms(func(index int) string {
				return fmt.Sprintf(category, strconv.Itoa(index))
			})
		}

	}

	return results, nil
}

func (d *JableTV) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	code := strings.Split(file.GetName(), " ")

	url := fmt.Sprintf("https://jable.tv/videos/%s/", code[1])
	//log.Infof("影片访问地址:%s\n", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
		}).
		Post("http://103.140.9.114:7856/transfer")
	if err != nil {
		log.Errorf("出错了：%s,%s\n", err, res)
		return nil, err
	}

	videoRegexp, _ := regexp.Compile(".*var hlsUrl = '(.*)'.*")
	videoUrl := videoRegexp.FindString(string(res.Body()))

	//log.Infof("res:%s,url:%s\n", res, videoUrl)
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://jable.tv/"},
		},
		URL: videoRegexp.ReplaceAllString(videoUrl, "$1"),
	}, nil

}

var _ driver.Driver = (*JableTV)(nil)
