package virtual_file

import (
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

func GetFilms(source, dirName string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error)) ([]model.ObjThumb, error) {

	results := make([]model.ObjThumb, 0)
	films := make([]model.ObjThumb, 0)

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

func GetFilmsWitchStorage(source, dirName, actorId string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, preFilms []model.ObjThumb) ([]model.ObjThumb, bool, error), option Option) ([]model.ObjThumb, error) {

	results := make([]model.ObjThumb, 0)
	films := make([]model.ObjThumb, 0)

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
				films = []model.ObjThumb{}
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

func GeoStorageFilms(source, dirName string, cacheFile bool) []model.ObjThumb {
	films := db.QueryByActor(source, dirName)
	return convertFilm(source, dirName, films, []model.ObjThumb{}, cacheFile)
}

func convertFilm(source, dirName string, actor []model.Film, results []model.ObjThumb, cacheFile bool) []model.ObjThumb {
	for _, film := range actor {

		thumb := model.ObjThumb{
			Object: model.Object{
				IsFolder: false,
				ID:       film.Url,
				Size:     1417381701,
				Modified: film.Date,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		}

		sourceName := film.Name

		if strings.HasSuffix(film.Name, "mp4") {
			thumb.Name = film.Name
			strings.LastIndex(film.Name, ".")
			sourceName = film.Name[0:strings.LastIndex(film.Name, ".")]
		} else {
			thumb.Name = AppendFilmName(film.Name)
		}

		if cacheFile {
			_ = CacheImage(source, dirName, AppendImageName(sourceName), film.Image)
		}

		results = append(results, thumb)
	}
	return results
}

func convertObj(source, dirName string, actor []model.ObjThumb, results []model.ObjThumb) []model.ObjThumb {

	for _, film := range actor {
		parse, _ := time.Parse(time.DateTime, "2024-01-02 15:04:05")
		results = append(results, model.ObjThumb{
			Object: model.Object{
				Name:     AppendFilmName(film.Name),
				IsFolder: false,
				ID:       film.ID,
				Size:     1417381701,
				Modified: parse,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Thumb()},
		})

		_ = CacheImage(source, dirName, AppendImageName(film.Name), film.Thumb())

	}
	return results

}

func CacheImage(source, dir, name, img string) int {

	actorNfo := cacheActorNfo(dir, name, source)
	if actorNfo == Exist {
		return Exist
	}

	if img == "" {
		return CreatedFailed
	}

	if utils.Exists(filepath.Join(flags.DataDir, "emby", source, dir, name)) {
		return Exist
	}

	imgResp, err := base.RestyClient.R().Get(img)
	if err != nil {
		utils.Log.Info("图片下载失败", err)
		return CreatedFailed
	}

	err = os.MkdirAll(filepath.Join(flags.DataDir, "emby", source, dir), 0777)
	if err != nil {
		utils.Log.Info("图片缓存文件夹创建失败", err)
		return CreatedFailed
	}

	err = os.WriteFile(filepath.Join(flags.DataDir, "emby", source, dir, name), imgResp.Body(), 0777)
	if err != nil {
		utils.Log.Info("图片缓存失败", err)
		return CreatedFailed
	}

	return CreatedSuccess

}

func cacheActorNfo(dir, name, source string) int {

	if name == "" {
		return CreatedFailed
	}

	sourceName := name[0:strings.LastIndex(name, ".")]

	if utils.Exists(filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".nfo")) {
		return Exist
	}

	err := os.MkdirAll(filepath.Join(flags.DataDir, "emby", source, dir), 0777)
	if err != nil {
		utils.Log.Info("nfo缓存文件夹创建失败", err)
		return CreatedFailed
	}

	media := Media{
		Plot:  Inner{Inner: fmt.Sprintf("<![CDATA[%s]]>", sourceName)},
		Title: Inner{Inner: sourceName},
		Actor: []Actor{
			{
				Name: dir,
			},
		},
	}

	xml, err := mediaToXML(&media)
	if err != nil {
		utils.Log.Info("xml格式转换失败", err)
		return CreatedFailed
	}
	err = os.WriteFile(filepath.Join(flags.DataDir, "emby", source, dir, sourceName+".nfo"), xml, 0777)
	if err != nil {
		utils.Log.Infof("文件:%s的xml缓存失败:%v", name, err)
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

func AppendFilmName(name string) string {

	if strings.HasSuffix(name, ".mp4") {
		return name
	}

	if strings.HasSuffix(name, ".") {
		// 仅有.
		return name + "mp4"
	}

	// 返回原始文件名
	return name + ".mp4"

}

func ClearFilmName(name string) string {

	if strings.HasSuffix(name, ".mp4") {
		return name[0 : len(name)-4]
	}

	if strings.HasSuffix(name, ".") {
		// 仅有.
		return name[0 : len(name)-1]
	}

	// 返回原始文件名
	return name
}

func AppendImageName(name string) string {

	if strings.HasSuffix(name, ".jpg") {
		return name
	}

	if strings.HasSuffix(name, ".") {
		// 仅有.
		return name + "jpg"
	}

	// 是影片结尾的图片名
	if strings.HasSuffix(name, ".mp4") {
		return name[0:strings.LastIndex(name, ".")] + ".jpg"
	}

	// 返回原始文件名
	return name + ".jpg"

}
