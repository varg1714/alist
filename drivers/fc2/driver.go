package fc2

import (
	"context"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/av"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
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

	duration := time.Minute * time.Duration(d.ReleaseScanTime)
	if duration <= 0 {
		duration = time.Minute * 60
	}

	d.cron = cron.NewCron(duration)
	d.cron.Do(func() {
		d.reMatchReleaseTime()
		if d.RefreshNfo {
			d.refreshNfo()
		}
	})

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

	dirName := dir.GetName()

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		url := actor.Url
		categories[actor.Name] = url
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
	} else if dirName == "个人收藏" {
		return utils.SliceConvert(d.getStars(), func(src model.EmbyFileObj) (model.Obj, error) {
			return &src, nil
		})
	} else if categories[dirName] != "" {
		// 自定义目录
		var films []model.EmbyFileObj
		var err error
		if strings.Contains(categories[dirName], "missav.ai/dm99") {
			films, err = d.getMissAvFilms(dirName, func(index int) string {
				return d.ScraperApi + fmt.Sprintf(categories[dirName], index)
			})
		} else {
			films, err = d.getFilms(func(index int) string {
				return fmt.Sprintf(categories[dirName], index)
			})
		}
		if err != nil {
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

func (d *FC2) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	mockedLink := &model.Link{
		URL: d.MockedLink,
	}

	if d.MockedByMatchUa != "" && args.Header.Get("User-Agent") != d.MockedByMatchUa && d.MockedLink != "" {
		return mockedLink, nil
	}

	if d.Mocked && d.MockedLink != "" {
		return mockedLink, nil
	}

	return tool.CloudPlay(ctx, args, d.CloudPlayDriverType, d.CloudPlayDownloadPath, file, func(obj model.Obj) (string, error) {
		return d.getMagnet(obj)
	})

}

func (d *FC2) Remove(ctx context.Context, obj model.Obj) error {

	if obj.IsDir() {
		err := db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
		if err != nil {
			return err
		}

		return db.DeleteFilmsByActor("fc2", obj.GetName())
	} else {
		err := db.DeleteAllMagnetCacheByCode(av.GetFilmCode(obj.GetName()))
		if err != nil {
			utils.Log.Warnf("影片缓存信息删除失败：%s", err.Error())
		}
		err = virtual_file.DeleteImageAndNfo("fc2", "个人收藏", obj.GetName())
		if err != nil {
			utils.Log.Warnf("影片附件信息删除失败：%s", err.Error())
		}

		err = db.CreateMissedFilms([]string{av.GetFilmCode(obj.GetName())})
		if err != nil {
			utils.Log.Warnf("影片黑名单信息失败：%s", err.Error())
		}

		err = db.DeleteFilmsByCode("fc2", "个人收藏", av.GetFilmCode(obj.GetName()))
		if err != nil {
			utils.Log.Warnf("影片删除失败：%s", err.Error())
		}

		return err
	}

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
		url = fmt.Sprintf("https://fc2ppvdb.com/actresses/%s", split[1]) + "?page=%d"
	} else if actorType == 1 {
		// 贩卖者
		url = fmt.Sprintf("https://fc2ppvdb.com/writers/%s", split[1]) + "?page=%d"
	} else if actorType == 2 {
		// missAv fc2收藏榜
		url = "https://missav.ai/dm99/cn/fc2?sort=saved&page=%d"
	} else {
		return errors.New("illegal actorType")
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), split[0], url)

}

func (d *FC2) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	star, err := d.addStar(stream.GetName(), []string{})
	return &star, err

}

var _ driver.Driver = (*FC2)(nil)
