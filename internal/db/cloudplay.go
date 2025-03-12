package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func CreateMagnetCache(magnetCache model.MagnetCache) error {

	if magnetCache.Code == "" {
		magnetCache.Code = GetFilmCode(magnetCache.Name)
	}

	err := DeleteCacheByName(magnetCache.DriverType, []string{magnetCache.Name})
	if err != nil {
		return err
	}
	return errors.WithStack(db.Create(&magnetCache).Error)
}

func BatchCreateMagnetCache(magnetCaches []model.MagnetCache) error {

	var names []string
	for index := range magnetCaches {
		if magnetCaches[index].Code == "" {
			magnetCaches[index].Code = GetFilmCode(magnetCaches[index].Name)
		}
		names = append(names, magnetCaches[index].Name)
	}

	err := DeleteCacheByName(magnetCaches[0].DriverType, names)
	if err != nil {
		return err
	}

	return errors.WithStack(db.CreateInBatches(&magnetCaches, 100).Error)

}

func QueryMagnetCacheByName(driverType, name string) model.MagnetCache {

	fileCache := model.MagnetCache{
		Name:       name,
		DriverType: driverType,
	}

	db.Where(fileCache).First(&fileCache)

	return fileCache

}

func QueryMagnetCacheByCode(code string) model.MagnetCache {

	code = GetFilmCode(code)

	fileCache := model.MagnetCache{
		Code: code,
	}

	db.Where(fileCache).First(&fileCache)

	return fileCache

}

func DeleteAllMagnetCacheByCode(code string) error {

	fileCache := model.MagnetCache{
		Code: GetFilmCode(code),
	}

	return errors.WithStack(db.Where(fileCache).Delete(&fileCache).Error)

}

func DeleteCacheByName(driveType string, names []string) error {

	return errors.WithStack(db.Where("driver_type = ?", driveType).Where("name in ?", names).Delete(&model.MagnetCache{}).Error)

}
