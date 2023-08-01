package miss_av

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm/utils"
	"regexp"
	"time"
)

var subTitles, _ = regexp.Compile(".*<div class=\".*\">\\s*?<div class=\".*\">\\s*?<a href=\"(.*)\" title=\".*\">\\s*?<span class=\"name\">.*</span>\\s*?<br />\\s*?<span class=\"meta\">[\\s|\\S]*?</span>\\s*?<br/>\\s*?<div class=\"tags\">\\s*?<span class=\".*\">高清</span>\\s*?<span class=\".*\">字幕</span>\\s*?</div>\\s*?</a>\\s*?</div>")
var hd, _ = regexp.Compile(".*<div class=\".*\">\\s*?<div class=\".*\">\\s*?<a href=\"(.*)\" title=\".*\">\\s*?<span class=\"name\">.*</span>\\s*?<br />\\s*?<span class=\"meta\">[\\s]*.*[\\s]*</span>\\s*?<br/>\\s*?<div class=\"tags\">\\s*?<span class=\".*\">高清</span>\\s*?</div>\\s*?</a>\\s*?</div>")

func convertToModel(films []string, images []string, urls []string) []model.ObjThumb {

	results := make([]model.ObjThumb, 0)

	for index, film := range films {

		var image string
		if index < cap(images) {
			image = images[index]
		}
		//log.Infof("index:%s,image:%s,cap:%s,images:%s\n", index, image, cap(images), images)

		results = append(results, model.ObjThumb{
			Object: model.Object{
				Name:     film,
				IsFolder: true,
				ID:       urls[index],
				Size:     622857143,
				Modified: time.Now(),
			},
			Thumbnail: model.Thumbnail{Thumbnail: image},
		})
	}
	return results
}

func (d *MIssAV) findPage(url string) (*resty.Response, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
		}).
		Post(d.Addition.SpiderServer)

	return res, err
}

func (d *MIssAV) getFilms(dirName string, urlFunc func(index int) string) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	films := make([]string, 0)
	images := make([]string, 0)
	urls := make([]string, 0)
	nextPage := false
	var err error

	films, images, urls, nextPage, err = d.getPageInfo(urlFunc, 1, films, images, urls)
	if err != nil {
		return results, err
	}

	existFilms := db.QueryByUrls(dirName, urls)

	// not exists
	for index := 2; index <= 20 && nextPage && len(existFilms) == 0; index++ {

		films, images, urls, nextPage, err = d.getPageInfo(urlFunc, index, films, images, urls)
		if err != nil {
			return results, err
		}

		existFilms = db.QueryByUrls(dirName, urls)

	}
	// exist
	for index, url := range urls {
		if utils.Contains(existFilms, url) {
			if index == 0 {
				urls = []string{}
				images = []string{}
				films = []string{}
			} else {
				urls = urls[:index]
				images = images[:index]
				films = films[:index]
			}
			break
		}
	}

	if len(urls) != 0 {
		err = db.CreateFilms("javdb", dirName, convertToModel(films, images, urls))
		if err != nil {
			return results, nil
		}
	}

	return d.convertFilm(db.QueryByActor("javdb", dirName), results), nil

}

func (d *MIssAV) convertFilm(actor []model.Film, results []model.Obj) []model.Obj {
	for index, film := range actor {
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     fmt.Sprintf("%04d", index) + " " + film.Name,
				IsFolder: true,
				ID:       film.Url,
				Size:     622857143,
				Modified: time.Now(),
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		})
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     fmt.Sprintf("%04d", index) + " " + film.Name + ".jpg",
				IsFolder: false,
				ID:       film.Image,
				Size:     622857143,
				Modified: time.Now(),
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		})
	}
	return results
}

func (d *MIssAV) getMagnet(file model.Obj) (string, error) {

	pageUrl := file.GetID()

	res, err := d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	page := string(res.Body())

	url := subTitles.FindString(page)
	if url != "" {
		return subTitles.ReplaceAllString(url, "$1"), nil
	}

	url = hd.FindString(page)
	if url != "" {
		return hd.ReplaceAllString(url, "$1"), nil
	}

	return "", nil

}

func (d *MIssAV) getPageInfo(urlFunc func(index int) string, index int, films []string, images []string, urls []string) ([]string, []string, []string, bool, error) {

	urlsRegexp, _ := regexp.Compile(".*<a href=\"(.*)\" class=\"box\" title=\".*\">.*")
	filmsRegexp, _ := regexp.Compile(".*<div class=\"video-title\"><strong>(.*)</strong>(.*)</div>.*")
	imageRegexp, _ := regexp.Compile(".*<img loading=\"lazy\" src=\"(.*)\" />.*")
	pagesRegexp, _ := regexp.Compile(".*<a rel=\"next\" class=\"pagination-next\" href=\".*\">下一頁</a>.*")

	pageUrl := urlFunc(index)
	//log.Infof("开始查询%s", pageUrl)

	res, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, false, nil
	}

	page := string(res.Body())

	tempUrls := urlsRegexp.FindAllString(page, -1)
	tempFilms := filmsRegexp.FindAllString(page, -1)
	imageUrls := imageRegexp.FindAllString(page, -1)
	pages := pagesRegexp.FindAllString(page, -1)

	for _, file := range tempFilms {
		films = append(films, filmsRegexp.ReplaceAllString(file, "$1$2"))
	}
	for _, imageUrl := range imageUrls {
		images = append(images, imageRegexp.ReplaceAllString(imageUrl, "$1"))
	}
	for _, tempUrl := range tempUrls {
		urls = append(urls, "https://javdb.com/"+urlsRegexp.ReplaceAllString(tempUrl, "$1"))
	}

	return films, images, urls, len(pages) != 0, nil

}
