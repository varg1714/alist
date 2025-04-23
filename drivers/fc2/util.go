package fc2

import (
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/av"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/open_ai"
	"github.com/alist-org/alist/v3/internal/spider"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/tebeka/selenium"
	"regexp"
	"strings"
	"time"
)

var subTitles, _ = regexp.Compile(".*<a href=\"(.*)\" title=\".*</a>.*")
var magnetUrl, _ = regexp.Compile(".*<a href=\"(.*)\" class=\".*\"><i class=\".*\"></i>Magnet</a>.*")

var actorUrlsRegexp, _ = regexp.Compile(".*/article_search.php\\?id=(.*)")

var dateRegexp, _ = regexp.Compile("\\d{4}-\\d{2}-\\d{2}")

func (d *FC2) findMagnet(url string) (string, error) {

	res, err := base.RestyClient.R().Get(url)
	if err != nil {
		return "", err
	}

	return res.String(), err
}

func (d *FC2) getFilms(urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	var result []model.EmbyFileObj
	var filmIds []string
	page := 1
	preSize := len(filmIds)

	for page == 1 || (preSize != len(filmIds)) {

		ids, err2 := d.getPageFilms(urlFunc(page))
		if err2 != nil {
			utils.Log.Warnf("影片爬取失败: %s", err2.Error())
			return result, nil
		} else {
			page++
			preSize = len(filmIds)
			filmIds = append(filmIds, ids...)
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

	code := av.GetFilmCode(file.GetName())

	magnetCache := db.QueryMagnetCacheByCode(code)
	if magnetCache.Magnet != "" {
		utils.Log.Infof("返回缓存中的磁力地址:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	res, err := d.findMagnet(fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", code))
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
		err = db.CreateMagnetCache(model.MagnetCache{
			Magnet: magnet,
			Name:   file.GetName(),
			Code:   code,
		})
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
						Size:     622857143,
					},
					Thumbnail: model.Thumbnail{Thumbnail: image},
				},
				Title: title,
				Url:   id,
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

	fc2Id := code
	if !strings.HasPrefix(fc2Id, "FC2-PPV") {
		fc2Id = fmt.Sprintf("FC2-PPV-%s", code)
	}

	// 1. get cache from db
	magnetCache := db.QueryMagnetCacheByCode(fc2Id)
	if magnetCache.Magnet != "" {
		return model.EmbyFileObj{}, errors.New("已存在该文件")
	}

	// 2. get magnet from suke
	sukeMeta, err := av.GetMetaFromSuke(fc2Id)
	if err != nil {
		utils.Log.Warn("failed to get the magnet info from suke:", err.Error())
		return model.EmbyFileObj{}, err
	} else if len(sukeMeta.Magnets) == 0 || sukeMeta.Magnets[0].GetMagnet() == "" {
		return model.EmbyFileObj{}, errors.New("查询结果为空")
	}

	// 3. translate film name
	title := open_ai.Translate(virtual_file.ClearFilmName(sukeMeta.Magnets[0].GetName()))
	magnet := sukeMeta.Magnets[0].GetMagnet()

	// 4. save film info

	// 4.1 get film thumbnail
	ppvFilmInfo, _ := d.getPpvdbFilm(code)
	if len(ppvFilmInfo.Actors) == 0 {
		ppvFilmInfo.Actors = append(ppvFilmInfo.Actors, "个人收藏")
	}
	if ppvFilmInfo.ReleaseTime.Year() == 1 {
		ppvFilmInfo.ReleaseTime = time.Now()
	}

	// 4.2 build the film info to be cached
	cachingFiles := buildCacheFile(len(sukeMeta.Magnets[0].GetFiles()), fc2Id, title, ppvFilmInfo.ReleaseTime)
	if len(cachingFiles) > 0 {
		cachingFiles[0].Thumbnail.Thumbnail = ppvFilmInfo.Thumb()
	}

	// 4.3 save the magnets info
	var magnetCaches []model.MagnetCache
	for _, file := range cachingFiles {
		magnetCaches = append(magnetCaches, model.MagnetCache{
			DriverType: "fc2",
			Magnet:     magnet,
			Name:       file.Name,
			Code:       av.GetFilmCode(file.Name),
		})
	}
	err = db.BatchCreateMagnetCache(magnetCaches)
	if err != nil {
		utils.Log.Warn("failed to cache film magnet:", err.Error())
		return model.EmbyFileObj{}, err
	}

	// 4.4 save the film info
	err = db.CreateFilms("fc2", "个人收藏", "个人收藏", cachingFiles)
	if err != nil {
		utils.Log.Warn("failed to cache film info:", err.Error())
		return model.EmbyFileObj{}, err
	}

	// 4.5 save the film meta, including nfo and images
	_ = virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
		Source:   "fc2",
		Dir:      "个人收藏",
		FileName: virtual_file.AppendImageName(cachingFiles[0].Name),
		Title:    title,
		ImgUrl:   ppvFilmInfo.Thumb(),
		Actors:   ppvFilmInfo.Actors,
		Release:  ppvFilmInfo.ReleaseTime,
	})

	var noImageFiles []model.EmbyFileObj
	for _, file := range cachingFiles {
		if file.Thumb() == "" {
			noImageFiles = append(noImageFiles, file)
		}
	}
	if len(noImageFiles) > 0 {

		whatLinkInfo := d.getWhatLinkInfo(magnet)
		imgs := whatLinkInfo.Screenshots
		if len(imgs) > 0 {
			for index, file := range noImageFiles {
				_ = virtual_file.CacheImage(virtual_file.MediaInfo{
					Source:   "fc2",
					Dir:      "个人收藏",
					FileName: virtual_file.AppendImageName(file.Name),
					Title:    title,
					ImgUrl:   imgs[index%len(imgs)].Screenshot,
					ImgUrlHeaders: map[string]string{
						"Referer": "https://mypikpak.com/",
					},
					Actors:  ppvFilmInfo.Actors,
					Release: ppvFilmInfo.ReleaseTime,
				})
			}
		}

	}

	return cachingFiles[0], err

}

func buildCacheFile(fileCount int, fc2Id string, title string, releaseTime time.Time) []model.EmbyFileObj {

	var cachingFiles []model.EmbyFileObj
	if fileCount <= 1 {
		cachingFiles = append(cachingFiles, model.EmbyFileObj{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					Name:     virtual_file.AppendFilmName(fc2Id),
					IsFolder: false,
					Size:     622857143,
					Modified: time.Now(),
					Path:     "个人收藏",
				},
			},
			Title:       title,
			ReleaseTime: releaseTime,
			Url:         fc2Id,
		})
	} else {
		for index := range fileCount {
			realName := virtual_file.AppendFilmName(fmt.Sprintf("%s-cd%d", fc2Id, index+1))
			cachingFiles = append(cachingFiles, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name:     realName,
						IsFolder: false,
						Size:     622857143,
						Modified: time.Now(),
						Path:     "个人收藏",
					},
				},
				Title:       title,
				ReleaseTime: releaseTime,
				Url:         fc2Id,
			})
		}
	}
	return cachingFiles
}

func (d *FC2) getPpvdbFilm(code string) (model.EmbyFileObj, error) {

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

	var releaseTime time.Time
	collector.OnHTML("div[class*='lg:pl-8'][class*='lg:w-3/5']", func(element *colly.HTMLElement) {

		element.ForEach("span", func(i int, spanElement *colly.HTMLElement) {
			timeStr := spanElement.Text
			if dateRegexp.MatchString(timeStr) {
				tempTime, err := time.Parse("2006-01-02", timeStr)
				if err == nil {
					releaseTime = tempTime
				} else {
					utils.Log.Infof("failed to parse release time:%s,error message:%v", timeStr, err)
				}
			}
		})

	})

	title := ""
	collector.OnHTML(".items-center.text-white.text-lg.title-font.font-medium.mb-1 a", func(element *colly.HTMLElement) {
		title = element.Text
	})

	err := collector.Visit(fmt.Sprintf("https://fc2ppvdb.com/articles/%s", code))
	if err != nil {
		utils.Log.Infof("failed to query fc2 film info for:[%s], error message:%s", code, err.Error())
		return model.EmbyFileObj{}, err
	}

	for actor, _ := range actorMap {
		actors = append(actors, actor)
	}

	return model.EmbyFileObj{
		ObjThumb: model.ObjThumb{
			Object: model.Object{
				IsFolder: false,
				Name:     title,
			},
			Thumbnail: model.Thumbnail{Thumbnail: imageUrl},
		},
		Title:       title,
		Actors:      actors,
		ReleaseTime: releaseTime,
	}, nil

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

func (d *FC2) refreshNfo() {

	utils.Log.Info("start refresh nfo for fc2")

	films := d.getStars()
	fileNames := make(map[string][]string)

	for _, film := range films {
		virtual_file.UpdateNfo(virtual_file.MediaInfo{
			Source:   "fc2",
			Dir:      film.Path,
			FileName: virtual_file.AppendImageName(film.Name),
			Release:  film.ReleaseTime,
			Title:    film.Title,
			Actors:   film.Actors,
		})
		fileNames[film.Path] = append(fileNames[film.Path], film.Name)
	}

	// clear unused files
	for dir, names := range fileNames {
		virtual_file.ClearUnUsedFiles("fc2", dir, names)
	}

	utils.Log.Info("finish refresh nfo")
}

func (d *FC2) reMatchReleaseTime() {

	// rematch release time

	utils.Log.Infof("start rematching release time for fc2")

	incompleteFilms, err := db.QueryIncompleteFilms("fc2")

	if err != nil {
		utils.Log.Warnf("failed to query no date films: %s", err.Error())
		return
	}

	filmMap := make(map[string]model.Film)

	for _, film := range incompleteFilms {

		code := av.GetFilmCode(film.Name)

		if existFilm, exist := filmMap[code]; exist {
			if film.Title == "" {
				film.Title = existFilm.Title
			}
			if len(film.Actors) == 0 {
				if len(existFilm.Actors) > 0 {
					film.Actors = append(film.Actors, existFilm.Actors...)
				}
			}
		} else {

			ppvdbMediaInfo, err1 := d.getPpvdbFilm(code)
			if err1 != nil {
				if strings.Contains(err1.Error(), "Not Found") {
					film.Actors = []string{"个人收藏"}
				} else {
					return
				}
			} else {
				if ppvdbMediaInfo.ReleaseTime.Year() != 1 {
					film.Date = ppvdbMediaInfo.ReleaseTime
				} else {
					film.Date = film.CreatedAt
				}

				if film.Title == "" && ppvdbMediaInfo.Title != "" {
					film.Title = open_ai.Translate(ppvdbMediaInfo.Title)
				}

				if len(film.Actors) == 0 {
					if len(ppvdbMediaInfo.Actors) > 0 {
						film.Actors = ppvdbMediaInfo.Actors
					} else {
						film.Actors = []string{"个人收藏"}
					}
				}
			}

		}

		if film.Title == "" {
			sukeMediaInfo, err2 := av.GetMetaFromSuke(code)
			if err2 != nil {
				utils.Log.Warnf("failed to query suke: %s", code)
			} else if len(sukeMediaInfo.Magnets) > 0 {
				film.Title = open_ai.Translate(sukeMediaInfo.Magnets[0].GetName())
			}
		}
		filmMap[code] = film

		err1 := db.UpdateFilm(film)
		if err1 != nil {
			utils.Log.Warnf("failed to update film info: %s", err1.Error())
		}

		// avoid 429
		time.Sleep(time.Duration(d.ScanTimeLimit) * time.Second)

	}

	utils.Log.Info("rematching completed")

}

func (d *FC2) getPageFilms(url string) ([]string, error) {

	var ids []string

	err := spider.Visit(d.SpiderServer, url, time.Duration(d.SpiderMaxWaitTime)*time.Second, func(wd selenium.WebDriver) {
		elements, _ := wd.FindElements(selenium.ByCSSSelector, ".absolute.top-0.left-0.text-white.bg-gray-800.px-1")
		for _, element := range elements {
			text, err1 := element.Text()
			if err1 != nil {
				utils.Log.Warnf("failed to fetch element: %s", err1.Error())
			} else {
				ids = append(ids, fmt.Sprintf("FC2-PPV-%s", text))
			}
		}
	})

	return ids, err

}
