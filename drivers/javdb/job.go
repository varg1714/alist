package javdb

import (
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/av"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"strconv"
	"strings"
	"time"
)

func (d *Javdb) reMatchSubtitles() {

	utils.Log.Info("start rematching subtitles for films without subtitles")

	caches, err := db.QueryNoSubtitlesCache(DriverName)
	if err != nil {
		utils.Log.Warnf("failed to query the films without subtitles")
		return
	}
	if len(caches) != 0 {
		var savingCaches []model.MagnetCache
		var unFindCaches []model.MagnetCache

		for _, cache := range caches {

			film, err1 := db.QueryFilmByCode(DriverName, cache.Code)
			if err1 != nil {
				utils.Log.Warn("failed to query film:", err1.Error())
			} else {
				if film.Url != "" {
					javdbMeta, err2 := av.GetMetaFromJavdb(film.Url)
					if err2 != nil {
						utils.Log.Warn("failed to get javdb magnet info:", err2.Error())
					} else if len(javdbMeta.Magnets) > 0 && javdbMeta.Magnets[0].IsSubTitle() {
						cache.Subtitle = true
						cache.Magnet = javdbMeta.Magnets[0].GetMagnet()
					}
				}
			}

			if !cache.Subtitle {
				sukeMeta, err2 := av.GetMetaFromSuke(cache.Code)
				if err2 != nil {
					utils.Log.Warn("failed to get suke magnet info:", err2.Error())
				} else {
					if len(sukeMeta.Magnets) > 0 && sukeMeta.Magnets[0].IsSubTitle() {
						cache.Subtitle = true
						cache.Magnet = sukeMeta.Magnets[0].GetMagnet()
					}
				}
			}

			if cache.Subtitle {
				cache.ScanAt = time.Now()
				savingCaches = append(savingCaches, cache)
			} else {
				unFindCaches = append(unFindCaches, cache)
			}

		}

		if len(savingCaches) > 0 {
			err2 := db.BatchCreateMagnetCache(savingCaches)
			if err2 != nil {
				utils.Log.Warn("failed to create magnet cache:", err2.Error())
			}
			utils.Log.Infof("update films magnet cache:[%v]", savingCaches)
		}

		if len(unFindCaches) > 0 {
			var names []string
			for _, cache := range unFindCaches {
				names = append(names, cache.Name)
			}
			err2 := db.UpdateScanData(DriverName, names, time.Now())
			if err2 != nil {
				utils.Log.Warn("failed to update scan data:", err2.Error())
			}
			utils.Log.Infof("films:[%v] still have not matched with subtitles, update the scan info", names)
		}
	}

	noMatchCaches, err2 := db.QueryNoMatchCache(DriverName)
	if err2 != nil {
		utils.Log.Warn("failed to query film:", err2.Error())
		return
	}

	if len(noMatchCaches) > 0 {
		deletingCache := make(map[string][]string)
		for _, cache := range noMatchCaches {
			deletingCache[cache.DriverType] = append(deletingCache[cache.DriverType], cache.Name)
		}

		for driverType, names := range deletingCache {
			err3 := db.DeleteCacheByName(driverType, names)
			if err3 != nil {
				utils.Log.Warn("failed to delete cache:", err3.Error())
			}
		}
		utils.Log.Infof("Delete the cached films that do not match the subtitles:[%v]", noMatchCaches)
	}

	utils.Log.Info("rematching completed")

}

func (d *Javdb) refreshNfo() {

	utils.Log.Info("start refresh nfo for javdb")

	var actorNames []string
	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		actorNames = append(actorNames, actor.Name)
	}
	for _, actor := range actorNames {

		films := virtual_file.GetStorageFilms(DriverName, actor, false)

		// refresh nfo
		mappingNameFilms, err := d.mappingNames(actor, films)
		if err != nil {
			utils.Log.Warn("failed to get mapping names:", err.Error())
			continue
		}

		var filmNames []string
		for _, film := range mappingNameFilms {
			virtual_file.UpdateNfo(virtual_file.MediaInfo{
				Source:   DriverName,
				Dir:      film.Path,
				FileName: virtual_file.AppendImageName(film.Name),
				Release:  film.ReleaseTime,
				Title:    film.Title,
			})
			filmNames = append(filmNames, film.Name)
		}

		// clear unused files
		virtual_file.ClearUnUsedFiles(DriverName, actor, filmNames)

	}

	utils.Log.Info("finish refresh nfo")

}

func (d *Javdb) filterFilms() {

	utils.Log.Info("start to filter javdb films")

	films, err := db.QueryFilmsByNamePrefix(DriverName, strings.Split(d.Filter, ","))
	if err != nil {
		utils.Log.Warn("failed to query films:", err.Error())
		return
	}

	if len(films) > 0 {
		utils.Log.Infof("deleting films:[%v]", films)
		for _, film := range films {
			err1 := d.deleteFilm(film.Actor, virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(film.Name))), film.Url)
			if err1 != nil {
				utils.Log.Warn("failed to delete film:", err1.Error())
			}
		}
	}

	utils.Log.Info("finish filter javdb films")

}

func (d *Javdb) reMatchTags() {

	if d.MatchFilmTagLimit <= 0 {
		return
	}

	utils.Log.Info("start to match javdb tags")

	films, err := db.QueryNoTagFilms(DriverName, d.MatchFilmTagLimit)
	if err != nil {
		utils.Log.Warn("failed to query films:", err.Error())
		return
	}

	for _, film := range films {
		file := virtual_file.ConvertFilmToEmbyFile(film, "")
		_, err1 := d.getMagnet(&file, true)
		if err1 != nil {
			utils.Log.Infof("failed to get film: %s tag info: %s", film.Name, err1.Error())
			return
		}
		time.Sleep(3 * time.Second)
	}

	utils.Log.Info("finish match javdb tags")

}
