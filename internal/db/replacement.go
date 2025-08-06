package db

import (
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

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

func CreateReplacement(replacement model.Replacement) error {
	return errors.WithStack(db.Create(&replacement).Error)
}
