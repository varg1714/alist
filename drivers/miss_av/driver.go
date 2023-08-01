package miss_av

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/pikpak"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/cron"
	json "github.com/json-iterator/go"
	"strings"
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

	categories := make(map[string]string)
	results := make([]model.Obj, 0)
	err := json.Unmarshal([]byte(d.Categories), &categories)
	if err != nil {
		return results, err
	}

	storage := op.GetBalancedStorage(d.PikPakPath)
	pikPak, ok := storage.(*pikpak.PikPak)
	if !ok {
		return results, nil
	}

	dirName := dir.GetName()
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
	} else if strings.Contains(dir.GetID(), "https://") && !strings.Contains(dir.GetID(), ".jpg") {
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

func (d *MIssAV) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

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

func (d *MIssAV) Remove(ctx context.Context, obj model.Obj) error {

	categories := make(map[string]string)
	err := json.Unmarshal([]byte(d.Categories), &categories)
	if err != nil {
		return err
	}

	if categories[obj.GetName()] != "" {
		err = db.DeleteByActor("javdb", obj.GetName())
		if err != nil {
			return err
		}
	}

	return err
}

var _ driver.Driver = (*MIssAV)(nil)
