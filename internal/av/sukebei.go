package av

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly/v2"
	"strconv"
	"strings"
	"time"
)

type SukebeiMagnet struct {
	magnet        string
	name          string
	size          uint64
	subtitle      bool
	downloadCount uint64
	files         []File
	releaseDate   time.Time
	source        string
	queried       bool
}

func (s *SukebeiMagnet) GetMagnet() string {
	if !s.queried {
		s.findSukeMagnetInfo()
	}
	return s.magnet
}

func (s *SukebeiMagnet) GetName() string {
	return s.name
}

func (s *SukebeiMagnet) GetSize() uint64 {
	return s.size
}

func (s *SukebeiMagnet) IsSubTitle() bool {
	return s.subtitle
}

func (s *SukebeiMagnet) GetTags() []string {
	return []string{}
}

func (s *SukebeiMagnet) GetDownloadCount() uint64 {
	return s.downloadCount
}

func (s *SukebeiMagnet) GetFiles() []File {
	if !s.queried {
		s.findSukeMagnetInfo()
	}
	return s.files
}

func (s *SukebeiMagnet) GetReleaseDate() time.Time {
	return s.releaseDate
}

func GetMetaFromSuke(code string) (Meta, error) {

	code = GetFilmCode(code)

	searchUrl := fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", code)
	var meta Meta

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(`.table-responsive tbody`, func(element *colly.HTMLElement) {

		element.ForEach("tr", func(i int, trElement *colly.HTMLElement) {

			// title
			href := trElement.ChildAttr("td:nth-child(2) a:last-of-type", "href")
			title := strings.ReplaceAll(trElement.ChildAttr("td:nth-child(2) a:last-of-type", "title"), "+++ ", "")

			// size
			size := uint64(0)
			sizeStr := trElement.ChildText("td:nth-child(4)")
			if !strings.Contains(sizeStr, "bytes") {
				size, _ = humanize.ParseBytes(sizeStr)
			}

			// time
			timeStr := trElement.ChildText("td:nth-child(5)")
			releaseTime, err := time.Parse("2006-01-02 15:04", timeStr)
			if err != nil {
				utils.Log.Infof("failed to parse release time:%s,error message:%v", timeStr, err)
			}

			// download count
			downloadStr := trElement.ChildText("td:nth-child(8)")
			count, _ := strconv.ParseInt(downloadStr, 10, 64)

			meta.Magnets = append(meta.Magnets, &SukebeiMagnet{
				name:          strings.ReplaceAll(title, "+++", ""),
				subtitle:      strings.Contains(title, "中文字幕"),
				downloadCount: uint64(count),
				size:          size,
				source:        fmt.Sprintf("https://sukebei.nyaa.si/%s", href),
				releaseDate:   releaseTime,
			})

		})

	})

	var respErr error
	collector.OnError(func(_ *colly.Response, err error) {
		respErr = err
	})

	err := collector.Visit(searchUrl)
	if err != nil {
		return meta, err
	} else if respErr != nil {
		return meta, respErr
	}

	if len(meta.Magnets) == 0 {
		return meta, nil
	}
	sortMagnet(&meta)

	return meta, nil

}

func (s *SukebeiMagnet) findSukeMagnetInfo() {

	if s.source == "" {
		return
	}

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	// 1. get magnet info
	collector.OnHTML(".card-footer-item", func(element *colly.HTMLElement) {
		s.magnet = element.Attr("href")
	})

	// 2. get file list from the magnet
	collector.OnHTML(`.torrent-file-list.panel-body ul[data-show="yes"]`, func(element *colly.HTMLElement) {
		element.ForEach("li", func(i int, liElement *colly.HTMLElement) {

			fileName := liElement.Text
			fileSize := liElement.ChildText(".file-size")

			if len(fileSize) > 2 && !strings.Contains(fileSize, "Bytes") {
				bytes, err := humanize.ParseBytes(fileSize[1 : len(fileSize)-1])
				if err != nil {
					utils.Log.Warnf("failed to format file size:%s, error message:%s", fileSize, err.Error())
				}
				if bytes/(1024*1024) > 100 {
					s.files = append(s.files, File{
						Size: bytes,
						Name: fileName,
					})
				}
			}
		})
	})

	err := collector.Visit(s.source)
	if err != nil {
		utils.Log.Warn("failed to get the magnet info from suke:", err.Error())
	}
	s.queried = true

}
