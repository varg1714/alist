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

func QueryNoDateFilms(source string) ([]model.Film, error) {

	films := make([]model.Film, 0)
	err := db.Where("source = ?", source).Where("date is null").Find(&films).Error

	return films, err

}

func UpdateFilmDate(film model.Film) error {
	return db.Model(&film).
		Where("source = ?", film.Source).
		Where("name = ?", film.Name).
		Update("date", film.Date).Error
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

func DeleteFilmsByActor(source string, actor string) error {

	return errors.WithStack(db.Where("source = ?", source).Where("actor = ?", actor).Delete(&model.Film{}).Error)

}

func DeleteFilmsByUrl(source, actor, url string) error {

	return errors.WithStack(db.Where("source = ?", source).Where("actor = ?", actor).Where("url = ?", url).Delete(&model.Film{}).Error)

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

func CreateActor(dir string, name string, url string) error {

	actor := model.Actor{
		Dir:  dir,
		Name: name,
		Url:  url,
	}

	err := db.Create(&actor).Error
	return errors.WithStack(err)

}

func QueryActor(source string) []model.Actor {

	actors := make([]model.Actor, 0)
	actor := model.Actor{
		Dir: source,
	}

	db.Where(actor).Order("updated_at desc").Find(&actors)

	return actors

}

func DeleteActor(source string, actor string) error {

	return errors.WithStack(db.Where("dir = ?", source).Where("name = ?", actor).Delete(&model.Actor{}).Error)

}

func QueryVirtualFiles(storageId string) []model.VirtualFile {

	names := make([]model.VirtualFile, 0)
	db.Where("storage_id = ?", storageId).Order("modified DESC").Find(&names)

	return names

}

func QueryVirtualFilm(storageId uint, name string) model.VirtualFile {

	file := model.VirtualFile{
		StorageId: storageId,
		Name:      name,
	}

	db.Where(file).Take(&file)

	return file

}

func CreateVirtualFile(virtualFile model.VirtualFile) error {

	return errors.WithStack(db.Create(&virtualFile).Error)

}

func DeleteVirtualFile(storageId uint, obj model.Obj) error {

	virtualFile := model.VirtualFile{}
	db.Where("storage_id = ?", storageId).Where("name = ?", obj.GetName()).Take(&virtualFile)
	if virtualFile.ShareID != "" {
		// virtual share file
		return errors.WithStack(db.Where("storage_id = ?", storageId).Where("name = ?", obj.GetName()).Delete(&model.VirtualFile{}).Error)
	} else {
		// delete file
		replacement := model.Replacement{
			StorageId: storageId,
			DirName: func() string {
				virtualFile = QueryVirtualFilm(storageId, strings.Split(obj.GetPath(), "/")[0])
				return virtualFile.ShareID
			}(),
			Type:    1,
			OldName: obj.GetID(),
		}
		return errors.WithStack(db.Create(&replacement).Error)
	}

}

func Rename(storageId uint, dir, oldName, newName string) error {

	replacement := model.Replacement{
		StorageId: storageId,
		DirName:   dir,
		OldName:   oldName,
	}
	db.Where("storage_id = ?", storageId).Where("dir_name = ?", dir).Where("old_name = ?", oldName).Take(&replacement)

	if replacement.NewName != "" {
		replacement.NewName = newName
		return db.Where("storage_id = ? and dir_name = ? and old_name = ?", storageId, dir, oldName).Save(replacement).Error
	} else {
		replacement.NewName = newName
		return errors.WithStack(db.Create(&replacement).Error)
	}

}

func QueryReplacements(storageId uint, dir string) []model.Replacement {

	param := model.Replacement{
		StorageId: storageId,
		DirName:   dir,
	}

	result := make([]model.Replacement, 0)
	db.Where(param).Find(&result)

	return result

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
