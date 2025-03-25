package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"strings"
)

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
