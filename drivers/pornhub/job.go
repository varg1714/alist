package pornhub

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"time"
)

func (d *Pornhub) reMatchTags() {

	if d.MatchFilmTagLimit <= 0 {
		return
	}

	utils.Log.Info("start to match porn tags")
	defer utils.Log.Info("finish match porn tags")

	films, err := db.QueryNotMatchTagFilms(DriverName, []string{}, DriverName, d.MatchFilmTagLimit)
	if err != nil {
		utils.Log.Warn("failed to query films:", err.Error())
		return
	}

	for _, film := range films {

		collector := colly.NewCollector(func(c *colly.Collector) {
			c.SetRequestTimeout(time.Second * 10)
		})

		tagMap := make(map[string]bool)
		for _, tag := range film.Tags {
			tagMap[tag] = true
		}

		collector.OnHTML(".tagsWrapper", func(tagEle *colly.HTMLElement) {

			tagEle.ForEach(".gtm-event-video-underplayer.item span", func(i int, element *colly.HTMLElement) {

				if !tagMap[element.Text] {
					film.Tags = append(film.Tags, element.Text)
					tagMap[element.Text] = true
				}

			})

		})

		err1 := collector.Visit(fmt.Sprintf("%s/view_video.php?viewkey=%s", d.ServerUrl, film.Url))

		if err1 != nil {
			utils.Log.Infof("failed to get film: %s tag info, error message: %s", film.Name, err1.Error())
			return
		}

		film.Tags = append(film.Tags, DriverName)
		err1 = db.UpdateFilm(film)
		if err1 != nil {
			utils.Log.Infof("failed to update film: %s tag info, error: %s", film.Name, err1.Error())
			return
		}

		embyObj := virtual_file.ConvertFilmToEmbyFile(film, film.Actor)

		virtual_file.UpdateNfo(virtual_file.MediaInfo{
			Source:   DriverName,
			Dir:      embyObj.Path,
			FileName: virtual_file.AppendImageName(embyObj.Name),
			Title:    embyObj.Title,
			Actors:   embyObj.Actors,
			Release:  embyObj.ReleaseTime,
			Tags:     embyObj.Tags,
		})

		time.Sleep(3 * time.Second)
	}

}
