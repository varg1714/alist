package pornhub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type Pornhub struct {
	model.Storage
	Addition
	AccessToken string
	ShareToken  string
	DriveId     string
}

func (d *Pornhub) Config() driver.Config {
	return config
}

func (d *Pornhub) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Pornhub) Init(ctx context.Context) error {
	return nil
}

func (d *Pornhub) Drop(ctx context.Context) error {
	return nil
}

func (d *Pornhub) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	categories := make(map[string]string)
	results := make([]model.Obj, 0)

	dirName := dir.GetName()

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		url := actor.Url
		categories[actor.Name] = url
	}

	if d.RootID.GetRootId() == dirName {
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
		films, err := d.getFilms(dirName, categories[dirName])
		if err != nil {
			return nil, err
		}
		return utils.SliceConvert(films, func(src model.EmbyFileObj) (model.Obj, error) {
			return &src, nil
		})
	} else {
		return results, nil
	}

}

func (d *Pornhub) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	videoLink := &model.Link{
		URL: d.MockedLink,
	}

	if d.MockedByMatchUa != "" && !virtual_file.AllowUA(args.Header.Get("User-Agent"), d.MockedByMatchUa) && d.MockedLink != "" {
		return videoLink, nil
	}

	if d.Mocked && d.MockedLink != "" {
		return videoLink, nil
	}

	if embyFile, ok := file.(*model.EmbyFileObj); ok {
		link, err := d.getVideoLink(embyFile.Url)
		if err != nil {
			utils.Log.Warnf("failed to get video link: %v", err.Error())
			return videoLink, nil
		}

		videoLink.URL = link
		return videoLink, nil
	} else {
		return nil, errors.New("invalid file type")
	}

}

func (d *Pornhub) Remove(ctx context.Context, obj model.Obj) error {

	if !obj.IsDir() {
		return nil
	}

	err := db.DeleteActor(strconv.Itoa(int(d.ID)), obj.GetName())
	if err != nil {
		return err
	}

	return db.DeleteFilmsByActor("pornhub", obj.GetName())

}

func (d *Pornhub) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	actorType := 0
	name := ""
	url := ""

	if strings.HasPrefix(dirName, "{") {
		var param MakeDirParam
		err := json.Unmarshal([]byte(dirName), &param)
		if err != nil {
			return err
		}
		name = param.Name
		url = param.Url
		actorType = param.Type
	} else {
		split := strings.Split(dirName, " ")
		if len(split) != 3 {
			return errors.New("illegal dirName")
		}

		tempType, err := strconv.Atoi(split[2])
		if err != nil {
			return errors.New("illegal dirName")
		}
		actorType = tempType
		name = split[0]
		url = split[1]
	}

	if actorType == PlayList {
		// playlist
		url = fmt.Sprintf("/playlist/%s", url)
	} else if actorType == ACTOR {
		// actor
		url = fmt.Sprintf("/model/%s", url)
	} else {
		return errors.New("illegal actorType")
	}

	return db.CreateActor(strconv.Itoa(int(d.ID)), name, url)

}

var _ driver.Driver = (*Pornhub)(nil)
