package javdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/av"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/offline_download/tool"
	"github.com/OpenListTeam/OpenList/v4/internal/open_ai"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (d *Javdb) getFilms(dirName string, urlFunc func(index int) string) ([]model.EmbyFileObj, error) {

	// 1. 获取所有影片
	javFilms, err := virtual_file.GetFilmsWithStorage(DriverName, dirName, dirName, urlFunc,
		func(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error) {
			return d.getJavPageInfo(urlFunc, index, data)
		}, virtual_file.Option{CacheFile: false, MaxPageNum: 20})

	if err != nil && len(javFilms) == 0 {
		utils.Log.Info("javdb影片获取失败", err)
		return javFilms, err
	}

	// 2. 根据影片名字映射名称
	return d.mappingNames(dirName, javFilms)

}

func (d *Javdb) mappingNames(dirName string, javFilms []model.EmbyFileObj) ([]model.EmbyFileObj, error) {

	if len(javFilms) == 0 {
		return javFilms, nil
	}

	var noTitleFilms []model.EmbyFileObj
	for _, film := range javFilms {
		if !film.Translated {
			noTitleFilms = append(noTitleFilms, film)
		}
	}
	if len(noTitleFilms) == 0 {
		return javFilms, nil
	}

	// 2.1 获取所有映射名称
	namingFilms, err := d.getAiravNamingFilms(noTitleFilms, dirName)
	if err != nil || len(namingFilms) == 0 {
		utils.Log.Info("中文影片名称获取失败", err)
		return javFilms, nil
	}

	// 2.2 进行映射
	var savingFilms []model.EmbyFileObj
	var deletingFilms []string

	for index, film := range javFilms {
		if !film.Translated {
			code := splitCode(film.Name)
			if newName, exist := namingFilms[code]; exist {
				javFilms[index].Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(newName)))
				javFilms[index].Title = virtual_file.ClearFilmName(newName)

				savingFilms = append(savingFilms, javFilms[index])
				deletingFilms = append(deletingFilms, film.Url)
			}
		}
	}
	if len(savingFilms) > 0 {
		err1 := db.DeleteFilmsByUrl(DriverName, dirName, deletingFilms)
		if err1 != nil {
			utils.Log.Warnf("failed to delete films:[%s], error message: %s", deletingFilms, err1.Error())
		} else {
			err2 := db.CreateFilms(DriverName, dirName, dirName, savingFilms)
			if err2 != nil {
				utils.Log.Infof("failed to save films:[%s], error message: %s", deletingFilms, err2.Error())
			}
		}
	}

	for _, film := range javFilms {
		created := virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
			Source:   DriverName,
			Dir:      dirName,
			FileName: virtual_file.AppendImageName(film.Name),
			Title:    film.Title,
			ImgUrl:   film.Thumb(),
			Actors:   []string{dirName},
			Release:  film.ReleaseTime,
		})

		if created == virtual_file.Exist && d.QuickCache {
			// 已经创建过了，后续不再创建
			break
		}

	}

	utils.Log.Infof("name mapping finished for:[%s]", dirName)

	return javFilms, err
}

func (d *Javdb) getStars() []model.EmbyFileObj {
	films := virtual_file.GetStorageFilms(DriverName, "个人收藏", true)

	if d.RefreshNfo {
		var filmNames []string
		for _, film := range films {
			filmNames = append(filmNames, film.Name)
		}
		virtual_file.ClearUnUsedFiles(DriverName, "个人收藏", filmNames)
	}

	return films
}

func (d *Javdb) addStar(code string, tags []string) (model.EmbyFileObj, error) {

	existFilms, err := db.QueryFilmsByNamePrefix(DriverName, []string{code})
	if err != nil {
		utils.Log.Warnf("failed to query films: [%s], error message: %s", tags, err.Error())
		return model.EmbyFileObj{}, err
	} else if len(existFilms) > 0 {
		existFilm := existFilms[0]
		if existFilm.Actor == "个人收藏" && len(tags) > 0 {
			d.updateExistFilm(&existFilm, []string{}, tags)
		} else if existFilm.Actor != "个人收藏" {
			d.updateExistFilm(&existFilm, []string{"个人收藏"}, tags)
		}
		return virtual_file.ConvertFilmToEmbyFile(existFilm, ""), nil
	}

	javFilms, _, err := d.getJavPageInfo(func(index int) string {
		return fmt.Sprintf("https://javdb.com/search?f=download&q=%s", code)
	}, 1, []model.EmbyFileObj{})
	if err != nil {
		utils.Log.Info("jav影片查询失败:", err)
		return model.EmbyFileObj{}, err
	}

	if len(javFilms) == 0 || strings.ToLower(code) != strings.ToLower(splitCode(javFilms[0].Name)) {
		return model.EmbyFileObj{}, errors.New(fmt.Sprintf("影片:%s未查询到", code))
	}

	cachingFilm := javFilms[0]
	_, airavFilm := d.getAiravNamingAddr(cachingFilm)
	if airavFilm.Name != "" {
		cachingFilm.Title = airavFilm.Title
		cachingFilm.Name = airavFilm.Name
	} else {
		tempCode, name := splitName(cachingFilm.Name)

		translatedText := open_ai.Translate(virtual_file.ClearFilmName(name))
		if translatedText != "" {
			translatedText = fmt.Sprintf("%s %s", tempCode, translatedText)
			cachingFilm.Name = translatedText
			cachingFilm.Title = translatedText
		}
	}

	d.fetchFilmMeta(cachingFilm.Url, &cachingFilm)
	cachingFilm.Actors = append(cachingFilm.Actors, "个人收藏")
	for _, tag := range tags {
		cachingFilm.Tags = append(cachingFilm.Tags, tag)
	}

	err = db.CreateFilms(DriverName, "个人收藏", "个人收藏", []model.EmbyFileObj{cachingFilm})
	cachingFilm.Name = virtual_file.AppendFilmName(virtual_file.CutString(virtual_file.ClearFilmName(cachingFilm.Name)))
	cachingFilm.Path = "个人收藏"

	_ = virtual_file.CacheImageAndNfo(virtual_file.MediaInfo{
		Source:   DriverName,
		Dir:      "个人收藏",
		FileName: virtual_file.AppendImageName(cachingFilm.Name),
		Title:    cachingFilm.Title,
		ImgUrl:   cachingFilm.Thumb(),
		Actors:   cachingFilm.Actors,
		Release:  cachingFilm.ReleaseTime,
		Tags:     cachingFilm.Tags,
	})

	return cachingFilm, err

}

func (d *Javdb) updateExistFilm(existFilm *model.Film, actors, tags []string) {

	embyFile := virtual_file.ConvertFilmToEmbyFile(*existFilm, "")

	updateTagFlag := false
	updateActorTag := false

	existTagMap := make(map[string]bool)
	for _, tag := range embyFile.Tags {
		existTagMap[tag] = true
	}
	for _, tag := range tags {
		if !existTagMap[tag] {
			embyFile.Tags = append(embyFile.Tags, tag)
			updateTagFlag = true
		}
	}

	existActorMap := make(map[string]bool)
	for _, actor := range embyFile.Actors {
		existActorMap[actor] = true
	}
	for _, actor := range actors {
		if !existActorMap[actor] {
			embyFile.Actors = append(embyFile.Actors, actor)
			updateActorTag = true
		}
	}

	if !updateTagFlag && !updateActorTag {
		return
	}

	virtual_file.UpdateNfo(virtual_file.MediaInfo{
		Source:   DriverName,
		Dir:      embyFile.Path,
		FileName: virtual_file.AppendImageName(embyFile.Name),
		Title:    embyFile.Title,
		Actors:   embyFile.Actors,
		Release:  embyFile.ReleaseTime,
		Tags:     embyFile.Tags,
	})

	existFilm.Tags = embyFile.Tags
	existFilm.Actors = embyFile.Actors

	err1 := db.UpdateFilm(*existFilm)
	if err1 != nil {
		utils.Log.Warnf("failed to update films:[%s], error message: %s", tags, err1.Error())
	}

}

func (d *Javdb) getMagnet(file model.Obj, reMatchFilmMeta bool) (string, error) {

	embyObj, ok := file.(*model.EmbyFileObj)
	if !ok {
		return "", errors.New("this film doesn't contains film info")
	}

	magnetCache := db.QueryMagnetCacheByName(DriverName, embyObj.GetName())
	if magnetCache.Magnet != "" && !reMatchFilmMeta {
		utils.Log.Infof("return the magnet link from the cache:%s", magnetCache.Magnet)
		return magnetCache.Magnet, nil
	}

	javdbMeta, err := av.GetMetaFromJavdb(embyObj.Url)
	if err != nil || len(javdbMeta.Magnets) == 0 {

		if reMatchFilmMeta {
			return "", err
		}

		utils.Log.Warnf("failed to get javdb magnet info: %v,error message: %v, using the suke magnet instead.", javdbMeta, err)
		sukeMeta, err2 := av.GetMetaFromSuke(embyObj.GetName())
		if err2 != nil {
			utils.Log.Warn("failed to get suke magnet info:", err2.Error())
			return "", err2
		} else {
			if len(sukeMeta.Magnets) > 0 {
				return sukeMeta.Magnets[0].GetMagnet(), nil
			}
		}
		return "", err
	}

	d.updateFilmMeta(javdbMeta, embyObj)

	if magnetCache.Magnet == "" {
		return d.cacheMagnet(javdbMeta, embyObj)
	} else {
		return magnetCache.Magnet, nil
	}

}

func (d *Javdb) cacheMagnet(javdbMeta av.Meta, embyObj *model.EmbyFileObj) (string, error) {

	magnet := ""
	subtitle := false
	if javdbMeta.Magnets[0].IsSubTitle() {
		magnet = javdbMeta.Magnets[0].GetMagnet()
		subtitle = true
	} else {
		sukeMeta, err2 := av.GetMetaFromSuke(embyObj.GetName())
		if err2 != nil {
			utils.Log.Warn("failed to get suke magnet info:", err2.Error())
		} else {
			if len(sukeMeta.Magnets) > 0 {
				magnet = sukeMeta.Magnets[0].GetMagnet()
				subtitle = true
			}
		}

	}

	if magnet == "" {
		magnet = javdbMeta.Magnets[0].GetMagnet()
		subtitle = javdbMeta.Magnets[0].IsSubTitle()
	}

	err := db.CreateMagnetCache(model.MagnetCache{
		DriverType: DriverName,
		Magnet:     magnet,
		Name:       embyObj.GetName(),
		Subtitle:   subtitle,
		Code:       av.GetFilmCode(embyObj.GetName()),
		ScanAt:     time.Now(),
	})
	return magnet, err
}

func (d *Javdb) updateFilmMeta(javdbMeta av.Meta, embyObj *model.EmbyFileObj) {

	actorMapping := make(map[string]string)
	for _, actor := range javdbMeta.Actors {
		actorMapping[actor.Id] = actor.Name
	}
	for _, actor := range embyObj.Actors {
		actorMapping[actor] = actor
	}

	actors := db.QueryActor(strconv.Itoa(int(d.ID)))
	for _, actor := range actors {
		if actorMapping[actor.Url] != "" {
			actorMapping[actor.Url] = actor.Name
		}
	}

	var actorNames []string
	for _, name := range actorMapping {
		actorNames = append(actorNames, name)
	}

	var tags []string
	tagMap := make(map[string]bool)

	for _, tag := range embyObj.Tags {
		tagMap[tag] = true
	}

	if len(javdbMeta.Magnets) > 0 {
		for _, tag := range javdbMeta.Magnets[0].GetTags() {
			tagMap[tag] = true
		}
	}

	for tag, _ := range tagMap {
		tags = append(tags, tag)
	}

	virtual_file.UpdateNfo(virtual_file.MediaInfo{
		Source:   DriverName,
		Dir:      embyObj.Path,
		FileName: virtual_file.AppendImageName(embyObj.Name),
		Title:    embyObj.Title,
		Actors:   actorNames,
		Release:  embyObj.ReleaseTime,
		Tags:     tags,
	})

	tempId, err1 := strconv.ParseInt(embyObj.ID, 10, 64)
	if err1 == nil {
		err1 = db.UpdateFilm(model.Film{
			ID:     uint(tempId),
			Actors: actorNames,
			Tags:   tags,
		})
		if err1 != nil {
			utils.Log.Warnf("failed to save film: %s, error message: %s", embyObj.GetName(), err1.Error())
		}
	} else {
		utils.Log.Warnf("failed to parse films: %s id to int, error message: %s", embyObj.GetName(), err1.Error())
	}

}

func (d *Javdb) deleteFilm(dir, fileName, id string) error {
	err := db.DeleteAllMagnetCacheByCode(av.GetFilmCode(fileName))
	if err != nil {
		utils.Log.Warnf("failed to delete film cache:[%s], error message:[%s]", fileName, err.Error())
	}

	err = db.DeleteFilmById(id)
	if err != nil {
		utils.Log.Infof("failed to delete film:[%s], error message:[%s]", fileName, err.Error())
		return err
	}

	err = virtual_file.DeleteImageAndNfo(DriverName, dir, fileName)
	if err != nil {
		utils.Log.Infof("failed to delete film nfo:[%s], error message:[%s]", fileName, err)
		return err
	}
	return nil
}

func (d *Javdb) tryAcquireLink(ctx context.Context, file model.Obj, args model.LinkArgs, magnetGetter func(obj model.Obj) (string, error)) (*model.Link, error) {

	link, err := tool.CloudPlay(ctx, args, d.CloudPlayDriverType, d.CloudPlayDownloadPath, file, magnetGetter)

	if err != nil {
		utils.Log.Infof("The first cloud drive download failed:[%s]", err.Error())
		if d.BackPlayDriverType != "" {
			utils.Log.Infof("using the second cloud drive instead.")
			return tool.CloudPlay(ctx, args, d.BackPlayDriverType, d.CloudPlayDownloadPath, file, magnetGetter)
		}
	}

	return link, err
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

func splitName(sourceName string) (string, string) {

	index := strings.Index(sourceName, " ")
	if index <= 0 || index == len(sourceName)-1 {
		return sourceName, sourceName
	}

	return sourceName[:index], sourceName[index+1:]

}

func splitCode(sourceName string) string {

	code, _ := splitName(sourceName)
	return code

}
