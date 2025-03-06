package javdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"strconv"
	"strings"
	"time"
)

type Javdb struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *Javdb) Config() driver.Config {
	return config
}

func (d *Javdb) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Javdb) Init(ctx context.Context) error {
	return nil
}

func (d *Javdb) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *Javdb) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	categories := linkedhashmap.New[string, model.Actor]()
	results := make([]model.Obj, 0)

	dirName := dir.GetName()

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		categories.Put(actor.Name, actor)
	}

	if d.RootID.GetRootId() == dirName {
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     "关注演员",
				IsFolder: true,
				ID:       "关注演员",
				Size:     622857143,
				Modified: time.Now(),
			},
		}, &model.ObjThumb{
			Object: model.Object{
				Name:     "个人收藏",
				IsFolder: true,
				ID:       "个人收藏",
				Size:     622857143,
				Modified: time.Now(),
			},
		})
		return results, nil
	} else if dirName == "关注演员" {
		// 1. 关注演员
		categories.Each(func(name string, actor model.Actor) {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     name,
					IsFolder: true,
					ID:       name,
					Size:     622857143,
					Modified: actor.Model.UpdatedAt,
				},
			})
		})
		return results, nil
	} else if dirName == "个人收藏" {
		// 2. 个人收藏
		return utils.SliceConvert(d.getStars(), func(src model.EmbyFileObj) (model.Obj, error) {
			return &src, nil
		})
	} else if actor, exist := categories.Get(dirName); exist {
		// 自定义目录
		url := actor.Url
		if !strings.HasPrefix(url, "http") {
			url = "https://javdb.com/actors/" + url + "?page=%d&sort_type=0"
		}

		films, err := d.getFilms(dirName, func(index int) string {
			return fmt.Sprintf(url, index)
		})
		if err != nil {
			utils.Log.Info("影片获取失败", err)
			return nil, err
		}
		return utils.SliceConvert(films, func(src model.EmbyFileObj) (model.Obj, error) {
			return &src, nil
		})

	} else {
		// pikPak文件
		return results, nil
	}

}

func (d *Javdb) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	if d.Mocked && d.MockedLink != "" {
		utils.Log.Infof("jdvdb返回的地址: %s", d.MockedLink)
		return &model.Link{
			URL: d.MockedLink,
		}, nil
	}

	if strings.HasSuffix(file.GetID(), "jpg") {
		return &model.Link{
			URL: file.GetID(),
		}, nil
	}

	return tool.CloudPlay(ctx, args, d.CloudPlayDriverType, d.CloudPlayDownloadPath, file, func(obj model.Obj) (string, error) {
		return d.getMagnet(obj)
	})
}

func (d *Javdb) Remove(ctx context.Context, obj model.Obj) error {

	if obj.IsDir() {
		err := db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
		if err != nil {
			return err
		}

		return db.DeleteFilmsByActor("javdb", obj.GetName())
	} else {

		err := db.DeleteAllMagnetCacheByCode(obj.GetName())
		if err != nil {
			utils.Log.Warnf("影片缓存信息删除失败：%s", err.Error())
		}

		err = db.DeleteFilmsByUrl("javdb", obj.GetPath(), obj.GetID())
		if err != nil {
			utils.Log.Info("收藏影片删除失败", err)
			return err
		}

		err = virtual_file.DeleteImageAndNfo("javdb", obj.GetPath(), obj.GetName())
		if err != nil {
			utils.Log.Info("影片附件删除失败", err)
			return err
		}

	}

	return nil

}

func (d *Javdb) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	split := strings.Split(dirName, " ")
	if len(split) != 2 {
		return errors.New("illegal dirName")
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), split[0], split[1])

}

func (d *Javdb) Move(ctx context.Context, srcObj, dstDir model.Obj) error {

	if len(db.QueryByUrls("个人收藏", []string{srcObj.GetID()})) == 0 {
		thumb := srcObj.(*model.EmbyFileObj)
		return db.CreateFilms("javdb", "个人收藏", "个人收藏", []model.EmbyFileObj{*thumb})
	}

	return nil

}

func (d *Javdb) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	star, err := d.addStar(stream.GetName())
	return &star, err

}

var _ driver.Driver = (*Javdb)(nil)
