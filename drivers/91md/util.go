package _91_md

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"time"
)

func convertToModel(films []string, images []string, urls []string, results []model.Obj) []model.Obj {
	for index, film := range films {

		var image string
		if index < cap(images) {
			image = images[index]
		}
		//log.Infof("index:%s,image:%s,cap:%s,images:%s\n", index, image, cap(images), images)

		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     fmt.Sprintf("%04d", index) + " " + film + ".mp4",
				IsFolder: false,
				ID:       urls[index],
				Size:     622857143,
				Modified: time.Now(),
			},
			Thumbnail: model.Thumbnail{Thumbnail: image},
		})
	}
	return results
}

func (d *_91MD) findPage(url string) (string, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"Host": "https://91md.me",
			},
		}).
		Post(d.Addition.TransferServer)

	if err != nil {
		return "", err
	}

	return string(res.Body()), err
}

func (d *_91MD) getActorFilms(urlFunc func(index int) string) ([]model.Obj, error) {

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

	for index := 2; nextPage; index++ {
		films, images, urls, nextPage, err = d.getPageInfo(urlFunc, index, films, images, urls)
		if err != nil {
			return results, err
		}
	}

	return convertToModel(films, images, urls, results), nil

}

func (d *_91MD) getCategoryFilms(urlFunc func(index int) string) ([]model.Obj, error) {

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

	for index := 2; nextPage && index < 3; index++ {
		films, images, urls, nextPage, err = d.getPageInfo(urlFunc, index, films, images, urls)
		if err != nil {
			return results, err
		}
	}

	return convertToModel(films, images, urls, results), nil

}

func (d *_91MD) getLink(file model.Obj) (string, error) {

	pageUrl := file.GetID()

	page, err := d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	jumpUrlRegexp, _ := regexp.Compile(".*\"url\":\"(https:.*?)\".*")
	realUrlRegexp, _ := regexp.Compile("\\\\")

	page, err = d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	encodedUrl := jumpUrlRegexp.ReplaceAllString(jumpUrlRegexp.FindString(page), "$1")
	return realUrlRegexp.ReplaceAllString(encodedUrl, ""), nil

}

func (d *_91MD) getPageInfo(urlFunc func(index int) string, pageNo int, films []string, images []string, urls []string) ([]string, []string, []string, bool, error) {

	urlRegexp, _ := regexp.Compile(".*<p class=\"img\">[.|\\s]*?<img class=\".*\" src=\"(.*?)\" alt=\".*?\" title=\"(.*?)\">[.|\\s]*?<a href=\"(.*?)\"></a>[.|\\s]*?</p>.*")
	hostRegexp, _ := regexp.Compile("(https://.*?)/.*")
	nextPageRegexp, _ := regexp.Compile(".*<a class=\"pagelink_b\" href=\".*?/page/(\\d+).*\" title=\"下一页\">下一页</a>.*")

	pageUrl := urlFunc(pageNo)

	page, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, false, nil
	}

	tempUrls := urlRegexp.FindAllString(page, -1)
	hostPrefix := hostRegexp.ReplaceAllString(hostRegexp.FindString(pageUrl), "$1")

	for _, file := range tempUrls {
		films = append(films, urlRegexp.ReplaceAllString(file, "$2"))
		images = append(images, urlRegexp.ReplaceAllString(file, "$1"))
		urls = append(urls, fmt.Sprintf("%s/%s", hostPrefix, urlRegexp.ReplaceAllString(file, "$3")))
	}

	nextPage, err := strconv.Atoi(nextPageRegexp.ReplaceAllString(nextPageRegexp.FindString(page), "$1"))
	if err != nil {
		log.Warnf("%s:分页信息获取失败%s", pageUrl, err)
		return films, images, urls, false, nil
	}

	return films, images, urls, pageNo < nextPage, nil

}
