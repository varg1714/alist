package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func CreateFilms(source string, actor string, models []model.ObjThumb) error {

	if len(models) == 0 {
		return nil
	}

	films := make([]model.Film, 0)

	for _, obj := range models {
		films = append(films, model.Film{
			Url:    obj.GetID(),
			Name:   obj.GetName(),
			Image:  obj.Thumb(),
			Source: source,
			Actor:  actor,
		})
	}

	return errors.WithStack(db.CreateInBatches(films, 100).Error)

}

func QueryByActor(source string, actor string) []model.Film {

	films := make([]model.Film, 0)
	film := model.Film{
		Source: source,
		Actor:  actor,
	}

	db.Where(film).Find(&films)

	return films

}

func QueryByUrls(actor string, urls []string) []string {

	films := make([]model.Film, 0)
	db.Select("url").Where("url IN (?)", urls).Where("actor = ?", actor).Find(&films)

	result := make([]string, 0)

	for _, film := range films {
		result = append(result, film.Url)
	}

	return result

}

func DeleteByActor(source string, actor string) error {

	return errors.WithStack(db.Where("source = ?", source).Where("actor = ?", actor).Delete(&model.Film{}).Error)

}

func QueryFileId(magnet string) string {

	magnetCache := model.MagnetCache{
		Magnet: magnet,
	}
	db.Where(magnetCache).First(&magnetCache)

	return magnetCache.FileId

}

func CreateCacheFile(magnet string, fileId string) error {

	magnetCache := model.MagnetCache{
		Magnet: magnet,
		FileId: fileId,
	}

	return errors.WithStack(db.Create(magnetCache).Error)

}

func UpdateCacheFile(magnet string, fileId string) error {

	magnetCache := model.MagnetCache{
		Magnet: magnet,
		FileId: fileId,
	}

	return errors.WithStack(db.Where("magnet = ?", magnet).Save(magnetCache).Error)

}
