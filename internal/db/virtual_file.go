package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func QueryVirtualFiles(storageId uint, parent string) []model.VirtualFile {

	names := make([]model.VirtualFile, 0)
	db.Where(&model.VirtualFile{
		StorageId: storageId,
		Parent:    parent,
	}).Order("modified DESC").Find(&names)

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
