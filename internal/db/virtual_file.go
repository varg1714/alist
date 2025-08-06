package db

import (
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

func QueryVirtualFiles(storageId uint, parent string) []model.VirtualFile {

	names := make([]model.VirtualFile, 0)

	tx := db.Where("storage_id = ?", storageId)
	if parent == "" {
		tx.Where("(parent is null or parent = '')")
	} else {
		tx.Where("parent = ?", parent)
	}

	tx.Order("modified DESC").Find(&names)

	return names

}

func QueryVirtualFilesById(storageId uint, ids []string) ([]model.VirtualFile, error) {
	var files []model.VirtualFile
	tx := db.Where("id in ?", ids).Find(&files)
	return files, errors.WithStack(tx.Error)
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

func DeleteVirtualFile(virtualFile model.VirtualFile) error {

	return errors.WithStack(db.Delete(&virtualFile).Error)

}

func UpdateVirtualFile(virtualFile model.VirtualFile) error {
	return errors.WithStack(db.Updates(&virtualFile).Error)
}
