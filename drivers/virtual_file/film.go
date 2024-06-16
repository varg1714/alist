package virtual_file

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"strings"
)

func GetFilms(dirName string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, data []model.ObjThumb) ([]model.ObjThumb, bool, error)) ([]model.ObjThumb, error) {

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

	return convertObj(dirName, films, results), nil

}

func GetFilmsWitchStorage(source, dirName string, urlFunc func(index int) string, pageFunc func(urlFunc func(index int) string, index int, preFilms []model.ObjThumb) ([]model.ObjThumb, bool, error)) ([]model.ObjThumb, error) {

	results := make([]model.ObjThumb, 0)
	films := make([]model.ObjThumb, 0)

	films, nextPage, err := pageFunc(urlFunc, 1, films)
	if err != nil {
		return convertFilm(dirName, db.QueryByActor(source, dirName), results), err
	}

	var urls []string
	for _, item := range films {
		urls = append(urls, item.ID)
	}

	existFilms := db.QueryByUrls(dirName, urls)

	// not exists
	for index := 2; index <= 20 && nextPage && len(existFilms) == 0; index++ {

		films, nextPage, err = pageFunc(urlFunc, index, films)
		if err != nil {
			return convertFilm(dirName, db.QueryByActor(source, dirName), results), err
		}
		clear(urls)
		for _, item := range films {
			urls = append(urls, item.ID)
		}

		existFilms = db.QueryByUrls(dirName, urls)

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
		err = db.CreateFilms(source, dirName, films)
		if err != nil {
			return convertFilm(dirName, db.QueryByActor(source, dirName), results), nil
		}
	}

	return convertFilm(dirName, db.QueryByActor(source, dirName), results), nil

}

func GeoStorageFilms(source, dirName string) []model.ObjThumb {
	films := db.QueryByActor(source, dirName)
	return convertFilm(dirName, films, []model.ObjThumb{})
}

func convertFilm(dirName string, actor []model.Film, results []model.ObjThumb) []model.ObjThumb {
	for _, film := range actor {
		thumb := model.ObjThumb{
			Object: model.Object{
				IsFolder: false,
				ID:       film.Url,
				Size:     622857143,
				Modified: film.Date,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Image},
		}

		if strings.HasSuffix(film.Name, "mp4") {
			thumb.Name = film.Name
		} else {
			thumb.Name = film.Name + ".mp4"
		}
		results = append(results, thumb)
	}
	return results
}

func convertObj(dirName string, actor []model.ObjThumb, results []model.ObjThumb) []model.ObjThumb {
	for _, film := range actor {
		results = append(results, model.ObjThumb{
			Object: model.Object{
				Name:     film.Name + ".mp4",
				IsFolder: false,
				ID:       film.ID,
				Size:     622857143,
				Modified: film.Modified,
				Path:     dirName,
			},
			Thumbnail: model.Thumbnail{Thumbnail: film.Thumb()},
		})
	}
	return results
}
