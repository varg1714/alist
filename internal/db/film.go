package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"time"
)

func CreateFilms(source string, actor string, models []model.ObjThumb) error {

	if len(models) == 0 {
		return nil
	}

	films := make([]model.Film, 0)

	now := time.Now()
	for _, obj := range models {
		now = now.Add(-1 * time.Hour)
		films = append(films, model.Film{
			Url:       obj.GetID(),
			Name:      obj.GetName(),
			Image:     obj.Thumb(),
			Source:    source,
			Actor:     actor,
			CreatedAt: now,
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

	db.Where(film).Order("created_at desc").Find(&films)

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

func CreateActor(dir string, name string, url string) error {

	actor := model.Actor{
		Dir:  dir,
		Name: name,
		Url:  url,
	}

	return errors.WithStack(db.Create(actor).Error)

}

func QueryActor(source string) []model.Actor {

	actors := make([]model.Actor, 0)
	actor := model.Actor{
		Dir: source,
	}

	db.Where(actor).Find(&actors)

	return actors

}

func DeleteActor(source string, actor string) error {

	return errors.WithStack(db.Where("dir = ?", source).Where("name = ?", actor).Delete(&model.Actor{}).Error)

}
