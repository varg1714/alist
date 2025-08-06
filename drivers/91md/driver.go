package _91_md

import (
	"context"
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/cron"
	json "github.com/json-iterator/go"
	"net/http"
	"strconv"
	"time"
)

type _91MD struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
	cron        *cron.Cron
}

func (d *_91MD) Config() driver.Config {
	return config
}

func (d *_91MD) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *_91MD) Init(ctx context.Context) error {
	return nil
}

func (d *_91MD) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *_91MD) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	dirName := dir.GetName()

	searchers := make(map[string]string)
	categories := make(map[string]string)

	err := json.Unmarshal([]byte(d.Addition.Categories), &categories)
	if err != nil {
		return results, err
	}

	err = json.Unmarshal([]byte(d.Addition.Searchers), &searchers)
	if err != nil {
		return results, err
	}

	// 1. 构建虚拟文件夹
	rootDirectory := []string{"搜索目录"}
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
	} else if dirName == "搜索目录" {
		// 搜索目录
		for searcher := range searchers {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     searcher,
					IsFolder: true,
					ID:       searcher,
					Size:     622857143,
					Modified: time.Now(),
				},
			})
		}
		return results, nil
	} else if searchers[dirName] != "" {
		// 搜索目录
		return d.getActorFilms(func(index int) string {
			return fmt.Sprintf(searchers[dirName], strconv.Itoa(index))

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

func (d *_91MD) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	link, err := d.getLink(file)
	if err != nil {
		return nil, err
	}

	return &model.Link{
		Header: http.Header{
			"Referer": []string{"https://91md.me"},
		},
		URL: link,
	}, nil

}

var _ driver.Driver = (*_91MD)(nil)
