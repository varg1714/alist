package virtual_file

import (
	"encoding/xml"
	"fmt"
	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func GetFilms(source, dirName string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, data []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error)) ([]model.EmbyFileObj, error) {

	results := make([]model.EmbyFileObj, 0)
	films := make([]model.EmbyFileObj, 0)

	films, nextPage, err := pageFunc(urlFunc, 1, films)
	if err != nil {
		return results, err
	}

	// not exists
	for index := 2; index <= 20 && nextPage; index++ {

		films, nextPage, err = pageFunc(urlFunc, index, films)
		if err != nil {
			return results, err
		}

	}

	return convertObj(source, dirName, films, results), nil

}

func GetFilmsWithStorage(source, dirName, actorId string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, preFilms []model.EmbyFileObj) ([]model.EmbyFileObj, bool, error), option Option) ([]model.EmbyFileObj, error) {

	results := make([]model.EmbyFileObj, 0)
	films := make([]model.EmbyFileObj, 0)

	films, nextPage, err := pageFunc(urlFunc, 1, films)
	if err != nil {
		return convertFilm(source, dirName, db.QueryByActor(source, dirName), results, option.CacheFile), err
	}

	var urls []string
	for _, item := range films {
		urls = append(urls, item.ID)
	}

	existFilms := db.QueryByUrls(actorId, urls)

	// not exists
	for index := 2; index <= option.MaxPageNum && nextPage && len(existFilms) == 0; index++ {

		films, nextPage, err = pageFunc(urlFunc, index, films)
		if err != nil {
			return convertFilm(source, dirName, db.QueryByActor(source, dirName), results, option.CacheFile), err
		}
		clear(urls)
		for _, item := range films {
			urls = append(urls, item.ID)
		}

		existFilms = db.QueryByUrls(actorId, urls)

	}
	// exist
	for index, item := range films {
		if utils.SliceContains(existFilms, item.ID) {
			if index == 0 {
				films = []model.EmbyFileObj{}
			} else {
				films = films[:index]
			}
			break
		}
	}

	if len(films) != 0 {
		err = db.CreateFilms(source, dirName, actorId, films)
		if err != nil {
			return convertFilm(source, dirName, db.QueryByActor(source, dirName), results, option.CacheFile), nil
		}
	}

	return convertFilm(source, dirName, db.QueryByActor(source, dirName), results, option.CacheFile), nil

}

func GetStorageFilms(source, dirName string, cacheFile bool) []model.EmbyFileObj {
	return convertFilm(source, dirName, db.QueryByActor(source, dirName), []model.EmbyFileObj{}, cacheFile)
}

func convertFilm(source, dirName string, films []model.Film, results []model.EmbyFileObj, cacheFile bool) []model.EmbyFileObj {

	for _, film := range films {

		thumb := model.EmbyFileObj{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					IsFolder: false,
					ID:       film.Url,
					Size:     1417381701,
					Modified: film.Date,
					Path:     dirName,
				},
				Thumbnail: model.Thumbnail{Thumbnail: film.Image},
			},
			Title:  ClearFilmName(film.Name),
			Actors: film.Actors,
		}

		if strings.HasSuffix(film.Name, "mp4") {
			thumb.Name = AppendFilmName(CutString(ClearFilmName(film.Name)))
		} else {
			thumb.Name = AppendFilmName(CutString(film.Name))
		}

		if cacheFile {
			_ = CacheImageAndNfo(source, dirName, AppendImageName(thumb.Name), thumb.Title, film.Image, []string{dirName})
		}

		results = append(results, thumb)
	}
	return results
}

func convertObj(source, dirName string, actor []model.EmbyFileObj, results []model.EmbyFileObj) []model.EmbyFileObj {

	for _, film := range actor {
		parse, _ := time.Parse(time.DateTime, "2024-01-02 15:04:05")
		results = append(results, model.EmbyFileObj{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					Name:     AppendFilmName(film.Name),
					IsFolder: false,
					ID:       film.ID,
					Size:     1417381701,
					Modified: parse,
					Path:     dirName,
				},
				Thumbnail: model.Thumbnail{Thumbnail: film.Thumb()},
			},
			Title: film.Name,
		})

		_ = CacheImageAndNfo(source, dirName, AppendImageName(film.Name), film.Name, film.Thumb(), []string{dirName})

	}
	return results

}

func CacheImageAndNfo(source, dir, fileName, title, img string, actors []string) int {

	actorNfo := cacheActorNfo(dir, fileName, title, actors, source)
	if actorNfo == Exist {
		return Exist
	}

	return CacheImage(source, dir, fileName, img, map[string]string{})

}

func DeleteImageAndNfo(source, dir, fileName string) error {

	sourceName := fileName[0:strings.LastIndex(fileName, ".")]

	nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+)")

	if nameRegexp.MatchString(sourceName) {

		// 有多个文件
		sourceName = nameRegexp.ReplaceAllString(sourceName, "$1")

		filePath := filepath.Join(flags.DataDir, "emby", source, dir, fmt.Sprintf("%s-cd1.nfo", sourceName))
		for i := 1; utils.Exists(filePath); {
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
			i++
			filePath = filepath.Join(flags.DataDir, "emby", source, dir, fmt.Sprintf("%s-cd%d.nfo", sourceName, i))
		}

		filePath = filepath.Join(flags.DataDir, "emby", source, dir, fmt.Sprintf("%s-cd1.jpg", sourceName))
		for i := 1; utils.Exists(filePath); {
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
			i++
			filePath = filepath.Join(flags.DataDir, "emby", source, dir, fmt.Sprintf("%s-cd%d.jpg", sourceName, i))
		}

	} else {
		// 删除nfo文件
		filePath := filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".nfo")
		if utils.Exists(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
		}

		// 删除img文件
		filePath = filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".jpg")
		if utils.Exists(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func CacheImage(source string, dir string, fileName string, img string, requestHeader map[string]string) int {
	if img == "" {
		return CreatedFailed
	}

	filePath := filepath.Join(flags.DataDir, "emby", source, dir, fileName)
	if utils.Exists(filePath) {
		return Exist
	}

	imgResp, err := base.RestyClient.R().SetHeaders(requestHeader).Get(img)
	if err != nil {
		utils.Log.Info("图片下载失败", err)
		return CreatedFailed
	}

	err = os.MkdirAll(filepath.Join(flags.DataDir, "emby", source, dir), 0777)
	if err != nil {
		utils.Log.Info("图片缓存文件夹创建失败", err)
		return CreatedFailed
	}

	err = os.WriteFile(filePath, imgResp.Body(), 0777)
	if err != nil {
		utils.Log.Info("图片缓存失败", err)
		return CreatedFailed
	}

	return CreatedSuccess
}

func UpdateNfo(source, dir, fileName, title string, actors []string) {

	cacheResult := cacheActorNfo(dir, fileName, title, actors, source)

	if cacheResult == Exist {
		sourceName := fileName[0:strings.LastIndex(fileName, ".")]
		filePath := filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".nfo")

		file, err := os.ReadFile(filePath)
		if err != nil {
			utils.Log.Warnf("nfo文件读取失败:%s", err.Error())
			return
		}

		var media Media

		err = xml.Unmarshal(file, &media)
		if err != nil {
			utils.Log.Warnf("nfo文件转换失败:%s", err.Error())
			return
		}
		var actorInfos []Actor
		for _, actor := range actors {
			actorInfos = append(actorInfos, Actor{
				Name: actor,
			})
		}

		media.Actor = actorInfos

		mediaXml, err := mediaToXML(&media)
		if err != nil {
			utils.Log.Info("xml格式转换失败", err)
			return
		}
		err = os.WriteFile(filePath, mediaXml, 0777)
		if err != nil {
			utils.Log.Infof("文件:%s的xml更新失败:%v", fileName, err)
		}

	}

}

func cacheActorNfo(dir, fileName, title string, actors []string, source string) int {

	if fileName == "" {
		return CreatedFailed
	}

	sourceName := fileName[0:strings.LastIndex(fileName, ".")]

	filePath := filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".nfo")
	if utils.Exists(filePath) {
		return Exist
	}

	err := os.MkdirAll(filepath.Join(flags.DataDir, "emby", source, dir), 0777)
	if err != nil {
		utils.Log.Info("nfo缓存文件夹创建失败", err)
		return CreatedFailed
	}

	var actorInfos []Actor
	for _, actor := range actors {
		actorInfos = append(actorInfos, Actor{
			Name: actor,
		})
	}

	media := Media{
		Plot:  Inner{Inner: fmt.Sprintf("<![CDATA[%s]]>", title)},
		Title: Inner{Inner: title},
		Actor: actorInfos,
	}

	xml, err := mediaToXML(&media)
	if err != nil {
		utils.Log.Info("xml格式转换失败", err)
		return CreatedFailed
	}
	err = os.WriteFile(filePath, xml, 0777)
	if err != nil {
		utils.Log.Infof("文件:%s的xml缓存失败:%v", fileName, err)
	}

	return CreatedSuccess

}

func CutString(name string) string {

	prettyNameRegexp, _ := regexp.Compile("[\\/\\\\\\*\\?\\:\\\"\\<\\>\\|]")
	name = prettyNameRegexp.ReplaceAllString(name, "")

	// 将字符串转换为 rune 切片
	runes := []rune(name)

	if len(runes) <= 70 {
		return name
	}

	// 检查长度并截取
	runes = runes[:70]

	// 将 rune 切片转换回字符串
	return string(runes)

}

func ClearFilmName(name string) string {

	if strings.HasSuffix(name, ".mp4") {
		return name[0 : len(name)-4]
	}

	if strings.HasSuffix(name, ".jpg") {
		return name[0 : len(name)-4]
	}

	if strings.HasSuffix(name, ".") {
		// 仅有.
		return name[0 : len(name)-1]
	}

	// 返回原始文件名
	return name
}

func AppendFilmName(name string) string {
	// 返回原始文件名
	return ClearFilmName(name) + ".mp4"

}

func AppendImageName(name string) string {
	return ClearFilmName(name) + ".jpg"
}
