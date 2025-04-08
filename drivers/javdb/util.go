package javdb

import (
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/av"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/open_ai"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (d *Javdb) getFilms(dirName string, urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	// 1. 获取所有影片
	javFilms, err := virtual_file.GetFilmsWithStorage("javdb", dirName, dirName, urlFunc,
		func(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {
			return d.getJavPageInfo(urlFunc, index, data)
		}, virtual_file.Option{CacheFile: false, MaxPageNum: 20})

	if err != nil && len(javFilms) == 0 {
		utils.Log.Info("javdb影片获取失败", err)
		return javFilms, err
	}

	// 2. 根据影片名字映射名称
	return d.mappingNames(dirName, javFilms)

}

func (d *Javdb) mappingNames(dirName string, javFilms []model.EmbyFileObj) ([]model.EmbyFileObj, error) {

	if len(javFilms) == 0 {
		return javFilms, nil
	}

	var noTitleFilms []model.EmbyFileObj
	for _, film := range javFilms {
		if film.Title == "" {
			noTitleFilms = append(noTitleFilms, film)
		}
	}
	if len(noTitleFilms) == 0 {
		return javFilms, nil
	}

	// 2.1 获取所有映射名称
	namingFilms, err := d.getAiravNamingFilms(noTitleFilms, dirName)
	if err != nil || len(namingFilms) == 0 {
		utils.Log.Info("中文影片名称获取失败", err)
		return javFilms, nil
	}

	// 2.2 进行映射
	var savingFilms []model.EmbyFileObj
	var deletingFilms []string

	for index, film := range javFilms {
		if film.Title == "" {
			code := splitCode(film.Name)
			if newName, exist := namingFilms[code]; exist {
				javFilms[index].Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(newName)))
				javFilms[index].Title = virtual_file.ClearFilmName(newName)

				savingFilms = append(savingFilms, javFilms[index])
				deletingFilms = append(deletingFilms, film.ID)
			}
		}
	}
	if len(savingFilms) > 0 {
		err1 := db.DeleteFilmsByUrl("javdb", dirName, deletingFilms)
		if err1 != nil {
			utils.Log.Warnf("failed to delete films:[%s], error message: %s", deletingFilms, err1.Error())
		} else {
			err2 := db.CreateFilms("javdb", dirName, dirName, savingFilms)
			if err2 != nil {
				utils.Log.Infof("failed to save films:[%s], error message: %s", deletingFilms, err2.Error())
			}
		}
	}

	for _, film := range javFilms {
		created := virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
			Source:   "javdb",
			Dir:      dirName,
			FileName: virtual_file.AppendImageName(film.Name),
			Title:    film.Title,
			ImgUrl:   film.Thumb(),
			Actors:   []string{dirName},
			Release:  film.ReleaseTime,
		})

		if created == virtual_file.Exist && d.QuickCache {
			// 已经创建过了，后续不再创建
			break
		}

	}

	utils.Log.Infof("name mapping finished for:[%s]", dirName)

	return javFilms, err
}

func (d *Javdb) getStars() []model.EmbyFileObj {
	films := virtual_file.GetStorageFilms("javdb", "个人收藏", true)

	if d.RefreshNfo {
		var filmNames []string
		for _, film := range films {
			filmNames = append(filmNames, film.Name)
		}
		virtual_file.ClearUnUsedFiles("javdb", "个人收藏", filmNames)
	}

	return films
}

func (d *Javdb) addStar(code string) (model.EmbyFileObj, error) {

	javFilms, _, err := d.getJavPageInfo(func(index int) string {
		return fmt.Sprintf("https://javdb.com/search?f=download&q=%s", code)
	}, 1, []model.EmbyFileObj{})
	if err != nil {
		utils.Log.Info("jav影片查询失败:", err)
		return model.EmbyFileObj{}, err
	}

	if len(javFilms) == 0 || strings.ToLower(code) != strings.ToLower(splitCode(javFilms[0].Name)) {
		return model.EmbyFileObj{}, errors.New(fmt.Sprintf("影片:%s未查询到", code))
	}

	cachingFilm := javFilms[0]
	_, airavFilm := d.getAiravNamingAddr(cachingFilm)
	if airavFilm.Name != "" {
		cachingFilm.Title = airavFilm.Title
		cachingFilm.Name = airavFilm.Name
	} else {
		tempCode, name := splitName(cachingFilm.Name)

		translatedText := open_ai.Translate(virtual_file.ClearFilmName(name))
		if translatedText != "" {
			translatedText = fmt.Sprintf("%s %s", tempCode, translatedText)
			cachingFilm.Name = translatedText
			cachingFilm.Title = airavFilm.Title
		}
	}

	err = db.CreateFilms("javdb", "个人收藏", "个人收藏", []model.EmbyFileObj{cachingFilm})
	cachingFilm.Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(cachingFilm.Name)))
	cachingFilm.Path = "个人收藏"

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	actorMapping := make(map[string]string)
	for _, actor := range actors {
		actorMapping[actor.Url] = actor.Name
	}

	actorNames := d.getJavActorNames(cachingFilm.ID, actorMapping)
	if len(actorNames) == 0 {
		actorNames = append(actorNames, "个人收藏")
	}

	_ = virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
		Source:   "javdb",
		Dir:      "个人收藏",
		FileName: virtual_file.AppendImageName(cachingFilm.Name),
		Title:    cachingFilm.Name,
		ImgUrl:   cachingFilm.Thumb(),
		Actors:   actorNames,
		Release:  cachingFilm.ReleaseTime,
	})

	return cachingFilm, err

}

func (d *Javdb) getMagnet(file model.Obj) (string, error) {

	magnetCache := db.QueryMagnetCacheByName("javdb", file.GetName())
	if magnetCache.Magnet != "" {
		utils.Log.Infof("return the magnet link from the cache:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	javdbMeta, err := av.GetMetaFromJavdb(file.GetID())
	if err != nil || len(javdbMeta.Magnets) == 0 {
		utils.Log.Warnf("failed to get javdb magnet info: %v,error message: %v, using the suke magnet instead.", javdbMeta, err)
		sukeMeta, err2 := av.GetMetaFromSuke(db.GetFilmCode(file.GetName()))
		if err2 != nil {
			utils.Log.Warn("failed to get suke magnet info:", err2.Error())
			return "", err2
		} else {
			if len(sukeMeta.Magnets) > 0 {
				return sukeMeta.Magnets[0].GetMagnet(), nil
			}
		}
		return "", err
	}

	actorMapping := make(map[string]string)
	for _, actor := range javdbMeta.Actors {
		actorMapping[actor.Id] = actor.Name
	}

	if embyObj, ok := file.(*model.EmbyFileObj); ok && len(embyObj.Actors) == 0 {

		actors := db.QueryActor(strconv.Itoa(int(d.ID)))
		for _, actor := range actors {
			if actorMapping[actor.Url] != "" {
				actorMapping[actor.Url] = actor.Name
			}
		}

		var actorNames []string
		for _, name := range actorMapping {
			actorNames = append(actorNames, name)
		}

		virtual_file.UpdateNfo(virtual_file.MediaInfo{
			Source:   "javdb",
			Dir:      embyObj.Path,
			FileName: virtual_file.AppendImageName(embyObj.Name),
			Title:    embyObj.Title,
			Actors:   actorNames,
			Release:  embyObj.ReleaseTime,
		})

	}

	magnet := ""
	subtitle := false
	if javdbMeta.Magnets[0].IsSubTitle() {
		magnet = javdbMeta.Magnets[0].GetMagnet()
		subtitle = true
	} else {
		sukeMeta, err2 := av.GetMetaFromSuke(db.GetFilmCode(file.GetName()))
		if err2 != nil {
			utils.Log.Warn("failed to get suke magnet info:", err2.Error())
		} else {
			if len(sukeMeta.Magnets) > 0 {
				magnet = sukeMeta.Magnets[0].GetMagnet()
				subtitle = true
			}
		}

	}

	if magnet == "" {
		magnet = javdbMeta.Magnets[0].GetMagnet()
		subtitle = javdbMeta.Magnets[0].IsSubTitle()
	}

	err = db.CreateMagnetCache(model.MagnetCache{
		DriverType: "javdb",
		Magnet:     magnet,
		Name:       file.GetName(),
		Subtitle:   subtitle,
	})
	return magnet, err

}

func (d *Javdb) getJavPageInfo(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {

	var nextPage bool

	filter := strings.Split(d.Addition.Filter, ",")

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
		_ = c.SetCookies("https://javdb.com", setCookieRaw(d.Cookie))
	})
	extensions.RandomUserAgent(collector)

	collector.OnHTML(".movie-list", func(element *colly.HTMLElement) {
		element.ForEach(".item", func(i int, element *colly.HTMLElement) {

			tag := element.ChildText(".tag")
			if tag == "" {
				return
			}

			title := element.ChildText(".video-title")

			for _, filterItem := range filter {
				if filterItem != "" && strings.Contains(title, filterItem) {
					return
				}
			}

			href := element.ChildAttr("a", "href")
			image := element.ChildAttr("img", "src")

			releaseTime, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			if releaseTime.Year() == 1 {
				releaseTime = time.Now()
			}

			data = append(data, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name:     title,
						IsFolder: false,
						ID:       "https://javdb.com/" + href,
						Size:     622857143,
						Modified: time.Now(),
					},
					Thumbnail: model.Thumbnail{Thumbnail: image},
				},
				ReleaseTime: releaseTime,
			})

		})
	})

	collector.OnHTML(".pagination-next", func(element *colly.HTMLElement) {
		nextPage = len(element.Attr("href")) != 0
	})

	url := d.SpiderServer + urlFunc(index)
	err := collector.Visit(url)
	utils.Log.Debugf("开始爬取javdb页面：%s，错误：%v", url, err)

	return data, nextPage, err

}

func (d *Javdb) getJavActorNames(filmUrl string, mapping map[string]string) []string {

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
		_ = c.SetCookies("https://javdb.com", setCookieRaw(d.Cookie))
	})
	extensions.RandomUserAgent(collector)

	actorMapping := make(map[string]string)

	collector.OnHTML(".panel.movie-panel-info", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {

			href := element.Attr("href")
			if strings.Contains(href, "/actors/") {
				actorUrl := strings.ReplaceAll(href, "/actors/", "")
				actorMapping[actorUrl] = element.Text
			}

		})
	})

	err := collector.Visit(filmUrl)

	if err != nil {
		utils.Log.Warnf("演员信息获取失败:%s", err.Error())
		return []string{}
	}

	var actors []string
	for url, name := range actorMapping {
		if mapping[url] != "" {
			actors = append(actors, mapping[url])
		} else {
			actors = append(actors, name)
		}
	}

	return actors

}

func (d *Javdb) getNajavPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	preLen := len(data)

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})
	extensions.RandomUserAgent(collector)

	collector.OnHTML(".row.box-item-list.gutter-20", func(element *colly.HTMLElement) {
		element.ForEach(".box-item", func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr(".thumb a", "href")
			title := element.ChildText(".detail a")

			parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     title,
					IsFolder: false,
					ID:       "https://njav.tv/zh/" + href,
					Size:     622857143,
					Modified: parse,
				},
			})
		})

	})

	url := urlFunc(index)
	utils.Log.Debugf("开始爬取njav页面：%s", url)
	err := collector.Visit(url)

	return data, preLen != len(data), err

}

func (d *Javdb) getAiravPageInfo(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {

	nextPage := false

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 20)
	})
	extensions.RandomUserAgent(collector)

	collector.OnHTML(".row.row-cols-2.row-cols-lg-4.g-2.mt-0", func(element *colly.HTMLElement) {
		element.ForEach(".col.oneVideo", func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr(".oneVideo-top a", "href")
			title := element.ChildText(".oneVideo-body h5")

			if !strings.Contains(title, "马赛克破坏版") {
				parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
				data = append(data, model.EmbyFileObj{
					ObjThumb: model.ObjThumb{
						Object: model.Object{
							Name:     title,
							IsFolder: false,
							ID:       "https://airav.io" + href,
							Size:     622857143,
							Modified: parse,
						},
					},
					Title: title,
				})
			}

		})

	})

	collector.OnHTML(".col-2.d-flex.align-items-center.px-4.page-input", func(element *colly.HTMLElement) {
		page := element.ChildAttr(".form-control", "max")
		pageNum, _ := strconv.Atoi(page)
		if page != "" && index < pageNum {
			nextPage = true
		}
	})

	url := urlFunc(index)

	utils.Log.Debugf("开始爬取airav页面：%s", url)
	err := collector.Visit(url)

	return data, nextPage, err

}

// set cookies raw
func setCookieRaw(cookieRaw string) []*http.Cookie {
	// 可以添加多个cookie
	var cookies []*http.Cookie
	cookieList := strings.Split(cookieRaw, "; ")
	for _, item := range cookieList {
		keyValue := strings.Split(item, "=")
		// fmt.Println(keyValue)
		name := keyValue[0]
		valueList := keyValue[1:]
		cookieItem := http.Cookie{
			Name:  name,
			Value: strings.Join(valueList, "="),
		}
		cookies = append(cookies, &cookieItem)
	}
	return cookies
}

func (d *Javdb) getNjavAddr(films model.ObjThumb) (string, model.ObjThumb) {

	actorUrl := ""
	actorPageUrl := ""

	code := splitCode(films.Name)

	searchResult, _, err := d.getNajavPageInfo(func(index int) string {
		return fmt.Sprintf("https://njav.tv/zh/search?keyword=%s", code)
	}, 1, []model.ObjThumb{})

	if err != nil {
		utils.Log.Info("njav页面爬取错误", err)
		return "", model.ObjThumb{}
	}

	if len(searchResult) > 0 && splitCode(searchResult[0].Name) == code {
		actorUrl = searchResult[0].ID
	}

	if actorUrl != "" {
		collector := colly.NewCollector(func(c *colly.Collector) {
			c.SetRequestTimeout(time.Second * 10)
		})

		collector.OnHTML("#details", func(element *colly.HTMLElement) {

			url := element.ChildAttr(".content a", "href")
			if url != "" {
				actorPageUrl = fmt.Sprintf("https://njav.tv/zh/%s?page=", url)
			}

		})

		err = collector.Visit(actorUrl)
		if err != nil {
			utils.Log.Info("演员主页爬取失败", err)
		}

		return actorPageUrl, searchResult[0]
	}

	return "", model.ObjThumb{}

}

func (d *Javdb) getAiravNamingAddr(film model.EmbyFileObj) (string, model.EmbyFileObj) {

	actorUrl := ""
	actorPageUrl := ""
	var matchedFilm model.EmbyFileObj

	code := splitCode(film.Name)

	searchResult, _, err := d.getAiravPageInfo(func(index int) string {
		return fmt.Sprintf("https://airav.io/cn/search_result?kw=%s", code)
	}, 1, []model.EmbyFileObj{})
	if err != nil {
		utils.Log.Info("airav页面爬取错误", err)
		return actorPageUrl, model.EmbyFileObj{}
	}

	for _, item := range searchResult {
		if splitCode(item.Name) == code {
			actorUrl = item.ID
			matchedFilm = item
			if actorUrl == "" {
				return actorPageUrl, item
			}
		}
	}

	if actorUrl == "" {
		return actorPageUrl, model.EmbyFileObj{}
	}

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(".list-group", func(element *colly.HTMLElement) {

		urls := element.ChildAttrs(".my-2 a", "href")

		var actors []string
		for _, url := range urls {
			if strings.Contains(url, "/cn/actor") {
				actors = append(actors, url)
			}
		}

		// 仅当演员只有一个的时候才进行爬取
		if len(actors) == 1 {
			actorPageUrl = fmt.Sprintf("https://airav.io%s&idx=", actors[0])
		}

	})

	err = collector.Visit(actorUrl)
	if err != nil {
		utils.Log.Info("演员主页爬取失败", err)
	}

	return actorPageUrl, matchedFilm

}

func (d *Javdb) getAiravNamingFilms(films []model.EmbyFileObj, dirName string) (map[string]string, error) {

	nameCache := make(map[string]string)
	actorCache := make(map[string]bool)

	var savingNamingMapping []model.EmbyFileObj

	// 1. 获取库中已爬取结果
	actors := db.QueryByActor("airav", dirName)
	for index := range actors {
		name := actors[index].Name
		nameCache[splitCode(name)] = virtual_file.AppendFilmName(name)
	}

	// 2. 爬取新的数据
	for index := range films {

		code, name := splitName(films[index].Name)

		// 2.1 仅当未爬取到才爬取
		if nameCache[code] == "" {
			// 2.2 首先爬取airav站点的
			addr, searchResult := d.getAiravNamingAddr(films[index])

			if addr != "" && !actorCache[addr] {
				// 2.2.1 爬取该主演所有作品
				namingFilms, err := virtual_file.GetFilmsWithStorage("airav", dirName, addr, func(index int) string {
					return addr + strconv.Itoa(index)
				},
					func(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {
						return d.getAiravPageInfo(urlFunc, index, data)
					}, virtual_file.Option{CacheFile: false, MaxPageNum: 40})

				if err != nil {
					utils.Log.Info("airav影片列表爬取失败", err)
				}
				for nameFileIndex := range namingFilms {
					tempName := namingFilms[nameFileIndex].Name
					tempCode := splitCode(tempName)
					if nameCache[tempCode] == "" {
						nameCache[tempCode] = tempName
						savingNamingMapping = append(savingNamingMapping, namingFilms[nameFileIndex])
					}
				}

				actorCache[addr] = true

			}

			if nameCache[code] == "" && searchResult.ID != "" {
				// 2.2.2 有该作品信息
				nameCache[splitCode(searchResult.Name)] = virtual_file.AppendFilmName(searchResult.Name)
				if addr == "" || actorCache[addr] {
					// 没有爬取到演员主页，直接记录该影片信息
					savingNamingMapping = append(savingNamingMapping, searchResult)
				}
			}

			if nameCache[code] == "" {

				// 2.2.3 AI翻译
				translatedText := open_ai.Translate(virtual_file.ClearFilmName(name))
				if translatedText != "" {
					translatedText = fmt.Sprintf("%s %s", code, translatedText)
					nameCache[code] = virtual_file.AppendFilmName(translatedText)
					savingNamingMapping = append(savingNamingMapping, model.EmbyFileObj{
						ObjThumb: model.ObjThumb{
							Object: model.Object{Name: translatedText},
						},
						Title: translatedText})
				} else {
					nameCache[code] = films[index].Name
					savingNamingMapping = append(savingNamingMapping, model.EmbyFileObj{
						ObjThumb: model.ObjThumb{
							Object: model.Object{Name: films[index].Name},
						},
						Title: films[index].Name})
				}

			}

		}

	}

	if len(savingNamingMapping) > 0 {
		err := db.CreateFilms("airav", dirName, dirName, savingNamingMapping)
		if err != nil {
			utils.Log.Infof("影片名称映射入库失败:%s", err.Error())
		}
	}

	return nameCache, nil

}

func (d *Javdb) reMatchSubtitles() {

	utils.Log.Info("start rematching subtitles for films without subtitles")

	caches, err := db.QueryNoSubtitlesCache("javdb")
	if err != nil {
		utils.Log.Warnf("failed to query the films without subtitles")
		return
	}
	if len(caches) != 0 {
		var savingCaches []model.MagnetCache
		var unFindCaches []model.MagnetCache

		for _, cache := range caches {

			film, err1 := db.QueryFilmByCode("javdb", cache.Code)
			if err1 != nil {
				utils.Log.Warn("failed to query film:", err1.Error())
			} else {
				if film.Url != "" {
					javdbMeta, err2 := av.GetMetaFromJavdb(film.Url)
					if err2 != nil {
						utils.Log.Warn("failed to get javdb magnet info:", err2.Error())
					} else if len(javdbMeta.Magnets) > 0 && javdbMeta.Magnets[0].IsSubTitle() {
						cache.Subtitle = true
						cache.Magnet = javdbMeta.Magnets[0].GetMagnet()
					}
				}
			}

			if !cache.Subtitle {
				sukeMeta, err2 := av.GetMetaFromSuke(cache.Code)
				if err2 != nil {
					utils.Log.Warn("failed to get suke magnet info:", err2.Error())
				} else {
					if len(sukeMeta.Magnets) > 0 && sukeMeta.Magnets[0].IsSubTitle() {
						cache.Subtitle = true
						cache.Magnet = sukeMeta.Magnets[0].GetMagnet()
					}
				}
			}

			if cache.Subtitle {
				savingCaches = append(savingCaches, cache)
			} else {
				unFindCaches = append(unFindCaches, cache)
			}

		}

		if len(savingCaches) > 0 {
			err2 := db.BatchCreateMagnetCache(savingCaches)
			if err2 != nil {
				utils.Log.Warn("failed to create magnet cache:", err2.Error())
			}
			utils.Log.Infof("update films magnet cache:[%v]", savingCaches)
		}

		if len(unFindCaches) > 0 {
			var names []string
			for _, cache := range unFindCaches {
				names = append(names, cache.Name)
			}
			err2 := db.UpdateScanData("javdb", names, time.Now())
			if err2 != nil {
				utils.Log.Warn("failed to update scan data:", err2.Error())
			}
			utils.Log.Infof("films:[%v] still have not matched with subtitles, update the scan info", names)
		}
	}

	noMatchCaches, err2 := db.QueryNoMatchCache("javdb")
	if err2 != nil {
		utils.Log.Warn("failed to query film:", err2.Error())
		return
	}

	if len(noMatchCaches) > 0 {
		deletingCache := make(map[string][]string)
		for _, cache := range noMatchCaches {
			deletingCache[cache.DriverType] = append(deletingCache[cache.DriverType], cache.Name)
		}

		for driverType, names := range deletingCache {
			err3 := db.DeleteCacheByName(driverType, names)
			if err3 != nil {
				utils.Log.Warn("failed to delete cache:", err3.Error())
			}
		}
		utils.Log.Infof("Delete the cached films that do not match the subtitles:[%v]", noMatchCaches)
	}

	utils.Log.Info("rematching completed")

}

func (d *Javdb) refreshNfo() {

	utils.Log.Info("start refresh nfo for javdb")

	var actorNames []string
	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		actorNames = append(actorNames, actor.Name)
	}
	for _, actor := range actorNames {

		films := virtual_file.GetStorageFilms("javdb", actor, false)

		// refresh nfo
		mappingNameFilms, err := d.mappingNames(actor, films)
		if err != nil {
			utils.Log.Warn("failed to get mapping names:", err.Error())
			continue
		}

		var filmNames []string
		for _, film := range mappingNameFilms {
			virtual_file.UpdateNfo(virtual_file.MediaInfo{
				Source:   "javdb",
				Dir:      film.Path,
				FileName: virtual_file.AppendImageName(film.Name),
				Release:  film.ReleaseTime,
				Title:    film.Title,
			})
			filmNames = append(filmNames, film.Name)
		}

		// clear unused files
		virtual_file.ClearUnUsedFiles("javdb", actor, filmNames)

	}

	utils.Log.Info("finish refresh nfo")

}

func (d *Javdb) filterFilms() {

	utils.Log.Info("start to filter javdb films")

	films, err := db.QueryFilmsByNamePrefix("javdb", strings.Split(d.Filter, ","))
	if err != nil {
		utils.Log.Warn("failed to query films:", err.Error())
		return
	}

	if len(films) > 0 {
		utils.Log.Infof("deleting films:[%v]", films)
		for _, film := range films {
			err1 := d.deleteFilm(film.Actor, virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(film.Name))), film.Url)
			if err1 != nil {
				utils.Log.Warn("failed to delete film:", err1.Error())
			}
		}
	}

	utils.Log.Info("finish filter javdb films")

}

func (d *Javdb) deleteFilm(dir, fileName, id string) error {
	err := db.DeleteAllMagnetCacheByCode(fileName)
	if err != nil {
		utils.Log.Warnf("failed to delete film cache:[%s], error message:[%s]", fileName, err.Error())
	}

	err = db.DeleteFilmsByUrl("javdb", dir, []string{id})
	if err != nil {
		utils.Log.Infof("failed to delete film:[%s], error message:[%s]", fileName, err.Error())
		return err
	}

	err = virtual_file.DeleteImageAndNfo("javdb", dir, fileName)
	if err != nil {
		utils.Log.Infof("failed to delete film nfo:[%s], error message:[%s]", fileName, err)
		return err
	}
	return nil
}

func splitName(sourceName string) (string, string) {

	index := strings.Index(sourceName, " ")
	if index <= 0 || index == len(sourceName)-1 {
		return sourceName, sourceName
	}

	return sourceName[:index], sourceName[index+1:]

}

func splitCode(sourceName string) string {

	code, _ := splitName(sourceName)
	return code

}
