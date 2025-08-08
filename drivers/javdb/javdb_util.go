package javdb

import (
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"strconv"
	"strings"
	"time"
)

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

	url := urlFunc(index)
	err := collector.Visit(url)
	utils.Log.Debugf("开始爬取javdb页面：%s，错误：%v", url, err)

	return data, nextPage, err

}

func (d *Javdb) fetchFilmMeta(filmUrl string, embyFile *model.EmbyFileObj) {

	existActors := db.QueryActor(strconv.Itoa(int(d.ID)))
	mapping := make(map[string]string)
	for _, actor := range existActors {
		mapping[actor.Url] = actor.Name
	}

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
			} else if strings.Contains(href, "/tags") {
				embyFile.Tags = append(embyFile.Tags, element.Text)
			}

		})
	})

	err := collector.Visit(filmUrl)

	if err != nil {
		utils.Log.Warnf("演员信息获取失败:%s", err.Error())
		return
	}

	var actors []string
	for url, name := range actorMapping {
		if mapping[url] != "" {
			actors = append(actors, mapping[url])
		} else {
			actors = append(actors, name)
		}
	}

	for _, actor := range actors {
		embyFile.Actors = append(embyFile.Actors, actor)
	}

}
