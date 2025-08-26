package fc2

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/gocolly/colly/v2"
	"strings"
	"time"
)

func (d *FC2) getMissAvPageFilms(url string) []model.EmbyFileObj {

	var films []model.EmbyFileObj

	collector := colly.NewCollector(func(c *colly.Collector) {
		c.SetRequestTimeout(time.Second * 30)
	})

	collector.OnHTML(".grid", func(table *colly.HTMLElement) {
		table.ForEach(".thumbnail.group", func(i int, element *colly.HTMLElement) {
			fc2Id := element.ChildAttr(".text-secondary", "alt")
			fc2Id = strings.ReplaceAll(fc2Id, "fc2-ppv", "FC2-PPV")
			films = append(films, model.EmbyFileObj{
				ObjThumb: model.ObjThumb{
					Object: model.Object{
						Name: fc2Id,
					},
				},
				Title: element.ChildText(".text-secondary"),
			})
		})
	})

	retryCount := 1
	collector.OnError(func(r *colly.Response, err error) {
		utils.Log.Infof("request failed, retryCount: %d", retryCount)
		if retryCount <= 3 {
			retryCount++
			err = r.Request.Retry()
			if err != nil {
				utils.Log.Warnf("retry failed: %s", err.Error())
			}
		}
	})

	err := collector.Visit(url)
	if err != nil {
		utils.Log.Warnf("failed to visit page: %s, error: %s", url, err.Error())
	}

	return films

}

func (d *FC2) getMissAvFilms(dirName string, urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	var queriedFilms []model.EmbyFileObj
	page := 1
	preSize := len(queriedFilms)

	// 1. query films
	for page == 1 || (preSize != len(queriedFilms) && page <= d.MissAvMaxPage) {

		preSize = len(queriedFilms)
		films := d.getMissAvPageFilms(urlFunc(page))
		queriedFilms = append(queriedFilms, films...)
		page++

	}

	filmMap := make(map[string]model.EmbyFileObj)
	var fc2Ids []string

	// 2. add tag
	for i := range queriedFilms {
		queriedFilms[i].Tags = append(queriedFilms[i].Tags, fmt.Sprintf("%s-Top%d", dirName, ((i/30)+1)*30), dirName)
		filmMap[queriedFilms[i].Name] = queriedFilms[i]
		fc2Ids = append(fc2Ids, queriedFilms[i].Name)
	}

	// 3. save film
	unCachedFilms := db.QueryNoMagnetFilms(fc2Ids)
	unCachedFilmMap := make(map[string]bool)
	for _, filmId := range unCachedFilms {
		unCachedFilmMap[filmId] = true
	}

	if len(unCachedFilms) > 0 {
		utils.Log.Infof("import new films：%v", unCachedFilms)
	}

	var notExitedFilms []string
	cachedFilmMap := make(map[string][]string)

	for _, film := range queriedFilms {
		if unCachedFilmMap[film.Name] {
			// add film
			_, err := d.addStar(film.Name, film.Tags)
			if err != nil {
				notExitedFilms = append(notExitedFilms, film.Name)
			}
		} else {
			// update film
			cachedFilmMap[film.Tags[0]] = append(cachedFilmMap[film.Tags[0]], film.Name)
		}
	}

	// 4. update tag
	if len(cachedFilmMap) > 0 {
		for tag, ids := range cachedFilmMap {
			notMatchTagFilms, err := db.QueryNotMatchTagFilms("fc2", ids, tag, 0)
			if err != nil {
				utils.Log.Warnf("failed to query notMatchTagFilms: %s", err.Error())
			} else {
				for _, film := range notMatchTagFilms {
					film.Tags = append(film.Tags, tag)
					err2 := db.UpdateFilm(film)
					if err2 != nil {
						utils.Log.Warnf("failed to update film: %s, error message: %s", film.Name, err2.Error())
					}
				}
			}
		}
	}

	return []model.EmbyFileObj{}, nil

}
