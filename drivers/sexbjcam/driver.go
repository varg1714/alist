package sexbj_cam

import (
	"context"
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/cron"
	json "github.com/json-iterator/go"
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

	dirName := dir.GetName()

	actors := strings.Split(d.Addition.Actors, ",")
	categories := make(map[string]string)

	err := json.Unmarshal([]byte(d.Categories), &categories)
	if err != nil {
		return results, err
	}

	// 1. 构建虚拟文件夹
	rootDirectory := []string{"关注演员"}
	for category := range categories {
		rootDirectory = append(rootDirectory, category)
	}

	if dirName == d.RootID.GetRootId() {
		// 根目录
		for _, category := range rootDirectory {
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
		// 关注演员目录
		for _, actor := range actors {
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
	} else if utils.Contains(actors, dirName) {
		// 关注演员影片
		return d.getActorFilms(func(index int) string {
			return fmt.Sprintf("https://sexbjcam.com/actor/%s/page/%s/", dirName, strconv.Itoa(index))

		})
	} else {

		category, exists := categories[dirName]
		if exists {
			films, err := d.getCategoryFilms(func(index int) string {
				return fmt.Sprintf(category, strconv.Itoa(index))

			})
			if err != nil {
				return results, err
			}
			return films, err
		} else {
			return results, nil
		}

	}

}

func (d *SexBjCam) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	link, err := d.getLink(file)
	if err != nil {
		return nil, err
	}

	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://sexbjcam.com/"},
		},
		URL: link,
	}, nil

}

var _ driver.Driver = (*SexBjCam)(nil)
