package fc2

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/gocolly/colly/v2"
	"regexp"
	"strings"
	"time"
)

var subTitles, _ = regexp.Compile(".*<a href=\"(.*)\" title=\".*</a>.*")
var magnetUrl, _ = regexp.Compile(".*<a href=\"(.*)\" class=\".*\"><i class=\".*\"></i>Magnet</a>.*")

var actorUrlsRegexp, _ = regexp.Compile(".*/article_search.php\\?id=(.*)")

func (d *FC2) findMagnet(url string) (string, error) {

	res, err := base.RestyClient.R().
		SetBody(base.Json{
			"url":        url,
			"httpMethod": "GET",
		}).Post(d.Addition.SpiderServer)

	if err != nil {
		return "", err
	}

	return res.String(), err
}

func (d *FC2) getFilms(dirName string, urlFunc func(index int) string) ([]model.ObjThumb, error) {

	if strings.HasPrefix(urlFunc(1), "https://adult.contents.fc2.com/users") {
		return virtual_file.GetFilmsWitchStorage("fc2", dirName, dirName, urlFunc,
			func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
				return d.getPageInfo(urlFunc, index, data)
			}, virtual_file.Option{CacheFile: true, MaxPageNum: 20})
	} else {
		return virtual_file.GetFilms("fc2", dirName, urlFunc,
			func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
				return d.getPageInfo(urlFunc, index, data)
			})
	}

}

func (d *FC2) getMagnet(file model.Obj) (string, error) {

	magnetCache := db.QueryCacheFileId(file.GetName())
	if magnetCache.Magnet != "" {
		utils.Log.Infof("返回缓存中的磁力地址:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	id := file.GetID()

	res, err := d.findMagnet(fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", id))
	if err != nil {
		return "", err
	}

	url := subTitles.FindString(res)
	if url == "" {
		return "", nil
	}

	magPage, err := d.findMagnet(fmt.Sprintf("https://sukebei.nyaa.si%s", subTitles.ReplaceAllString(url, "$1")))
	if err != nil {
		return "", err
	}

	tempMagnet := magnetUrl.FindString(magPage)
	magnet := magnetUrl.ReplaceAllString(tempMagnet, "$1")

	if magnet != "" {
		err = db.CreateCacheFile(magnet, "", file.GetName())
	}

	return magnet, err

}

func (d *FC2) getPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	pageUrl := urlFunc(index)
	preLen := len(data)

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	tableContainer := ""
	filmDetailContainer := ""
	filmUrlSelector := ""
	filmTitleSelector := ""
	filmImageSelector := ""

	if strings.HasPrefix(pageUrl, "https://adult.contents.fc2.com/users") {
		// user
		tableContainer = ".seller_user_articlesList"
		filmDetailContainer = ".c-cntCard-110-f"
		filmUrlSelector = ".c-cntCard-110-f_itemName"
		filmTitleSelector = ".c-cntCard-110-f_itemName"
		filmImageSelector = ".c-cntCard-110-f_thumb img"
	} else {
		// ranking
		tableContainer = ".c-rankbox-100"
		filmDetailContainer = ".c-ranklist-110"
		filmUrlSelector = ".c-ranklist-110_tmb a"
		filmTitleSelector = ".c-ranklist-110_info a"
		filmImageSelector = ".c-ranklist-110_tmb img"
	}

	collector.OnHTML(tableContainer, func(element *colly.HTMLElement) {
		element.ForEach(filmDetailContainer, func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr(filmUrlSelector, "href")
			title := element.ChildText(filmTitleSelector)

			var image string
			imageAttr := element.ChildAttr(filmImageSelector, "src")
			if strings.HasPrefix(imageAttr, "http") {
				image = imageAttr
			} else {
				image = "https:" + imageAttr
			}

			id := actorUrlsRegexp.ReplaceAllString(href, "$1")
			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     virtual_file.CutString(fmt.Sprintf("FC2-PPV-%s %s", id, title)),
					IsFolder: true,
					ID:       id,
					Size:     622857143,
				},
				Thumbnail: model.Thumbnail{Thumbnail: image},
			})
		})
	})

	err := collector.Visit(pageUrl)
	if err != nil && err.Error() == "Not Found" {
		err = nil
	}

	return data, len(data) != preLen, err

}
