package javdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/pikpak"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"strconv"
	"strings"
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

	storage := op.GetBalancedStorage(d.PikPakPath)
	pikPak, ok := storage.(*pikpak.PikPak)
	if !ok {
		return results, nil
	}

	dirName := dir.GetName()

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		categories.Put(actor.Name, actor)
	}

	if d.RootID.GetRootId() == dirName {
		// 1. 顶级目录
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
	} else if actor, exist := categories.Get(dirName); exist {
		// 自定义目录
		url := actor.Url
		if !strings.HasPrefix(url, "http") {
			url = "https://javdb.com/actors/" + url + "?page=%d&sort_type=0&t=d"
		}
		return d.getFilms(dirName, func(index int) string {
			return fmt.Sprintf(url, index)
		})
	} else if strings.Contains(dir.GetID(), "https://") && !strings.Contains(dir.GetID(), ".jpg") {
		// 临时文件
		magnet, err := d.getMagnet(dir)
		if err != nil || magnet == "" {
			return results, err
		}
		return pikPak.CloudDownload(ctx, d.PikPakCacheDirectory, dir.GetPath(), dir.GetName(), magnet)
	} else {
		// pikPak文件
		return results, nil
	}

}

func (d *Javdb) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	if strings.Contains(file.GetID(), ".jpg") {
		return &model.Link{
			URL: file.GetID(),
		}, nil
	}

	storage := op.GetBalancedStorage(d.PikPakPath)
	pikPak, ok := storage.(*pikpak.PikPak)
	if !ok {
		return &model.Link{
			URL: "",
		}, nil
	}

	return pikPak.Link(ctx, file, args)

}

func (d *Javdb) Remove(ctx context.Context, obj model.Obj) error {

	err := db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
	if err != nil {
		return err
	}

	return db.DeleteByActor("javdb", obj.GetName())

}

func (d *Javdb) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	split := strings.Split(dirName, " ")
	if len(split) != 2 {
		return errors.New("illegal dirName")
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), split[0], split[1])

}

var _ driver.Driver = (*Javdb)(nil)
