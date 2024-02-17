package miss_av

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/gocolly/colly/v2"
	"gorm.io/gorm/utils"
	"net/http"
	"strings"
	"time"
)

func (d *MIssAV) getFilms(dirName string, urlFunc func(index int) string) ([]model.Obj, error) {

	results := make([]model.Obj, 0)
	data := make([]model.ObjThumb, 0)

	data, nextPage, err := d.getPageInfo(urlFunc, 1, data)
	if err != nil {
		return results, err
	}

	var urls []string
	for _, item := range data {
		urls = append(urls, item.ID)
	}

	existFilms := db.QueryByUrls(dirName, urls)

	// not exists
	for index := 2; index <= 20 && nextPage && len(existFilms) == 0; index++ {

		data, nextPage, err = d.getPageInfo(urlFunc, index, data)
		//films, images, urls, dates, nextPage, err = d.getPageInfo(urlFunc, index, films, images, urls, dates)
		if err != nil {
			return results, err
		}
		clear(urls)
		for _, item := range data {
			urls = append(urls, item.ID)
		}

		existFilms = db.QueryByUrls(dirName, urls)

	}
	// exist
	for index, item := range data {
		if utils.Contains(existFilms, item.ID) {
			if index == 0 {
				data = []model.ObjThumb{}
			} else {
				data = data[:index]
			}
			break
		}
	}

	if len(data) != 0 {
		err = db.CreateFilms("javdb", dirName, data)
		if err != nil {
			return results, nil
		}
	}

	return d.convertFilm(dirName, db.QueryByActor("javdb", dirName), results), nil

}

func (d *MIssAV) convertFilm(dirName string, actor []model.Film, results []model.Obj) []model.Obj {
	for _, film := range actor {
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     film.Name,
				IsFolder: true,
				ID:       film.Url,
				Size:     622857143,
				Modified: film.Date,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		})
		results = append(results, &model.ObjThumb{
			Object: model.Object{
				Name:     film.Name + ".jpg",
				IsFolder: false,
				ID:       film.Image,
				Size:     622857143,
				Modified: film.Date,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		})
	}
	return results
}

func (d *MIssAV) getMagnet(file model.Obj) (string, error) {

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

func (d *MIssAV) getPageInfo(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error) {

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
