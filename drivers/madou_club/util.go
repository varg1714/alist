package jable_tv

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
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

func (d *MaDouClub) findPage(url string) (*resty.Response, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
			"headers": base.Json{
				"Host": "https://madou.club",
			},
		}).
		Post(d.Addition.SpiderServer)

	return res, err
}

func (d *MaDouClub) getFilms(urlFunc func(index int) string) ([]model.Obj, error) {

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

func (d *MaDouClub) getLink(file model.Obj) (string, error) {

	pageUrl := file.GetID()

	res, err := d.findPage(pageUrl)
	if err != nil {
		return "", err
	}

	page := string(res.Body())

	jumpUrlRegexp, _ := regexp.Compile(".*<p><iframe height=.* width=.* src=(.*) frameborder=0 allowfullscreen></iframe></p>.*")
	jumpUrl := jumpUrlRegexp.FindAllString(page, -1)
	if cap(jumpUrl) <= 0 {
		return "", nil
	}

	res, err = d.findPage(jumpUrlRegexp.ReplaceAllString(jumpUrl[0], "$1"))
	if err != nil {
		return "", err
	}

	page = string(res.Body())

	tokenRegexp, _ := regexp.Compile(".*var token = \"(.*)\".*")
	m3u8Regexp, _ := regexp.Compile(".*var m3u8 = '(.*)'.*")

	token := tokenRegexp.FindAllString(page, -1)
	m3u8 := m3u8Regexp.FindAllString(page, -1)

	if cap(token) <= 0 || cap(m3u8) <= 0 {
		log.Info("地址未获取到")
		return "", err
	}

	return fmt.Sprintf(
			"https://dash.madou.club%s?token=%s",
			m3u8Regexp.ReplaceAllString(m3u8[0], "$1"),
			tokenRegexp.ReplaceAllString(token[0], "$1")),
		nil

}

func (d *MaDouClub) getPageInfo(urlFunc func(index int) string, pageNo int, films []string, images []string, urls []string) ([]string, []string, []string, bool, error) {

	filmsRegexp, _ := regexp.Compile("<a target=\"_blank\" href=\".*?\">(.*?)</a>")
	imagesRegexp, _ := regexp.Compile("<img src=\".*?\" data-src=\"(.*?)\" class=\"thumb\">")
	urlRegexp, _ := regexp.Compile("<a target=\"_blank\" class=\"thumbnail\" href=\"(.*?)\">")

	pageUrl := urlFunc(pageNo)
	//fmt.Printf("开始查询%s\n", pageUrl)

	res, err := d.findPage(pageUrl)
	if err != nil {
		return films, images, urls, false, nil
	}

	page := string(res.Body())

	tempFilms := filmsRegexp.FindAllString(page, -1)
	tempImages := imagesRegexp.FindAllString(page, -1)
	tempUrls := urlRegexp.FindAllString(page, -1)

	for _, file := range tempFilms {
		films = append(films, filmsRegexp.ReplaceAllString(file, "$1"))
	}
	for _, image := range tempImages {
		images = append(images, imagesRegexp.ReplaceAllString(image, "$1"))
	}

	for _, tempUrl := range tempUrls {
		urls = append(urls, urlRegexp.ReplaceAllString(tempUrl, "$1"))
	}

	nextPage := strings.Index(page, "下一页")
	return films, images, urls, nextPage != -1, nil
}
