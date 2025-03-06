package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func CreateMagnetCache(magnetCache model.MagnetCache) error {

	if magnetCache.Code == "" {
		magnetCache.Code = GetFilmCode(magnetCache.Name)
	}

	err := DeleteCacheByName(magnetCache.DriverType, magnetCache.Name)
	if err != nil {
		return err
	}
	return errors.WithStack(db.Create(&magnetCache).Error)
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

func DeleteCacheByName(driveType, name string) error {

	fileCache := model.MagnetCache{
		Name:       name,
		DriverType: driveType,
	}

	return errors.WithStack(db.Where(fileCache).Delete(&fileCache).Error)

}
