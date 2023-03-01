package miss_av

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	js "github.com/dop251/goja"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
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

func (d *MIssAV) findPage(url string) (*resty.Response, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"Host": "https://missav.com/",
			},
		}).
		Post(d.Addition.SpiderServer)

	return res, err
}

func (d *MIssAV) getFilms(urlFunc func(index int) string) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	films := make([]string, 0)
	images := make([]string, 0)
	urls := make([]string, 0)
	var err error

	films, images, urls, err = d.getPageInfo(urlFunc, 1, films, images, urls)
	if err != nil {
		return results, err
	}

	for index := 2; index <= 20; index++ {
		films, images, urls, err = d.getPageInfo(urlFunc, index, films, images, urls)
		if err != nil {
			return results, err
		}
	}

	return convertToModel(films, images, urls, results), nil

}

func (d *MIssAV) getLink(file model.Obj) (string, error) {

	pageUrl := file.GetID()

	res, err := d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	page := string(res.Body())

	urlScriptRegexp, _ := regexp.Compile(".*(eval\\(.*\\)).*")
	urlScript := urlScriptRegexp.FindAllString(page, -1)
	if cap(urlScript) <= 0 {
		return "", nil
	}

	runString, err := js.New().RunString(urlScript[0])
	//log.Infof("js计算脚本%s", urlScript)
	if err != nil {
		log.Errorf("js计算错误%s", err)
		return "", err
	}

	realUrl := runString.Export().(string)
	//log.Infof("js计算访问地址:%s", realUrl)

	return strings.ReplaceAll(realUrl, d.Addition.PlayServer, d.Addition.PlayProxyServer), nil

}

func (d *MIssAV) getPageInfo(urlFunc func(index int) string, index int, films []string, images []string, urls []string) ([]string, []string, []string, error) {

	filmsRegexp, _ := regexp.Compile("<a class=\"text-secondary group-hover:text-primary\" href=\"(.*)\">\\s?(.*)\\s?</a>")
	imageRegexp, _ := regexp.Compile("<img x-cloak :class=\".*\" class=\".*\" data-src=\"(.*)\" src=\".*\" alt=\".*\">")

	pageUrl := urlFunc(index)
	//log.Infof("开始查询%s", pageUrl)

	res, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, nil
	}

	page := string(res.Body())

	tempFilms := filmsRegexp.FindAllString(page, -1)
	imageUrls := imageRegexp.FindAllString(page, -1)

	for _, file := range tempFilms {
		urls = append(urls, filmsRegexp.ReplaceAllString(file, "$1"))
		films = append(films, filmsRegexp.ReplaceAllString(file, "$2"))
	}

	for _, tempUrl := range imageUrls {
		images = append(images, imageRegexp.ReplaceAllString(tempUrl, "$1"))
	}

	return films, images, urls, nil
}
