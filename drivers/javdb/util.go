package javdb

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/gocolly/colly/v2"
	"net/http"
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
	if err != nil || len(javFilms) == 0 {
		return javFilms, err
	}

	// 2. 取第一部影片的编码查询获取到所有miss-av的影片
	addr := d.getNamingAddr(javFilms)
	if addr == "" {
		return javFilms, err
	}

	// 3. 根据影片名字映射名称
	// 3.1 获取所有映射名称
	namingFilms, err := virtual_file.GetFilmsWitchStorage("njav.tv", dirName, func(index int) string {
		return addr + strconv.Itoa(index)
	},
		func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
			return d.getNamingPageInfo(urlFunc, index, data)
		})
	if err != nil || len(namingFilms) == 0 {
		utils.Log.Info("中文影片名称获取失败", err)
		return javFilms, nil
	}

	// 3.2 提取名称map
	nameMap := make(map[string]string)
	for _, temp := range namingFilms {
		tempName := temp.GetName()
		index := strings.Index(tempName, " ")
		if index == 0 || index == len(namingFilms)-1 {
			continue
		}
		nameMap[tempName[:index]] = tempName[index+1:]
	}

	// 3.3 进行映射
	for index, film := range javFilms {
		code := strings.Split(film.Name, " ")[0]
		if newName, exist := nameMap[code]; exist {
			javFilms[index].Name = fmt.Sprintf("%s %s", code, strings.ReplaceAll(newName, "-", ""))
		}
	}

	return javFilms, err

}

func (d *Javdb) getMagnet(file model.Obj) (string, error) {

	magnet := ""
	subTitles := false

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(".magnet-links", func(element *colly.HTMLElement) {
		element.ForEach(".item", func(i int, element *colly.HTMLElement) {

			text := element.ChildText(".is-warning")
			if text != "" && (magnet == "" || !subTitles) {
				magnet = element.ChildAttr("a", "href")
				subTitles = true
			}

			if magnet == "" {
				magnet = element.ChildAttr("a", "href")
			}

		})
	})
	err := collector.Visit(file.GetID())

	return magnet, err

}

func (d *Javdb) getJavPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	var nextPage bool

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
		_ = c.SetCookies("https://javdb.com", setCookieRaw(d.Cookie))
	})

	collector.OnHTML(".movie-list", func(element *colly.HTMLElement) {
		element.ForEach(".item", func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr("a", "href")
			title := element.ChildText(".video-title")
			image := element.ChildAttr("img", "src")

			parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     title,
					IsFolder: true,
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

	err := collector.Visit(urlFunc(index))

	return data, nextPage, err

}

func (d *Javdb) getNamingPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

	preLen := len(data) - 1

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(".row.box-item-list.gutter-20", func(element *colly.HTMLElement) {
		element.ForEach(".box-item", func(i int, element *colly.HTMLElement) {

			href := element.ChildAttr(".thumb a", "href")
			title := element.ChildText(".detail a")

			parse, _ := time.Parse(time.DateOnly, element.ChildText(".meta"))
			data = append(data, model.ObjThumb{
				Object: model.Object{
					Name:     title,
					IsFolder: true,
					ID:       "https://njav.tv/zh/" + href,
					Size:     622857143,
					Modified: parse,
				},
			})
		})

	})

	err := collector.Visit(urlFunc(index))

	return data, preLen != len(data), err

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

func (d *Javdb) getNamingAddr(films []model.ObjThumb) string {

	actorUrl := ""
	actorPageUrl := ""

	for i := range 3 {
		code := strings.Split(films[i].Name, " ")[0]

		searchResult, _, err := d.getNamingPageInfo(func(index int) string {
			return fmt.Sprintf("https://njav.tv/zh/search?keyword=%s", code)
		}, 1, []model.ObjThumb{})
		if err != nil {
			return ""
		}
		if len(searchResult) > 0 && strings.Split(searchResult[0].Name, " ")[0] == code {
			actorUrl = searchResult[0].ID
			break
		}

	}

	if actorUrl != "" {
		collector := colly.NewCollector(func(c *colly.Collector) {
			c.SetRequestTimeout(time.Second * 10)
		})

		collector.OnHTML("#details", func(element *colly.HTMLElement) {

			//href := element.ChildAttr(".thumb a", "href")
			//title := element.ChildText(".detail a")

			url := element.ChildAttr(".content a", "href")
			if url != "" {
				actorPageUrl = fmt.Sprintf("https://njav.tv/zh/%s?page=", url)
			}

		})

		err := collector.Visit(actorUrl)
		if err != nil {
			utils.Log.Info("演员主页爬取失败", err)
		}

	}

	return actorPageUrl

}
