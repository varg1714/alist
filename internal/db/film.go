package db

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"time"
)

func CreateFilms(source string, actor, actorId string, models []model.EmbyFileObj) error {

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
			ActorId:   actorId,
			CreatedAt: now,
			Date:      obj.Modified,
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

func QueryFileCacheByName(name string) model.MagnetCache {

	fileCache := model.MagnetCache{
		Name: name,
	}

	db.Where(fileCache).First(&fileCache)

	return fileCache

}

func QueryFileCacheByCode(code string) model.MagnetCache {

	code = GetFilmCode(code)

	fileCache := model.MagnetCache{
		Code: code,
	}

	db.Where(fileCache).First(&fileCache)

	return fileCache

}

func CreateCacheFile(magnet string, fileId string, name string) error {

	code := GetFilmCode(name)

	magnetCache := model.MagnetCache{
		Magnet: magnet,
		FileId: fileId,
		Name:   name,
		Code:   code,
	}

	err := DeleteCacheByName(name)
	if err != nil {
		return err
	}

	return errors.WithStack(db.Create(&magnetCache).Error)

}

func GetFilmCode(name string) string {
	code := name
	split := strings.Split(name, " ")
	if len(split) >= 2 {
		code = split[0]
	} else {
		nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+)?.mp4")
		code = nameRegexp.ReplaceAllString(name, "$1")
	}
	return code
}

func UpdateCacheFile(magnet string, fileId string, name string) error {

	var code string
	split := strings.Split(name, " ")
	if len(split) >= 2 {
		code = split[0]
	}

	magnetCache := model.MagnetCache{
		Magnet: magnet,
		FileId: fileId,
		Name:   name,
		Code:   code,
	}

	return errors.WithStack(db.Where("code = ?", code).Save(&magnetCache).Error)

}

func DeleteCacheByCode(code string) error {

	fileCache := model.MagnetCache{
		Code: GetFilmCode(code),
	}

	return errors.WithStack(db.Where(fileCache).Delete(&fileCache).Error)

}

func DeleteCacheByName(name string) error {

	fileCache := model.MagnetCache{
		Name: name,
	}

	return errors.WithStack(db.Where(fileCache).Delete(&fileCache).Error)

}

func DeleteCacheFile(fileId string) error {

	magnetCache := model.MagnetCache{
		FileId: fileId,
	}

	return errors.WithStack(db.Where(magnetCache).Delete(&magnetCache).Error)

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
