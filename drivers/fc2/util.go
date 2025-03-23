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

	magnetCache := db.QueryMagnetCacheByCode(file.GetID())
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
		err = db.CreateMagnetCache(model.MagnetCache{
			Magnet: magnet,
			Name:   file.GetName(),
			Code:   id,
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

	// 1. get cache from db
	magnetCache := db.QueryMagnetCacheByCode(id)
	if magnetCache.Magnet != "" {
		return model.EmbyFileObj{}, errors.New("已存在该文件")
	}

	// 2. get magnet from suke
	sukeMeta, err := av.GetMetaFromSuke(id)
	if err != nil {
		utils.Log.Warn("failed to get the magnet info from suke:", err.Error())
		return model.EmbyFileObj{}, err
	} else if len(sukeMeta.Magnets) == 0 || sukeMeta.Magnets[0].Magnet == "" {
		return model.EmbyFileObj{}, errors.New("查询结果为空")
	}

	// 3. translate film name
	title := open_ai.Translate(virtual_file.ClearFilmName(sukeMeta.Magnets[0].Name))
	magnet := sukeMeta.Magnets[0].Magnet

	// 4. save film info

	// 4.1 get film thumbnail
	thumbnail, actors, releaseTime := d.getPpvdbFilm(code)
	if len(actors) == 0 {
		actors = append(actors, "个人收藏")
	}

	// 4.2 build the film info to be cached
	cachingFiles := buildCacheFile(len(sukeMeta.Magnets[0].Files), id, title)
	if len(cachingFiles) > 0 {
		cachingFiles[0].Thumbnail.Thumbnail = thumbnail
	}

	// 4.3 save the magnets info
	var magnetCaches []model.MagnetCache
	for _, file := range cachingFiles {
		magnetCaches = append(magnetCaches, model.MagnetCache{
			DriverType: "fc2",
			Magnet:     magnet,
			Name:       file.Name,
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
		ImgUrl:   thumbnail,
		Actors:   actors,
		Release:  releaseTime,
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
					Actors:  actors,
					Release: releaseTime,
				})
			}
		}

	}

	return cachingFiles[0], err

}

func buildCacheFile(fileCount int, id string, title string) []model.EmbyFileObj {

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
				},
				Title: title})
		}
	}
	return cachingFiles
}

func (d *FC2) getPpvdbFilm(code string) (string, []string, time.Time) {

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
		timeStr := element.ChildText("div:nth-child(6) span")
		if timeStr != "" {
			tempTime, err := time.Parse("2006-01-02", timeStr)
			if err == nil {
				releaseTime = tempTime
			} else {
				utils.Log.Infof("failed to parse release time:%s,error message:%v", timeStr, err)
			}
		}

	})

	err := collector.Visit(fmt.Sprintf("https://fc2ppvdb.com/articles/%s", code))
	if err != nil {
		utils.Log.Infof("failed to query fc2 film info for:[%s], error message:%s", code, err.Error())
	}

	for actor, _ := range actorMap {
		actors = append(actors, actor)
	}

	return imageUrl, actors, releaseTime

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

	unDateFilms, err := db.QueryNoDateFilms("fc2")
	if err != nil {
		utils.Log.Warnf("failed to query no date films: %s", err.Error())
		return
	}

	timeMap := make(map[string]time.Time)

	for _, film := range unDateFilms {

		code := db.GetFilmCode(film.Name)
		if existTime, exist := timeMap[code]; exist {
			film.Date = existTime
		} else {
			_, _, releaseTime := d.getPpvdbFilm(code)
			if releaseTime.Year() != 1 {
				film.Date = releaseTime
			} else {
				film.Date = film.CreatedAt
			}
			timeMap[code] = film.Date
		}

		err1 := db.UpdateFilmDate(film)
		if err1 != nil {
			utils.Log.Warnf("failed to update film info: %s", err1.Error())
		}

		// avoid 429
		time.Sleep(time.Duration(d.ScanTimeLimit) * time.Second)

	}

	utils.Log.Info("rematching completed")

}
