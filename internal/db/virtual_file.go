package db

import (
	"github.com/alist-org/alist/v3/internal/model"
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
