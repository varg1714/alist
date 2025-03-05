package fc2

import (
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/open_ai"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly/v2"
	"regexp"
	"strings"
	"time"
)

var subTitles, _ = regexp.Compile(".*<a href=\"(.*)\" title=\".*</a>.*")
var magnetUrl, _ = regexp.Compile(".*<a href=\"(.*)\" class=\".*\"><i class=\".*\"></i>Magnet</a>.*")

var actorUrlsRegexp, _ = regexp.Compile(".*/article_search.php\\?id=(.*)")

func (d *FC2) findMagnet(url string) (string, error) {

	res, err := base.RestyClient.R().Get(url)
	if err != nil {
		return "", err
	}

	return res.String(), err
}

func (d *FC2) getFilms(urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	var result []model.EmbyFileObj
	var filmIds []string
	page := 1
	preSize := len(filmIds)

	collector.OnHTML(".flex.flex-wrap.-m-4.pb-4", func(e *colly.HTMLElement) {

		e.ForEach(".absolute.top-0.left-0.text-white.bg-gray-800.px-1", func(i int, film *colly.HTMLElement) {
			filmIds = append(filmIds, fmt.Sprintf("FC2-PPV-%s", film.Text))
		})

	})

	for page == 1 || (preSize != len(filmIds)) {

		err := collector.Visit(urlFunc(page))
		if err != nil {
			utils.Log.Warnf("影片爬取失败: %s", err.Error())
			return result, nil
		} else {
			page++
			preSize = len(filmIds)
		}

	}

	unCachedFilms := db.QueryUnCachedFilms(filmIds)
	if len(unCachedFilms) == 0 {
		return result, nil
	}

	unMissedFilms := db.QueryUnMissedFilms(unCachedFilms)
	if len(unMissedFilms) == 0 {
		return result, nil
	}

	utils.Log.Infof("以下影片首次扫描到需添加入库：%v", unCachedFilms)
	var notExitedFilms []string
	for _, id := range unCachedFilms {
		_, err := d.addStar(id)
		if err != nil {
			notExitedFilms = append(notExitedFilms, id)
		}
	}

	if len(notExitedFilms) > 0 {
		utils.Log.Infof("以下影片未获取到下载信息：%v", notExitedFilms)
		err := db.CreateMissedFilms(notExitedFilms)
		if err != nil {
			utils.Log.Warnf("影片信息保存失败: %s", err.Error())
		}
	}

	return result, nil

}

func (d *FC2) getMagnet(file model.Obj) (string, error) {

	magnetCache := db.QueryFileCacheByCode(file.GetID())
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
		err = db.CreateCacheFile(magnet, "", id)
	}

	return magnet, err

}

func (d *FC2) getPageInfo(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {

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
			title = fmt.Sprintf("FC2-PPV-%s %s", id, title)
			data = append(data, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name:     title,
						IsFolder: true,
						ID:       id,
						Size:     622857143,
					},
					Thumbnail: model.Thumbnail{Thumbnail: image},
				},
				Title: title,
			})
		})
	})

	err := collector.Visit(pageUrl)
	if err != nil && err.Error() == "Not Found" {
		err = nil
	}

	return data, len(data) != preLen, err

}

func (d *FC2) getStars() []model.EmbyFileObj {
	return virtual_file.GetStorageFilms("fc2", "个人收藏", false)
}

func (d *FC2) addStar(code string) (model.EmbyFileObj, error) {

	id := code
	if !strings.HasPrefix(id, "FC2-PPV") {
		id = fmt.Sprintf("FC2-PPV-%s", code)
	}

	magnetCache := db.QueryFileCacheByCode(id)
	if magnetCache.Magnet != "" {
		return model.EmbyFileObj{}, errors.New("已存在该文件")
	}

	searchUrl := fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", id)

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	title := ""
	detailUrl := ""

	collector.OnHTML(`.table-responsive td[colspan="2"]`, func(element *colly.HTMLElement) {
		if title == "" {

			element.ForEach("a", func(i int, aElement *colly.HTMLElement) {

				if attr := aElement.Attr("class"); attr != "comments" {
					title = strings.ReplaceAll(aElement.Attr("title"), "+++ ", "")
					href := aElement.Attr("href")
					if href != "" {
						detailUrl = fmt.Sprintf("https://sukebei.nyaa.si/%s", href)
					}
				}

			})

		}
	})

	err := collector.Visit(searchUrl)
	if err != nil {
		return model.EmbyFileObj{}, err
	}

	if detailUrl == "" {
		return model.EmbyFileObj{}, errors.New("查询结果为空")
	}

	title = open_ai.Translate(virtual_file.ClearFilmName(title))
	magnet := ""

	collector.OnHTML(".card-footer-item", func(element *colly.HTMLElement) {
		magnet = element.Attr("href")
	})

	fileCount := 0
	collector.OnHTML(".torrent-file-list.panel-body", func(element *colly.HTMLElement) {
		element.ForEach(".file-size", func(i int, liElement *colly.HTMLElement) {
			text := liElement.Text
			if len(text) > 2 {
				bytes, _ := humanize.ParseBytes(text[1 : len(text)-1])
				if bytes/(1024*1024) > 100 {
					fileCount++
				}
			}
		})
	})

	err = collector.Visit(detailUrl)
	if err != nil {
		return model.EmbyFileObj{}, err
	}

	thumbnail, actors := d.getPpvdbFilm(code)
	if len(actors) == 0 {
		actors = append(actors, "个人收藏")
	}

	var cachingFiles []model.EmbyFileObj
	if fileCount <= 1 {
		cachingFiles = append(cachingFiles, model.EmbyFileObj{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					Name:     virtual_file.AppendFilmName(id),
					IsFolder: false,
					ID:       id,
					Size:     622857143,
					Modified: time.Now(),
					Path:     "个人收藏",
				},
				Thumbnail: model.Thumbnail{Thumbnail: thumbnail},
			},
			Title: title})
	} else {
		for index := range fileCount {
			realName := virtual_file.AppendFilmName(fmt.Sprintf("%s-cd%d", id, index+1))
			cachingFiles = append(cachingFiles, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name:     realName,
						IsFolder: false,
						ID:       realName,
						Size:     622857143,
						Modified: time.Now(),
						Path:     "个人收藏",
					},
					Thumbnail: model.Thumbnail{Thumbnail: thumbnail},
				},
				Title: title})
		}
	}

	// 缓存磁力
	for _, file := range cachingFiles {
		err = db.CreateCacheFile(magnet, "", file.ID)
	}

	// 保存影片信息
	err = db.CreateFilms("fc2", "个人收藏", "个人收藏", cachingFiles)
	_ = virtual_file.CacheImageAndNfo("fc2", "个人收藏", virtual_file.AppendImageName(cachingFiles[0].Name), title, thumbnail, actors)

	if len(cachingFiles) > 1 {

		whatLinkInfo := d.getWhatLinkInfo(magnet)

		cachingImageFiles := cachingFiles
		if thumbnail != "" {
			cachingImageFiles = cachingFiles[1:]
		}

		for index, file := range cachingImageFiles {
			if index < len(whatLinkInfo.Screenshots) {
				_ = virtual_file.CacheImage("fc2", "个人收藏", virtual_file.AppendImageName(file.Name), whatLinkInfo.Screenshots[index].Screenshot, map[string]string{
					"Referer": "https://mypikpak.com/",
				})
			} else {
				if thumbnail == "" && len(whatLinkInfo.Screenshots) > 0 {
					thumbnail = whatLinkInfo.Screenshots[0].Screenshot
				}
				_ = virtual_file.CacheImage("fc2", "个人收藏", virtual_file.AppendImageName(file.Name), thumbnail, map[string]string{})
			}
		}

	} else if len(cachingFiles) == 1 && thumbnail == "" {
		whatLinkInfo := d.getWhatLinkInfo(magnet)
		if len(whatLinkInfo.Screenshots) > 0 {
			_ = virtual_file.CacheImage("fc2", "个人收藏", virtual_file.AppendImageName(cachingFiles[0].Name), whatLinkInfo.Screenshots[0].Screenshot, map[string]string{
				"Referer": "https://mypikpak.com/",
			})
		}
	}

	return cachingFiles[0], err

}

func (d *FC2) getPpvdbFilm(code string) (string, []string) {

	split := strings.Split(code, "-")
	code = split[len(split)-1]

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	imageUrl := ""

	var actors []string
	actorMap := make(map[string]bool)

	collector.OnHTML(fmt.Sprintf(`img[alt="%s"]`, code), func(element *colly.HTMLElement) {

		srcImage := element.Attr("src")
		if srcImage != "" && !strings.Contains(srcImage, "no-image.jpg") {
			imageUrl = srcImage
		}
	})

	collector.OnHTML(".text-white.title-font.text-lg.font-medium", func(element *colly.HTMLElement) {
		title := element.Attr("title")
		if title != "" {
			actorMap[title] = true
		}

	})

	err := collector.Visit(fmt.Sprintf("https://fc2ppvdb.com/articles/%s", code))
	if err != nil {
		utils.Log.Infof("影片:%s的缩略图获取失败:%s", code, err.Error())
	}

	for actor, _ := range actorMap {
		actors = append(actors, actor)
	}

	return imageUrl, actors

}

func (d *FC2) getWhatLinkInfo(magnet string) WhatLinkInfo {

	var whatLinkInfo WhatLinkInfo

	_, err := base.RestyClient.R().SetHeaders(map[string]string{
		"Referer": "https://mypikpak.net/",
		"Origin":  "https://mypikpak.net/",
	}).SetQueryParam("url", magnet).SetResult(&whatLinkInfo).Get("https://whatslink.info/api/v1/link")

	if err != nil {
		utils.Log.Info("磁力图片获取失败", err.Error())
		return whatLinkInfo
	}

	return whatLinkInfo

}
