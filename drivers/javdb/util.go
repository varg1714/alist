package javdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/av"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/offline_download/tool"
	"github.com/OpenListTeam/OpenList/v4/internal/open_ai"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (d *Javdb) getFilms(dirName string, urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	// 1. 获取所有影片
	javFilms, err := virtual_file.GetFilmsWithStorage(DriverName, dirName, dirName, urlFunc,
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
		if !film.Translated {
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
		if !film.Translated {
			code := splitCode(film.Name)
			if newName, exist := namingFilms[code]; exist {
				javFilms[index].Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(newName)))
				javFilms[index].Title = virtual_file.ClearFilmName(newName)

				savingFilms = append(savingFilms, javFilms[index])
				deletingFilms = append(deletingFilms, film.Url)
			}
		}
	}
	if len(savingFilms) > 0 {
		err1 := db.DeleteFilmsByUrl(DriverName, dirName, deletingFilms)
		if err1 != nil {
			utils.Log.Warnf("failed to delete films:[%s], error message: %s", deletingFilms, err1.Error())
		} else {
			err2 := db.CreateFilms(DriverName, dirName, dirName, savingFilms)
			if err2 != nil {
				utils.Log.Infof("failed to save films:[%s], error message: %s", deletingFilms, err2.Error())
			}
		}
	}

	for _, film := range javFilms {
		created := virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
			Source:   DriverName,
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
	films := virtual_file.GetStorageFilms(DriverName, "个人收藏", true)

	if d.RefreshNfo {
		var filmNames []string
		for _, film := range films {
			filmNames = append(filmNames, film.Name)
		}
		virtual_file.ClearUnUsedFiles(DriverName, "个人收藏", filmNames)
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
			cachingFilm.Title = translatedText
		}
	}

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	actorMapping := make(map[string]string)
	for _, actor := range actors {
		actorMapping[actor.Url] = actor.Name
	}

	actorNames := d.getJavActorNames(cachingFilm.Url, actorMapping)
	if len(actorNames) == 0 {
		actorNames = append(actorNames, "个人收藏")
	}
	cachingFilm.Actors = actorNames

	err = db.CreateFilms(DriverName, "个人收藏", "个人收藏", []model.EmbyFileObj{cachingFilm})
	cachingFilm.Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(cachingFilm.Name)))
	cachingFilm.Path = "个人收藏"

	_ = virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
		Source:   DriverName,
		Dir:      "个人收藏",
		FileName: virtual_file.AppendImageName(cachingFilm.Name),
		Title:    cachingFilm.Title,
		ImgUrl:   cachingFilm.Thumb(),
		Actors:   actorNames,
		Release:  cachingFilm.ReleaseTime,
	})

	return cachingFilm, err

}

func (d *Javdb) getMagnet(file model.Obj, reMatchFilmMeta bool) (string, error) {

	embyObj, ok := file.(*model.EmbyFileObj)
	if !ok {
		return "", errors.New("this film doesn't contains film info")
	}

	magnetCache := db.QueryMagnetCacheByName(DriverName, embyObj.GetName())
	if magnetCache.Magnet != "" && !reMatchFilmMeta {
		utils.Log.Infof("return the magnet link from the cache:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	javdbMeta, err := av.GetMetaFromJavdb(embyObj.Url)
	if err != nil || len(javdbMeta.Magnets) == 0 {

		if reMatchFilmMeta {
			return "", err
		}

		utils.Log.Warnf("failed to get javdb magnet info: %v,error message: %v, using the suke magnet instead.", javdbMeta, err)
		sukeMeta, err2 := av.GetMetaFromSuke(embyObj.GetName())
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

	d.updateFilmMeta(javdbMeta, embyObj)

	if magnetCache.Magnet == "" {
		return d.cacheMagnet(javdbMeta, embyObj)
	} else {
		return magnetCache.Magnet, nil
	}

}

func (d *Javdb) cacheMagnet(javdbMeta av.Meta, embyObj *model.EmbyFileObj) (string, error) {

	magnet := ""
	subtitle := false
	if javdbMeta.Magnets[0].IsSubTitle() {
		magnet = javdbMeta.Magnets[0].GetMagnet()
		subtitle = true
	} else {
		sukeMeta, err2 := av.GetMetaFromSuke(embyObj.GetName())
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

	err := db.CreateMagnetCache(model.MagnetCache{
		DriverType: DriverName,
		Magnet:     magnet,
		Name:       embyObj.GetName(),
		Subtitle:   subtitle,
		Code:       av.GetFilmCode(embyObj.GetName()),
		ScanAt:     time.Now(),
	})
	return magnet, err
}

func (d *Javdb) updateFilmMeta(javdbMeta av.Meta, embyObj *model.EmbyFileObj) {

	actorMapping := make(map[string]string)
	for _, actor := range javdbMeta.Actors {
		actorMapping[actor.Id] = actor.Name
	}

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

	var tags []string
	if len(javdbMeta.Magnets) > 0 {
		tags = append(tags, javdbMeta.Magnets[0].GetTags()...)
	}

	virtual_file.UpdateNfo(virtual_file.MediaInfo{
		Source:   DriverName,
		Dir:      embyObj.Path,
		FileName: virtual_file.AppendImageName(embyObj.Name),
		Title:    embyObj.Title,
		Actors:   actorNames,
		Release:  embyObj.ReleaseTime,
		Tags:     tags,
	})

	tempId, err1 := strconv.ParseInt(embyObj.ID, 10, 64)
	if err1 == nil {
		err1 = db.UpdateFilm(model.Film{
			ID:     uint(tempId),
			Actors: actorNames,
			Tags:   tags,
		})
		if err1 != nil {
			utils.Log.Warnf("failed to save film: %s, error message: %s", embyObj.GetName(), err1.Error())
		}
	} else {
		utils.Log.Warnf("failed to parse films: %s id to int, error message: %s", embyObj.GetName(), err1.Error())
	}

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
						Size:     622857143,
						Modified: time.Now(),
					},
					Thumbnail: model.Thumbnail{Thumbnail: image},
				},
				ReleaseTime: releaseTime,
				Url:         "https://javdb.com/" + href,
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

func (d *Javdb) getNajavPageInfo(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {

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
			data = append(data, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name:     title,
						IsFolder: false,
						Size:     622857143,
						Modified: parse,
					},
				},
				Url: "https://njav.tv/zh/" + href,
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
							Size:     622857143,
							Modified: parse,
						},
					},
					Title: title,
					Url:   "https://airav.io" + href,
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

func (d *Javdb) getNjavAddr(films model.ObjThumb) (string, model.EmbyFileObj) {

	actorUrl := ""
	actorPageUrl := ""

	code := splitCode(films.Name)

	searchResult, _, err := d.getNajavPageInfo(func(index int) string {
		return fmt.Sprintf("https://njav.tv/zh/search?keyword=%s", code)
	}, 1, []model.EmbyFileObj{})

	if err != nil {
		utils.Log.Info("njav页面爬取错误", err)
		return "", model.EmbyFileObj{}
	}

	if len(searchResult) > 0 && splitCode(searchResult[0].Name) == code {
		actorUrl = searchResult[0].Url
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

	return "", model.EmbyFileObj{}

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
			actorUrl = item.Url
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
		title := actors[index].Title
		nameCache[splitCode(title)] = title
	}

	// 2. 爬取新的数据
	for index := range films {

		code, name := splitName(films[index].Title)

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
					tempName := namingFilms[nameFileIndex].Title
					tempCode := splitCode(tempName)
					if nameCache[tempCode] == "" {
						nameCache[tempCode] = tempName
					}
				}

				actorCache[addr] = true

			}

			if nameCache[code] == "" && searchResult.Url != "" {
				// 2.2.2 有该作品信息
				nameCache[splitCode(searchResult.Title)] = virtual_file.AppendFilmName(searchResult.Title)
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
					nameCache[code] = films[index].Title
					savingNamingMapping = append(savingNamingMapping, model.EmbyFileObj{
						ObjThumb: model.ObjThumb{
							Object: model.Object{Name: films[index].Name},
						},
						Title: films[index].Title})
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

func (d *Javdb) deleteFilm(dir, fileName, id string) error {
	err := db.DeleteAllMagnetCacheByCode(av.GetFilmCode(fileName))
	if err != nil {
		utils.Log.Warnf("failed to delete film cache:[%s], error message:[%s]", fileName, err.Error())
	}

	err = db.DeleteFilmById(id)
	if err != nil {
		utils.Log.Infof("failed to delete film:[%s], error message:[%s]", fileName, err.Error())
		return err
	}

	err = virtual_file.DeleteImageAndNfo(DriverName, dir, fileName)
	if err != nil {
		utils.Log.Infof("failed to delete film nfo:[%s], error message:[%s]", fileName, err)
		return err
	}
	return nil
}

func (d *Javdb) tryAcquireLink(ctx context.Context, file model.Obj, args model.LinkArgs, magnetGetter func(obj model.Obj) (string, error)) (*model.Link, error) {

	link, err := tool.CloudPlay(ctx, args, d.CloudPlayDriverType, d.CloudPlayDownloadPath, file, magnetGetter)

	if err != nil {
		utils.Log.Infof("The first cloud drive download failed:[%s]", err.Error())
		if d.BackPlayDriverType != "" {
			utils.Log.Infof("using the second cloud drive instead.")
			return tool.CloudPlay(ctx, args, d.BackPlayDriverType, d.CloudPlayDownloadPath, file, magnetGetter)
		}
	}

	return link, err
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
