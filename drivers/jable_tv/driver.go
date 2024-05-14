package jable_tv

import (
	"context"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cron"
	json "github.com/json-iterator/go"
	"net/http"
	"regexp"
	"sort"
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

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))

	actorsMap := make(map[string]model.Actor)

	for _, film := range actors {
		actorsMap[film.Name] = film
	}
	categoriesMap := make(map[string]string)

	err := json.Unmarshal([]byte(d.Categories), &categoriesMap)
	if err != nil {
		return results, err
	}

	var categories []string
	for category := range categoriesMap {
		categories = append(categories, category)
	}

	sort.Strings(categories)

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
		for _, category := range categories {
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

	if dirName == "关注演员" {
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
			return d.getActorFilms(actor.Url, results)
		}

		category, exist := categoriesMap[dirName]
		if exist {
			return d.getFilms(func(index int) string {
				return fmt.Sprintf(category, strconv.Itoa(index))
			})
		}

	}

	return results, nil
}

func (d *JableTV) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	page, err := d.getFilmPage(file.GetName())
	if err != nil {
		return nil, err
	}

	videoRegexp, _ := regexp.Compile(".*var hlsUrl = '(.*)'.*")
	videoUrl := videoRegexp.FindString(page)

	//log.Infof("res:%s,url:%s\n", res, videoUrl)
	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://jable.tv/"},
		},
		URL: videoRegexp.ReplaceAllString(videoUrl, "$1"),
	}, nil

}

func (d *JableTV) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	split := strings.Split(dirName, " ")
	if len(split) != 2 {
		return errors.New("illegal dirName")
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), split[0], split[1])

}

func (d *JableTV) Rename(ctx context.Context, srcObj model.Obj, newName string) error {

	page, err := d.getFilmPage(srcObj.GetName())
	if err != nil {
		return err
	}

	actorRegexp, _ := regexp.Compile(".*<a class=\"model\" href=\"https://jable.tv/models/(.*)/\">.*")
	actor := actorRegexp.FindString(page)

	if actor == "" {
		return errors.New("actor is null")
	}

	err = db.DeleteActor(strconv.Itoa(int(d.ID)), newName)
	if err != nil {
		return err
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), newName, actorRegexp.ReplaceAllString(actor, "$1"))

}

func (d *JableTV) Remove(ctx context.Context, obj model.Obj) error {
	return db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
}

var _ driver.Driver = (*JableTV)(nil)
