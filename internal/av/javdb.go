package av

import (
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly/v2"
	"strings"
	"time"
)

type JavdbMagnet struct {
	magnet        string
	name          string
	size          uint64
	subtitle      bool
	tags          []string
	downloadCount uint64
	files         []File
	releaseDate   time.Time
}

func (m *JavdbMagnet) IsSubTitle() bool {
	return m.subtitle
}

func (m *JavdbMagnet) GetTags() []string {
	return m.tags
}

func (m *JavdbMagnet) GetMagnet() string {
	return m.magnet
}

func (m *JavdbMagnet) GetName() string {
	return m.name
}

func (m *JavdbMagnet) GetSize() uint64 {
	return m.size
}

func (m *JavdbMagnet) GetDownloadCount() uint64 {
	return m.downloadCount
}

func (m *JavdbMagnet) GetFiles() []File {
	return m.files
}

func (m *JavdbMagnet) GetReleaseDate() time.Time {
	return m.releaseDate
}

func GetMetaFromJavdb(filmUrl string) (Meta, error) {

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	var meta Meta
	var filmTags []string

	collector.OnHTML(".panel.movie-panel-info", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {

			href := element.Attr("href")
			if strings.Contains(href, "/actors/") {
				actorUrl := strings.ReplaceAll(href, "/actors/", "")
				meta.Actors = append(meta.Actors, Actor{
					Id:   actorUrl,
					Name: element.Text,
				})
			} else if strings.Contains(href, "/tags") {
				filmTags = append(filmTags, element.Text)
			}

		})
	})

	collector.OnHTML(".magnet-links", func(element *colly.HTMLElement) {
		element.ForEach(".item", func(i int, magnetEle *colly.HTMLElement) {

			var tags []string
			tags = append(tags, filmTags...)

			magnetEle.ForEach(".tag", func(i int, tag *colly.HTMLElement) {
				tags = append(tags, tag.Text)
			})

			fileSizeText := magnetEle.ChildText(".meta")
			fileSize := strings.Split(fileSizeText, ",")[0]
			bytes, err := humanize.ParseBytes(fileSize)
			if err != nil {
				utils.Log.Infof("failed to format file size:%s,error message:%v", fileSizeText, err)
			}

			timeStr := magnetEle.ChildText(".time")
			var releaseTime time.Time
			if timeStr != "" {
				releaseTime, err = time.Parse(time.DateOnly, timeStr)
				if err != nil {
					utils.Log.Warnf("failed to parse release time:%s,error message:%v", timeStr, err)
				}
			}

			meta.Magnets = append(meta.Magnets, &JavdbMagnet{
				magnet: magnetEle.ChildAttr("a", "href"),
				tags:   tags,
				size:   bytes,
				subtitle: func() bool {
					for _, tag := range tags {
						if tag == "字幕" {
							return true
						}
					}
					return false
				}(),
				releaseDate: releaseTime,
			})

		})
	})

	err := collector.Visit(filmUrl)
	if err != nil {
		return meta, err
	}

	sortMagnet(&meta)

	return meta, nil

}
