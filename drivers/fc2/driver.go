package fc2

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
	"strconv"
	"strings"
	"time"
)

type FC2 struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *FC2) Config() driver.Config {
	return config
}

func (d *FC2) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *FC2) Init(ctx context.Context) error {
	return nil
}

func (d *FC2) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *FC2) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	categories := make(map[string]string)
	results := make([]model.Obj, 0)

	storage := op.GetBalancedStorage(d.PikPakPath)
	pikPak, ok := storage.(*pikpak.PikPak)
	if !ok {
		return results, nil
	}

	dirName := dir.GetName()

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		url := actor.Url
		categories[actor.Name] = url
	}

	if d.RootID.GetRootId() == dirName {
		// 1. 顶级目录
		for category := range categories {
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
	} else if categories[dirName] != "" {
		// 自定义目录
		return d.getFilms(dirName, func(index int) string {
			return fmt.Sprintf(categories[dirName], index)
		})
	} else if !strings.Contains(dir.GetID(), ".jpg") {
		// 临时文件
		magnet, err := d.getMagnet(dir)
		if err != nil || magnet == "" {
			return results, err
		}
		return pikPak.CloudDownload(ctx, d.PikPakCacheDirectory, magnet)
	} else {
		// pikPak文件
		return results, nil
	}

}

func (d *FC2) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

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

func (d *FC2) Remove(ctx context.Context, obj model.Obj) error {

	err := db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
	if err != nil {
		return err
	}

	return db.DeleteByActor("fc2", obj.GetName())

}

func (d *FC2) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	split := strings.Split(dirName, " ")
	if len(split) != 3 {
		return errors.New("illegal dirName")
	}

	actorType, err := strconv.Atoi(split[2])
	if err != nil {
		return errors.New("illegal dirName")
	}

	var url string
	if actorType == 0 {
		// 0 演员
		url = fmt.Sprintf("https://adult.contents.fc2.com/users/%s/articles?sort=popular&order=desc&deal=", split[1]) + "&page=%d"
	} else if actorType == 1 {
		// yearly
		url = fmt.Sprintf("https://adult.contents.fc2.com/ranking/article/yearly?year=%s", split[1]) + "&page=%d"
	} else if actorType == 2 {
		// monthly
		now := time.Now()
		url = fmt.Sprintf("https://adult.contents.fc2.com/ranking/article/monthly?year=%d&month=%d", now.Year(), now.Month()-1) + "&page=%d"
	} else if actorType == 3 {
		// weekly
		url = "https://adult.contents.fc2.com/ranking/article/weekly" + "?page=%d"
	} else {
		return err
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), split[0], url)

}

var _ driver.Driver = (*FC2)(nil)
