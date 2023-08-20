package fc2

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/gorm/utils"
	"regexp"
	"strings"
	"time"
)

var subTitles, _ = regexp.Compile(".*<a href=\"(.*)\" title=\".*</a>.*")
var magnetUrl, _ = regexp.Compile(".*<a href=\"(.*)\" class=\".*\"><i class=\".*\"></i>Magnet</a>.*")

var actorUrlsRegexp, _ = regexp.Compile(".*<a href=\"/article_search.php\\?id=(.*)\"data-counter=\".*\" data-counter-id=\".*?\"title=\"(.*)\"class=\".*\"id=\".*\"false\">.*")
var actorImageRegexp, _ = regexp.Compile(".*<img src=\"(.*?)\">.*")
var rankingUrlsRegexp, _ = regexp.Compile(".*<h3><a href=\"/article_search.php\\?id=(.*?)\">(.*?)</a></h3>.*")
var rankingImageRegexp, _ = regexp.Compile(".*<img src=\"(//.*?)\">.*")

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

func (d *FC2) findPage(url string) (string, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
				"referer":         "https://adult.contents.fc2.com",
				"X-PJAX":          true,
				"Accept-Language": "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7,ko;q=0.6,ja;q=0.5",
				"host":            "adult.contents.fc2.com",
			},
		}).Post(d.Addition.SpiderServer)

	if err != nil {
		return "", err
	}

	return res.String(), err
}

func (d *FC2) getFilms(dirName string, urlFunc func(index int) string) ([]model.Obj, error) {

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
		err = db.CreateFilms("fc2", dirName, convertToModel(films, images, urls))
		if err != nil {
			return results, nil
		}
	}

	return d.convertFilm(db.QueryByActor("fc2", dirName), results), nil

}

func (d *FC2) convertFilm(actor []model.Film, results []model.Obj) []model.Obj {
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

func (d *FC2) getMagnet(file model.Obj) (string, error) {

	id := file.GetID()

	res, err := d.findPage(fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", id))
	if err != nil {
		return "", err
	}

	url := subTitles.FindString(res)
	if url == "" {
		return "", nil
	}

	magPage, err := d.findPage(fmt.Sprintf("https://sukebei.nyaa.si%s", subTitles.ReplaceAllString(url, "$1")))
	if err != nil {
		return "", err
	}

	tempMagnet := magnetUrl.FindString(magPage)
	return magnetUrl.ReplaceAllString(tempMagnet, "$1"), nil

}

func (d *FC2) getPageInfo(urlFunc func(index int) string, index int, films []string, images []string, urls []string) ([]string, []string, []string, bool, error) {

	pageUrl := urlFunc(index)

	var urlsRegexp *regexp.Regexp
	var imageRegexp *regexp.Regexp

	if strings.HasPrefix(pageUrl, "https://adult.contents.fc2.com/users") {
		// user
		urlsRegexp = actorUrlsRegexp
		imageRegexp = actorImageRegexp
	} else {
		// ranking
		urlsRegexp = rankingUrlsRegexp
		imageRegexp = rankingImageRegexp
	}

	res, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, false, nil
	}

	tempUrls := urlsRegexp.FindAllString(res, -1)
	imageUrls := imageRegexp.FindAllString(res, -1)

	for _, file := range tempUrls {
		films = append(films, urlsRegexp.ReplaceAllString(file, "$2"))
	}
	for _, imageUrl := range imageUrls {
		images = append(images, "https:"+imageRegexp.ReplaceAllString(imageUrl, "$1"))
	}
	for _, tempUrl := range tempUrls {
		urls = append(urls, urlsRegexp.ReplaceAllString(tempUrl, "$1"))
	}

	return films, images, urls, len(tempUrls) != 0, nil

}
