package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"strings"
	"time"
)

func CreateFilms(source string, actor, actorId string, models []model.ObjThumb) error {

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

func QueryCacheFileId(name string) model.MagnetCache {

	var code string
	split := strings.Split(name, " ")
	if len(split) >= 2 {
		code = split[0]
	}

	fileCache := model.MagnetCache{
		Code: code,
	}
	db.Where(fileCache).First(&fileCache)

	return fileCache

}

func CreateCacheFile(magnet string, fileId string, name string) error {

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

	return errors.WithStack(db.Create(&magnetCache).Error)

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
