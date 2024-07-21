package javdb

import (
	"cmp"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (d *Javdb) getFilms(dirName string, urlFunc func(index int) string) ([]model.ObjThumb, error) {

	// 1. 获取所有影片
	javFilms, err := virtual_file.GetFilmsWitchStorage("javdb", dirName, urlFunc,
		func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
			return d.getJavPageInfo(urlFunc, index, data)
		})

	if err != nil && len(javFilms) == 0 {
		utils.Log.Info("javdb影片获取失败", err)
		return javFilms, err
	}

	// 2. 根据影片名字映射名称
	// 2.1 获取所有映射名称
	namingFilms, err := d.getAiravNamingFilms(javFilms, dirName)
	if err != nil || len(namingFilms) == 0 {
		utils.Log.Info("中文影片名称获取失败", err)
		return javFilms, nil
	}

	// 2.2 进行映射
	for index, film := range javFilms {
		code := splitCode(film.Name)
		if newName, exist := namingFilms[code]; exist && strings.HasSuffix(javFilms[index].Name, "mp4") {
			_, newName = splitName(newName)
			newName = virtual_file.AppendFilmName(virtual_file.CutString(newName))
			javFilms[index].Name = fmt.Sprintf("%s %s", code, strings.ReplaceAll(newName, "-", ""))

			virtual_file.CacheImage("javdb", dirName, virtual_file.AppendImageName(javFilms[index].Name), javFilms[index].Thumb())

		}
	}

	return javFilms, err

}

func (d *Javdb) getStars() []model.ObjThumb {
	return virtual_file.GeoStorageFilms("javdb", "个人收藏")
}

func (d *Javdb) addStar(code string) (model.ObjThumb, error) {

	javFilms, _, err := d.getJavPageInfo(func(index int) string {
		return fmt.Sprintf("https://javdb.com/search?f=download&q=%s", code)
	}, 1, []model.ObjThumb{})
	if err != nil {
		utils.Log.Info("jav影片查询失败:", err)
		return model.ObjThumb{}, err
	}

	if len(javFilms) == 0 || strings.ToLower(code) != strings.ToLower(splitCode(javFilms[0].Name)) {
		return model.ObjThumb{}, errors.New(fmt.Sprintf("影片:%s未查询到", code))
	}

	cachingFilm := javFilms[0]
	_, airavFilm := d.getAiravNamingAddr(cachingFilm)
	if airavFilm.Name != "" {
		cachingFilm.Name = airavFilm.Name
	} else {
		_, njavFilm := d.getNjavAddr(cachingFilm)
		if njavFilm.Name != "" {
			cachingFilm.Name = njavFilm.Name
		}
	}

	err = db.CreateFilms("javdb", "个人收藏", []model.ObjThumb{cachingFilm})
	cachingFilm.Name = virtual_file.AppendFilmName(cachingFilm.Name)
	cachingFilm.Path = "个人收藏"

	return cachingFilm, err

}

func (d *Javdb) getMagnet(file model.Obj) (string, error) {

	magnetCache := db.QueryCacheFileId(file.GetName())
	if magnetCache.Magnet != "" {
		utils.Log.Infof("返回缓存中的磁力地址:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	var magnets []Magnet

	collector.OnHTML(".magnet-links", func(element *colly.HTMLElement) {
		element.ForEach(".item", func(i int, magnetEle *colly.HTMLElement) {

			var tags []string

			magnetEle.ForEach(".tag", func(i int, tag *colly.HTMLElement) {
				tags = append(tags, tag.Text)
			})

			fileSizeText := magnetEle.ChildText(".meta")
			fileSize := strings.Split(fileSizeText, ",")[0]
			bytes, err := humanize.ParseBytes(fileSize)
			if err != nil {
				utils.Log.Infof("格式化文件大小失败:%s,错误原因:%v", fileSizeText, err)
			}

			magnets = append(magnets, Magnet{
				MagnetUrl: magnetEle.ChildAttr("a", "href"),
				Tag:       tags,
				FileSize:  bytes,
			})

		})
	})
	err := collector.Visit(d.SpiderServer + file.GetID())

	if len(magnets) == 0 {
		return "", err
	}

	slices.SortFunc(magnets, func(a, b Magnet) int {
		tagCmp := cmp.Compare(len(b.Tag), len(a.Tag))
		if tagCmp != 0 {
			return tagCmp
		}

		return cmp.Compare(a.FileSize, b.FileSize)

	})

	maxLen := len(magnets[0].Tag)
	magnetGroup := utils.GroupByProperty(magnets, func(t Magnet) int {
		return len(t.Tag)
	})
	mostTagMagnets := magnetGroup[maxLen]
	magnet := mostTagMagnets[len(mostTagMagnets)/2]

	err = db.CreateCacheFile(magnet.MagnetUrl, "", file.GetName())

	return magnet.MagnetUrl, err

}

func (d *Javdb) getJavPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	var nextPage bool

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

			href := element.ChildAttr("a", "href")
			title := element.ChildText(".video-title")
			image := element.ChildAttr("img", "src")

			parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			if parse.After(time.Now()) {
				parse = time.Now()
			}

			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     title,
					IsFolder: false,
					ID:       "https://javdb.com/" + href,
					Size:     622857143,
					Modified: parse,
				},
				Thumbnail: model.Thumbnail{Thumbnail: image},
			})

		})
	})

	collector.OnHTML(".pagination-next", func(element *colly.HTMLElement) {
		nextPage = len(element.Attr("href")) != 0
	})

	url := d.SpiderServer + urlFunc(index)
	err := collector.Visit(url)
	utils.Log.Infof("开始爬取javdb页面：%s，错误：%v", url, err)

	return data, nextPage, err

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
	utils.Log.Infof("开始爬取njav页面：%s", url)
	err := collector.Visit(url)

	return data, preLen != len(data), err

}

func (d *Javdb) getAiravPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	nextPage := false

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})
	extensions.RandomUserAgent(collector)

	collector.OnHTML(".row.row-cols-2.row-cols-lg-4.g-2.mt-0", func(element *colly.HTMLElement) {
		element.ForEach(".col.oneVideo", func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr(".oneVideo-top a", "href")
			title := element.ChildText(".oneVideo-body h5")

			parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     title,
					IsFolder: false,
					ID:       "https://airav.io" + href,
					Size:     622857143,
					Modified: parse,
				},
			})
		})

	})

	collector.OnHTML(".col-2.d-flex.align-items-center.px-4.page-input", func(element *colly.HTMLElement) {
		page := element.ChildAttr("input", "max")
		pageNum, _ := strconv.Atoi(page)
		if page != "" && index < pageNum {
			nextPage = true
		}
	})

	url := urlFunc(index)

	utils.Log.Infof("开始爬取airav页面：%s", url)
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

func (d *Javdb) getAiravNamingAddr(film model.ObjThumb) (string, model.ObjThumb) {

	actorUrl := ""
	actorPageUrl := ""

	code := splitCode(film.Name)

	searchResult, _, err := d.getAiravPageInfo(func(index int) string {
		return fmt.Sprintf("https://airav.io/cn/search_result?kw=%s", code)
	}, 1, []model.ObjThumb{})
	if err != nil {
		utils.Log.Info("airav页面爬取错误", err)
		return actorPageUrl, model.ObjThumb{}
	}
	if len(searchResult) > 0 && splitCode(searchResult[0].Name) == code {
		actorUrl = searchResult[0].ID
	}
	if actorUrl == "" {
		return actorPageUrl, model.ObjThumb{}
	}

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(".list-group", func(element *colly.HTMLElement) {

		url := element.ChildAttr(".my-2 a", "href")
		if url != "" && strings.Contains(url, "/cn/actor") {
			actorPageUrl = fmt.Sprintf("https://airav.io%s&idx=", url)
		}

	})

	err = collector.Visit(actorUrl)
	if err != nil {
		utils.Log.Info("演员主页爬取失败", err)
	}

	return actorPageUrl, searchResult[0]

}

func (d *Javdb) getAiravNamingFilms(films []model.ObjThumb, dirName string) (map[string]string, error) {

	init := false
	nameCache := make(map[string]string)

	// 1. 获取库中已爬取结果
	actors := db.QueryByActor("airav", dirName)
	for index := range actors {
		name := actors[index].Name
		nameCache[splitCode(name)] = virtual_file.AppendFilmName(name)
	}
	if len(nameCache) != 0 {
		init = true
	}

	// 2. 爬取新的数据
	for index := range films {

		code := splitCode(films[index].Name)

		// 2.1 仅当未爬取到才爬取，对于非第一条数据，若未爬取到则不再爬取
		if nameCache[code] == "" && (init == false || index == 0) {
			// 2.2 首先爬取airav站点到
			addr, searchResult := d.getAiravNamingAddr(films[index])

			if searchResult.ID != "" {
				// 2.2.1 有该作品信息
				nameCache[splitCode(searchResult.Name)] = virtual_file.AppendFilmName(searchResult.Name)
			}

			if addr != "" {
				// 2.2.2 爬取该主演所有作品
				namingFilms, err := virtual_file.GetFilmsWitchStorage("airav", dirName, func(index int) string {
					return addr + strconv.Itoa(index)
				},
					func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
						return d.getAiravPageInfo(urlFunc, index, data)
					})

				if err != nil {
					utils.Log.Info("airav影片列表爬取失败", err)
				}
				for nameFileIndex := range namingFilms {
					nameCache[splitCode(namingFilms[nameFileIndex].Name)] = namingFilms[nameFileIndex].Name
				}

			} else if addr == "" && searchResult.ID == "" {

				// 2.2.3 airav没有该作品信息，尝试爬取其它站点的
				njavSearchResult, _, err := d.getNajavPageInfo(func(index int) string {
					return fmt.Sprintf("https://njav.tv/zh/search?keyword=%s", code)
				}, 1, []model.ObjThumb{})

				if err != nil {
					utils.Log.Info("njav搜索影片失败", err)
				}
				if len(njavSearchResult) > 0 && splitCode(njavSearchResult[0].Name) == code {
					nameCache[code] = virtual_file.AppendFilmName(njavSearchResult[0].Name)
				} else {
					nameCache[code] = films[index].Name
				}

			}

		}

	}

	utils.Log.Info("影片名称映射结束")

	return nameCache, nil

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
