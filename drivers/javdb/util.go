package javdb

import (
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/gocolly/colly/v2"
	"net/http"
	"strings"
	"time"
)

func (d *Javdb) getFilms(dirName string, urlFunc func(index int) string) ([]model.Obj, error) {

	return virtual_file.GetFilmsWitchStorage("javdb", dirName, urlFunc,
		func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {
			return d.getPageInfo(urlFunc, index, data)
		})

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

func (d *Javdb) getPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

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
