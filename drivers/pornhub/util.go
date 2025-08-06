package pornhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/spider"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/robertkrimen/otto"
	"github.com/tebeka/selenium"
	"regexp"
	"strings"
	"time"
)

var viewKeyCompile = regexp.MustCompile(`/view_video.php\?viewkey=([^&\s]+)`)

func (d *Pornhub) getFilms(dirName, pageKey string) ([]model.EmbyFileObj, error) {

	var filmIds []string
	var films []model.EmbyFileObj

	key := ""

	if strings.Contains(pageKey, "/playlist/") {
		key = strings.ReplaceAll(pageKey, "/playlist/", "")
		playListFilms, err := d.getPlayListFilms(key, dirName)
		if err != nil {
			return virtual_file.GetStorageFilms("pornhub", dirName, false), nil
		}
		films = playListFilms
	} else {
		key = strings.ReplaceAll(pageKey, "/model/", "")
		actorFilms, err := d.getActorFilms(dirName, key)
		if err != nil {
			return virtual_file.GetStorageFilms("pornhub", dirName, false), nil
		}
		films = actorFilms
	}

	if len(films) == 0 {
		return virtual_file.GetStorageFilms("pornhub", dirName, false), nil
	}

	for _, film := range films {
		filmIds = append(filmIds, film.Url)
	}

	unSaveFilmIds := db.QueryUnSaveFilms(filmIds, dirName)
	if len(unSaveFilmIds) == 0 {
		return virtual_file.GetStorageFilms("pornhub", dirName, false), nil
	}

	unSaveFilmMap := make(map[string]bool)
	for _, filmId := range unSaveFilmIds {
		unSaveFilmMap[filmId] = true
	}

	utils.Log.Infof("saving porn films：%v", unSaveFilmIds)
	var notExitedFilms []model.EmbyFileObj

	for _, film := range films {
		if _, exist := unSaveFilmMap[film.Url]; exist {
			virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
				Source:   "pornhub",
				Dir:      dirName,
				FileName: virtual_file.AppendImageName(film.Url),
				Title:    film.Title,
				ImgUrl:   film.Thumb(),
				Actors:   film.Actors,
				Release:  film.ReleaseTime,
				Tags:     film.Tags,
			})
			notExitedFilms = append(notExitedFilms, film)
		}
	}

	err := db.CreateFilms("pornhub", dirName, dirName, notExitedFilms)
	if err != nil {
		return nil, err
	}

	return virtual_file.GetStorageFilms("pornhub", dirName, false), nil

}

func (d *Pornhub) getVideoLink(viewKey string) (string, error) {

	client := resty.New()
	res, err := client.R().SetQueryParam("viewkey", viewKey).Get(fmt.Sprintf("%s/view_video.php", d.ServerUrl))
	if err != nil {
		utils.Log.Warnf("failed to get film page info from pornhub, %s", err.Error())
		return "", err
	}

	html := res.String()

	scriptRegexp := regexp.MustCompile(`<script\b[^>]*>([\s\S]*?)</script>`)

	matchers := scriptRegexp.FindAllStringSubmatch(html, -1)
	var encryptedScript string

	for _, scripts := range matchers {
		script := scripts[1]
		if !strings.Contains(script, "flashvars_") {
			continue
		} else {
			encryptedScript = script
			break
		}
	}

	flashId := regexp.MustCompile(`flashvars_\d+`).FindString(encryptedScript)

	vm := otto.New()
	_, err = vm.Run(`var playerObjList = {};` + encryptedScript + fmt.Sprintf(`;var __VM__OUTPUT = JSON.stringify(%s.mediaDefinitions)`, flashId))
	if err != nil {
		utils.Log.Warnf("failed to run script, %s", err.Error())
		return "", err
	}

	value, err := vm.Get("__VM__OUTPUT")
	if err != nil {
		utils.Log.Warnf("failed to get console result, %v", err)
		return "", err
	}

	type MediaDefinition struct {
		Format   string `json:"format"`
		VideoURL string `json:"videoUrl"`
	}

	mediaDefinitions := make([]MediaDefinition, 0)

	if str, err1 := value.ToString(); err1 != nil {
		return "", err
	} else {
		if err2 := json.Unmarshal([]byte(str), &mediaDefinitions); err2 != nil {
			return "", err
		}
	}

	var mp4MediaDefinition *MediaDefinition

	for _, mediaDefinition := range mediaDefinitions {
		if mediaDefinition.Format == "mp4" {
			mp4MediaDefinition = &mediaDefinition
		}
	}

	if mp4MediaDefinition == nil {
		return "", errors.New("failed to find mp4 video")
	}

	pornVideos := make([]videoInfo, 0)

	_, err = client.R().SetHeaders(map[string]string{
		"Referer": mp4MediaDefinition.VideoURL,
	}).SetResult(&pornVideos).Get(mp4MediaDefinition.VideoURL)

	if err != nil {
		return "", err
	} else if len(pornVideos) == 0 {
		return "", errors.New("failed to find mp4 video")
	}

	return pornVideos[len(pornVideos)-1].VideoURL, nil

}

func (d *Pornhub) getPlayListFilms(playlistId, dirName string) ([]model.EmbyFileObj, error) {

	var films []PornFilm

	err := spider.Visit(d.SpiderServer, fmt.Sprintf("%s/playlist/%s", d.ServerUrl, playlistId), time.Duration(d.SpiderMaxWaitTime)*time.Second, func(wd selenium.WebDriver) {

		source, _ := wd.PageSource()

		compile := regexp.MustCompile("token=(.*)\"")
		tokenStr := compile.FindString(source)
		token := compile.ReplaceAllString(tokenStr, "$1")

		page := 1
		preLen := len(films)

		for page == 1 || preLen != len(films) {

			preLen = len(films)

			err := wd.Get(fmt.Sprintf("%s/playlist/viewChunked?id=%s&token=%s&page=%d", d.ServerUrl, playlistId, token, page))
			if err != nil {
				utils.Log.Warnf("failed to get playlist, %s", err.Error())
				return
			}

			time.Sleep(time.Duration(d.SpiderMaxWaitTime) * time.Second)

			films = append(films, resolveFilms(wd, PlayList)...)
			page++

		}

	})

	if err != nil {
		utils.Log.Warnf("failed to get pornhub films, %s", err.Error())
		return nil, err
	}

	embyFileObjs, err := convertFilms(dirName, films)
	if len(embyFileObjs) > 0 {
		for i := range embyFileObjs {
			embyFileObjs[i].Tags = append(embyFileObjs[i].Tags, dirName)
		}
	}

	return embyFileObjs, err

}

func (d *Pornhub) getActorFilms(dirName, actor string) ([]model.EmbyFileObj, error) {

	var films []PornFilm
	page := 1
	pageUrl := fmt.Sprintf("%s/model/%s/videos?o=mr&page=%d", d.ServerUrl, actor, page)

	err := spider.Visit(d.SpiderServer, pageUrl, time.Duration(d.SpiderMaxWaitTime)*time.Second, func(wd selenium.WebDriver) {

		var newFilmIds []string

		newFilms := resolveFilms(wd, ACTOR)
		films = append(films, newFilms...)
		page++

		for _, film := range newFilms {
			newFilmIds = append(newFilmIds, film.ViewKey)
		}

		nextPageFunc := func() bool {
			nextPage := false

			_, err := wd.FindElement(selenium.ByCSSSelector, ".page_next.omega")
			if err == nil {
				// find next button
				nextPage = true
				_, err1 := wd.FindElement(selenium.ByCSSSelector, ".page_next.disabled.omega")
				if err1 == nil {
					// next page is disabled
					nextPage = false
				}
			}

			return nextPage
		}

		for nextPageFunc() && len(db.QueryUnSaveFilms(newFilmIds, dirName)) > 0 {

			pageUrl = fmt.Sprintf("%s/model/%s/videos?page=%d", d.ServerUrl, actor, page)

			err := wd.Get(pageUrl)
			if err != nil {
				utils.Log.Warnf("failed to get actor films, %s", err.Error())
				return
			}

			time.Sleep(time.Duration(d.SpiderMaxWaitTime) * time.Second)

			pornFilms := resolveFilms(wd, ACTOR)
			clear(newFilmIds)
			for _, film := range pornFilms {
				newFilmIds = append(newFilmIds, film.ViewKey)
			}

			films = append(films, pornFilms...)
			page++

		}

	})

	if err != nil {
		utils.Log.Warnf("failed to get pornhub films, %s", err.Error())
		return nil, err
	}

	return convertFilms(actor, films)

}

func resolveFilms(wd selenium.WebDriver, actorType int) []PornFilm {

	var films []PornFilm

	var parentEle selenium.WebElement
	var err error

	if actorType == PlayList {
		parentEle, err = wd.FindElement(selenium.ByTagName, "body")
	} else {
		parentEle, err = wd.FindElement(selenium.ByCSSSelector, ".videoUList")
	}

	if err != nil {
		utils.Log.Warnf("failed to find parent element, %s", err.Error())
		return films
	}

	filmElements, _ := parentEle.FindElements(selenium.ByCSSSelector, ".wrap.flexibleHeight")
	for _, filmElement := range filmElements {

		aEles, err2 := filmElement.FindElements(selenium.ByCSSSelector, "a")
		if err2 != nil {
			utils.Log.Warnf("failed to find a elements, %s", err2.Error())
		}

		href := ""
		title := ""

		for i := 1; i < len(aEles) && (href == "" || title == ""); i++ {

			if tempHref, _ := aEles[i].GetAttribute("href"); tempHref != "" {
				href = tempHref
			}
			if tempTitle, _ := aEles[i].GetAttribute("title"); tempTitle != "" {
				title = tempTitle
			}

		}

		imgEle, err1 := filmElement.FindElement(selenium.ByCSSSelector, "img")
		if err1 != nil {
			utils.Log.Warnf("failed to find img, %s", err1.Error())
			return films
		}
		imgSrc, err1 := imgEle.GetAttribute("src")
		if err1 != nil {
			utils.Log.Warnf("failed to find src, %s", err1.Error())
			return films
		}

		username := ""
		usernameEle, err1 := filmElement.FindElement(selenium.ByCSSSelector, ".usernameBadgesWrapper")
		if err1 == nil {
			username, err1 = usernameEle.Text()
			if err1 != nil {
				utils.Log.Warnf("failed to find username, %s", err1.Error())
				return films
			}
		}
		viewKeyCompile.ReplaceAllString(href, "$1")
		films = append(films, PornFilm{
			Image: imgSrc,
			Title: title,
			ViewKey: func() string {
				findString := viewKeyCompile.FindString(href)
				return viewKeyCompile.ReplaceAllString(findString, "$1")
			}(),
			Username: username,
		})

	}
	return films
}

func convertFilms(actor string, films []PornFilm) ([]model.EmbyFileObj, error) {
	return utils.SliceConvert(films, func(src PornFilm) (model.EmbyFileObj, error) {
		return model.EmbyFileObj{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					Name:     src.ViewKey,
					IsFolder: false,
					Size:     622857143,
					Modified: time.Now(),
				},
				Thumbnail: model.Thumbnail{Thumbnail: src.Image},
			},
			ReleaseTime: time.Now(),
			Url:         src.ViewKey,
			Actors: func() []string {
				if src.Username != "" {
					return []string{src.Username}
				}
				return []string{actor}
			}(),
			Title: src.Title,
			Tags:  []string{"pornhub"},
		}, nil
	})
}
