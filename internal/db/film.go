package db

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

func CreateFilms(source string, actor, actorId string, models []model.EmbyFileObj) error {

	if len(models) == 0 {
		return nil
	}

	films := make([]model.Film, 0)

	for _, obj := range models {
		films = append(films, model.Film{
			Url:       obj.GetID(),
			Name:      obj.GetName(),
			Image:     obj.Thumb(),
			Source:    source,
			Actor:     actor,
			ActorId:   actorId,
			CreatedAt: obj.Modified,
			Date:      obj.ReleaseTime,
			Title:     obj.Title,
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
	return db.Model(&film).Updates(map[string]any{
		"date":   film.Date,
		"title":  film.Title,
		"actors": film.Actors,
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

func DeleteFilmsByPrefixUrl(source, actor, url string) error {

	return errors.WithStack(db.Where("source = ?", source).
		Where("actor = ?", actor).
		Where("url like ?", fmt.Sprintf("%s%%", GetFilmCode(url))).
		Delete(&model.Film{}).Error)

}

func GetFilmCode(name string) string {
	code := name
	split := strings.Split(name, " ")
	if len(split) >= 2 {
		code = split[0]
	} else {
		nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+)?.mp4")
		if nameRegexp.MatchString(name) {
			code = nameRegexp.ReplaceAllString(name, "$1")
		}
	}
	return code
}

func QueryUnCachedFilms(fileIds []string) []string {

	return queryNotExistData(fileIds, "x_magnet_caches")

}

func QueryUnMissedFilms(fileIds []string) []string {
	return queryNotExistData(fileIds, "x_missed_films")
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

func queryNotExistData(fileIds []string, dbName string) []string {

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
where temp.id not in (select code from %s);`, strings.Join(placeHolders, ","), dbName)

	err := db.Raw(query, tempIds...).Scan(&result).Error
	if err != nil {
		utils.Log.Errorf("sql查询失败:%s", err.Error())
	}

	return result
}
