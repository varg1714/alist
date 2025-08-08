package javdb

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"time"
)

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
