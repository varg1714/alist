package av

import (
	"cmp"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly/v2"
	"slices"
	"strconv"
	"strings"
	"time"
)

func GetMetaFromJavdb(filmUrl string) (Meta, error) {

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	var meta Meta

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
				utils.Log.Infof("failed to format file size:%s,error message:%v", fileSizeText, err)
			}

			meta.Magnets = append(meta.Magnets, Magnet{
				Magnet: magnetEle.ChildAttr("a", "href"),
				Tags:   tags,
				Size:   bytes,
				Subtitle: func() bool {
					for _, tag := range tags {
						if tag == "字幕" {
							return true
						}
					}
					return false
				}(),
				Source: filmUrl,
			})

		})
	})

	collector.OnHTML(".panel.movie-panel-info", func(element *colly.HTMLElement) {
		element.ForEach("a", func(i int, element *colly.HTMLElement) {

			href := element.Attr("href")
			if strings.Contains(href, "/actors/") {
				actorUrl := strings.ReplaceAll(href, "/actors/", "")
				meta.Actors = append(meta.Actors, Actor{
					Id:   actorUrl,
					Name: element.Text,
				})
			}

		})
	})

	err := collector.Visit(filmUrl)
	if err != nil {
		return meta, err
	}

	sortMagnet(&meta)

	return meta, nil

}

func GetMetaFromSuke(code string) (Meta, error) {

	searchUrl := fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%s&s=downloads&o=desc", code)
	var meta Meta

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	collector.OnHTML(`.table-responsive tbody`, func(element *colly.HTMLElement) {

		element.ForEach("tr", func(i int, trElement *colly.HTMLElement) {

			// title
			href := trElement.ChildAttr("td:nth-child(2) a", "href")
			title := strings.ReplaceAll(trElement.ChildAttr("td:nth-child(2) a", "title"), "+++ ", "")

			// size
			sizeStr := trElement.ChildText("td:nth-child(4)")
			size, _ := humanize.ParseBytes(sizeStr)

			// download count
			downloadStr := trElement.ChildText("td:nth-child(8)")
			count, _ := strconv.ParseInt(downloadStr, 10, 64)

			meta.Magnets = append(meta.Magnets, Magnet{
				Name:          strings.ReplaceAll(title, "+++", ""),
				Subtitle:      strings.Contains(title, "中文字幕"),
				DownloadCount: uint64(count),
				Size:          size,
				Source:        fmt.Sprintf("https://sukebei.nyaa.si/%s", href),
			})

		})

	})

	err := collector.Visit(searchUrl)
	if err != nil {
		return meta, err
	}

	sortMagnet(&meta)
	if len(meta.Magnets) == 0 {
		return meta, nil
	}

	err = findSukeMagnetInfo(&meta.Magnets[0])

	return meta, nil

}

func findSukeMagnetInfo(magnet *Magnet) error {

	if magnet == nil || magnet.Source == "" {
		return nil
	}

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 10)
	})

	// 1. get magnet info
	collector.OnHTML(".card-footer-item", func(element *colly.HTMLElement) {
		magnet.Magnet = element.Attr("href")
	})

	// 2. get file list from the magnet
	collector.OnHTML(`.torrent-file-list.panel-body ul[data-show="yes"]`, func(element *colly.HTMLElement) {
		element.ForEach("li", func(i int, liElement *colly.HTMLElement) {

			fileName := liElement.Text
			fileSize := liElement.ChildText(".file-size")

			if len(fileSize) > 2 {
				bytes, err := humanize.ParseBytes(fileSize[1 : len(fileSize)-1])
				if err != nil {
					utils.Log.Warnf("failed to format file size:%s, error message:%s", fileSize, err.Error())
				}
				if bytes/(1024*1024) > 100 {
					magnet.Files = append(magnet.Files, File{
						Size: bytes,
						Name: fileName,
					})
				}
			}
		})
	})

	err := collector.Visit(magnet.Source)
	if err != nil {
		utils.Log.Warn("failed to get the magnet info from suke:", err.Error())
		return err
	}

	return nil

}

func sortMagnet(meta *Meta) {

	slices.SortFunc(meta.Magnets, func(a, b Magnet) int {

		if a.Subtitle && !b.Subtitle {
			return -1
		} else if !a.Subtitle && b.Subtitle {
			return 1
		}

		tagCmp := cmp.Compare(len(b.Tags), len(a.Tags))

		if tagCmp != 0 {
			return tagCmp
		}

		countCmp := cmp.Compare(b.DownloadCount, a.DownloadCount)
		if countCmp != 0 {
			return countCmp
		}

		return cmp.Compare(b.Size, a.Size)

	})

}
