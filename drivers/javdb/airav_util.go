package javdb

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/open_ai"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"strconv"
	"strings"
	"time"
)

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
