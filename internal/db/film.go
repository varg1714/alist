package db

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/pkg/errors"
	"strings"
)

func CreateFilms(source string, actor, actorId string, models []model.EmbyFileObj) error {

	if len(models) == 0 {
		return nil
	}

	films := make([]model.Film, 0)

	for _, obj := range models {
		films = append(films, model.Film{
			Url:       obj.Url,
			Name:      obj.GetName(),
			Image:     obj.Thumb(),
			Source:    source,
			Actor:     actor,
			ActorId:   actorId,
			CreatedAt: obj.Modified,
			Date:      obj.ReleaseTime,
			Title:     obj.Title,
			Actors:    obj.Actors,
			Tags:      obj.Tags,
		})
	}

	return errors.WithStack(db.CreateInBatches(&films, 100).Error)

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

func QueryFilmByCode(source string, code string) (model.Film, error) {

	var film model.Film

	tx := db.Where("source = ? ", source).Where("name like ?", fmt.Sprintf("%s%%", code)).First(&film)

	return film, tx.Error

}

func QueryIncompleteFilms(source string) ([]model.Film, error) {

	films := make([]model.Film, 0)
	err := db.Where("source = ?", source).Where(`(date is null or (title is null or title = "") or (actors is null))`).Find(&films).Error

	return films, err

}

func UpdateFilm(film model.Film) error {
	return db.Model(&film).Updates(model.Film{
		Date:   film.Date,
		Title:  film.Title,
		Actors: film.Actors,
		Tags:   film.Tags,
	}).Error
}

func QueryByUrls(actor string, urls []string) []string {

	films := make([]model.Film, 0)
	db.Select("url").Where("url IN (?)", urls).Where("actor_id = ?", actor).Find(&films)

	result := make([]string, 0)

	for _, film := range films {
		result = append(result, film.Url)
	}

	return result

}

func QueryFilmsByNamePrefix(source string, prefixes []string) ([]model.Film, error) {

	var films []model.Film

	nameQuery := db.Where("name like ?", prefixes[0]+"%")
	for _, p := range prefixes[1:] {
		nameQuery = nameQuery.Or("name like ?", p+"%")
	}

	query := db.Where(db.Where("source = ?", source)).Where(db.Where(nameQuery))

	return films, query.Find(&films).Error

}

func DeleteFilmsByActor(source string, actor string) error {

	return errors.WithStack(db.Where("source = ?", source).Where("actor = ?", actor).Delete(&model.Film{}).Error)

}

func DeleteFilmsByUrl(source, actor string, urls []string) error {

	return errors.WithStack(db.Where("source = ?", source).Where("actor = ?", actor).Where("url in ?", urls).Delete(&model.Film{}).Error)

}

func DeleteFilmById(id string) error {
	return errors.WithStack(db.Delete(&model.Film{}, id).Error)
}

func DeleteFilmsByCode(source, actor, code string) error {

	if code == "" {
		return nil
	}

	return errors.WithStack(db.Where("source = ?", source).
		Where("actor = ?", actor).
		Where("url like ?", fmt.Sprintf("%s%%", code)).
		Delete(&model.Film{}).Error)

}

func QueryUnSaveFilms(fileIds []string, dir string) []string {

	return queryNotExistDataWithCondition(fileIds, fmt.Sprintf("select url from x_films where actor = '%s'", dir))

}

func QueryNoMagnetFilms(fileIds []string) []string {

	return queryNotExistData(fileIds, "x_magnet_caches", "code")

}

func QueryUnMissedFilms(fileIds []string) []string {
	return queryNotExistData(fileIds, "x_missed_films", "code")
}

func CreateMissedFilms(fileIds []string) error {

	var missedFilms []model.MissedFilm
	for _, fileId := range fileIds {
		missedFilms = append(missedFilms, model.MissedFilm{
			Code: fileId,
		})
	}

	return errors.WithStack(db.CreateInBatches(&missedFilms, 100).Error)

}

func queryNotExistData(fileIds []string, dbName, columnName string) []string {
	return queryNotExistDataWithCondition(fileIds, fmt.Sprintf("select %s from %s", columnName, dbName))
}

func queryNotExistDataWithCondition(fileIds []string, sql string) []string {

	if len(fileIds) == 0 {
		return []string{}
	}

	var result []string
	var placeHolders []string
	var tempIds []any

	for _, fileId := range fileIds {
		placeHolders = append(placeHolders, "(?)")
		tempIds = append(tempIds, fileId)
	}

	query := fmt.Sprintf(`with temp(id) as (values %s)
select temp.id
from temp
where temp.id not in (%s);`, strings.Join(placeHolders, ","), sql)

	err := db.Raw(query, tempIds...).Scan(&result).Error
	if err != nil {
		utils.Log.Errorf("sql查询失败:%s", err.Error())
	}

	return result
}

func QueryNotMatchTagFilms(url []string, tag string) ([]model.Film, error) {

	var result []model.Film

	tx := db.Where("tags is null or tags not like ?", fmt.Sprintf("%%%s%%", tag))
	if len(url) > 0 {
		tx = tx.Where("url in ?", url)
	}

	find := tx.Find(&result)
	return result, errors.WithStack(find.Error)

}
