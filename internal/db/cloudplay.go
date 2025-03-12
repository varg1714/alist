package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
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

func QueryNoSubtitlesCache(driverType string) ([]model.MagnetCache, error) {

	var caches []model.MagnetCache

	err := errors.WithStack(
		db.Where("(scan_at is null or scan_at <= ?)", time.Now().AddDate(0, 0, -3)).
			Where("(scan_count is null or scan_count < 10)").
			Where("(subtitle is null or subtitle = 0)").
			Where("driver_type = ?", driverType).
			Find(&caches).Error)

	return caches, err

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

func UpdateScanData(driveType string, names []string, scanAt time.Time) error {
	return errors.WithStack(db.Model(&model.MagnetCache{}).Where("driver_type = ?", driveType).Where("name in ?", names).
		Updates(map[string]any{
			"scan_at":    scanAt,
			"scan_count": gorm.Expr("ifnull(scan_count, 0) + ?", 1),
		}).Error)
}
