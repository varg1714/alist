package sexbj_cam

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"regexp"
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

func (d *SexBjCam) findPage(url string) (*resty.Response, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"Host": "sexbjcam.com",
			},
		}).
		Post(d.Addition.TransferServer)

	return res, err
}

func (d *SexBjCam) findRealUrl(url string) (string, error) {

	//log.Infof("开始查询:%s", url)

	hostPattern, err := regexp.Compile("https://(.*\\.com).*")
	if err != nil {
		return "", err
	}
	host := hostPattern.ReplaceAllString(url, "$1")

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"Host":            host,
				"watchsb":         "sbstream",
				"user-agent":      "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0",
				"alt-used":        host,
				"accept-language": "en-US,en;q=0.5",
			},
		}).
		Post(d.Addition.TransferServer)

	if err != nil {
		return "", err
	}

	return string(res.Body()), err
}

func (d *SexBjCam) spider(url string) (string, error) {

	realUrl := fmt.Sprintf("%s?pageUrl=%s&matchUrl=.*.mp4.*&matchUrl=.*/sources.*", d.SpiderServer, url)
	log.Infof("realUrl:%s", realUrl)

	res, err := base.RestyClient.R().Get(realUrl)
	if err != nil {
		return "", err
	}

	return string(res.Body()), err
}

func (d *SexBjCam) getActorFilms(urlFunc func(index int) string) ([]model.Obj, error) {

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

func (d *SexBjCam) getCategoryFilms(urlFunc func(index int) string) ([]model.Obj, error) {

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

	for index := 2; nextPage && index <= 10; index++ {
		films, images, urls, nextPage, err = d.getPageInfo(urlFunc, index, films, images, urls)
		if err != nil {
			return results, err
		}
	}

	return convertToModel(films, images, urls, results), nil

}

func (d *SexBjCam) getLink(file model.Obj) (string, error) {

	pageUrl := file.GetID()

	res, err := d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	page := string(res.Body())

	jumpUrlRegexp, _ := regexp.Compile(".*(https://.*\\.com/e/.*?)\".*")
	jumpUrls := jumpUrlRegexp.FindAllString(page, -1)
	if cap(jumpUrls) <= 0 {
		return "", nil
	}

	jumpUrl := jumpUrlRegexp.ReplaceAllString(jumpUrls[0], "$1")
	playPageUrl, err := d.spider(jumpUrl)
	if err != nil {
		return "", err
	}

	mp4Regexp, err := regexp.Compile(".*.mp4.*")
	if err != nil {
		return "", err
	}
	if mp4Regexp.MatchString(playPageUrl) {
		realUrl := fmt.Sprintf("%s/tapecontent?source=%s", d.PlayServer, playPageUrl)
		return realUrl, err
	}

	realUrlRes, err := d.findRealUrl(playPageUrl)
	if err != nil {
		return "", err
	}

	playPagePattern, err := regexp.Compile(".*\"file\":\"https://.*akamai-video-content.com/(.*?)\".*")
	realUrl := playPagePattern.ReplaceAllString(playPagePattern.FindString(realUrlRes), fmt.Sprintf("%s/akamai/$1", d.PlayServer))

	return realUrl, nil

}

func (d *SexBjCam) getPageInfo(urlFunc func(index int) string, pageNo int, films []string, images []string, urls []string) ([]string, []string, []string, bool, error) {

	urlRegexp, _ := regexp.Compile(".*<article data-video-uid=\".*\" data-post-id=\".*\" class=\".*\">[\\s|.]<a href=\"(.*)\" title=\".*\">.*")
	filmsRegexp, _ := regexp.Compile(".*<div.*class=\".*\"><img.*width=\".*\" height=\".*\".*data-src=\"(.*)\".*alt=\"(.*)\"></div>.*<span.*class=\".*\"><i class=\"fa fa-clock-o\"></i>.*</span>.*</div>.*")

	pageUrl := urlFunc(pageNo)
	//fmt.Printf("开始查询%s\n", pageUrl)

	res, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, false, nil
	}

	page := string(res.Body())

	tempFilms := filmsRegexp.FindAllString(page, -1)
	tempUrls := urlRegexp.FindAllString(page, -1)

	if cap(tempUrls) <= 0 || cap(tempFilms) <= 0 {
		return films, images, urls, false, nil
	}

	for _, file := range tempFilms {
		films = append(films, filmsRegexp.ReplaceAllString(file, "$2"))
		images = append(images, filmsRegexp.ReplaceAllString(file, "$1"))
	}

	for _, tempUrl := range tempUrls {
		urls = append(urls, urlRegexp.ReplaceAllString(tempUrl, "$1"))
	}

	return films, images, urls, true, nil
}
