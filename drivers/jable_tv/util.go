package jable_tv

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func findFilms(pageResponse string) ([]string, []string, int) {

	films := make([]string, 0)
	images := make([]string, 0)
	pages := 1

	strSplits := strings.Split(pageResponse, "\n")

	fileRegex, _ := regexp.Compile(".*<h6 class=\"title\"><a href=\".*\">(.*)</a></h6>.*")
	imageRegex, _ := regexp.Compile(".*data-src=\"(.*)\".*")
	filesNumberRegex, _ := regexp.Compile(".*?(\\d+) 部影片.*")

	for _, split := range strSplits {

		if fileRegex.MatchString(split) {
			film := fileRegex.ReplaceAllString(split, "$1")
			films = append(films, film)
			//log.Infof("影片正则匹配:%s\n", film)

		}

		if imageRegex.MatchString(split) {
			image := imageRegex.ReplaceAllString(split, "$1")
			images = append(images, image)
			//log.Infof("预览图正则匹配:%s\n", image)
		}

		if filesNumberRegex.MatchString(split) {
			filesNumber, err := strconv.Atoi(filesNumberRegex.ReplaceAllString(split, "$1"))
			if err != nil {
				log.Warnf("最大页码提取错误:%s\n", split)
			} else {
				pages = filesNumber
			}
		}

	}
	return films, images, pages
}

func (d *JableTV) findPage(url string) (*resty.Response, error) {

	//log.Infof("开始查询:%s", url)

	res, err := base.RestyClient.R().
		Get(fmt.Sprintf("%s%s", d.Addition.SpiderServer, url))

	return res, err
}

func (d *JableTV) getActorFilms(actorName string, results []model.Obj) ([]model.Obj, error) {

	url := fmt.Sprintf("https://jable.tv/models/%s/?mode=async&function=get_block&block_id="+
		"list_videos_common_videos_list&sort_by=post_date&from=%d", actorName, 1)
	res, err := d.findPage(url)

	if err != nil {
		log.Errorf("出错了：%s,%s\n", err, res)
		return results, err
	}

	pageResponse := string(res.Body())
	films, images, pages := findFilms(pageResponse)
	log.Infof("最大页码：%s", pages)

	for index := 2; index <= ((pages-1)/24 + 1); index++ {

		url = fmt.Sprintf("https://jable.tv/models/%s/?mode=async&function=get_block&block_id="+
			"list_videos_common_videos_list&sort_by=post_date&from=%d", actorName, index)
		res, err := d.findPage(url)
		if err != nil {
			log.Errorf("出错了：%s,%s\n", err, res)
			return results, err
		}
		pageResponse = string(res.Body())

		tempFilms, tempImages, _ := findFilms(pageResponse)
		films = append(films, tempFilms...)
		images = append(images, tempImages...)

	}

	// log.Infof("结果:%s,%s\n", films, images)

	results = convertToModel(films, images, results)
	return results, nil
}

func (d *JableTV) getFilms(urlFunc func(index int) string) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	films := make([]string, 0)
	images := make([]string, 0)

	for index := 1; index <= 5; index++ {

		url := urlFunc(index)

		//log.Infof("获取地址:%s", url)
		page, err := d.findPage(url)
		if err != nil {
			return results, err
		}

		tempFilms, tempImages, _ := findFilms(string(page.Body()))
		films = append(films, tempFilms...)
		images = append(images, tempImages...)

	}

	return convertToModel(films, images, results), nil

}

func convertToModel(films []string, images []string, results []model.Obj) []model.Obj {
	for index, film := range films {

		var image string
		if index < cap(images) {
			image = images[index]
		}
		//log.Infof("index:%s,image:%s,cap:%s,images:%s\n", index, image, cap(images), images)

		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     fmt.Sprintf("%03d", index) + " " + film + ".mp4",
				IsFolder: false,
				ID:       film,
				Size:     622857143,
				Modified: time.Now(),
			},
			Thumbnail: model.Thumbnail{Thumbnail: image},
		})
	}
	return results
}
