package tool

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	driver2 "github.com/SheltonZhu/115driver/pkg/driver"
	_115 "github.com/alist-org/alist/v3/drivers/115"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"math"
	"regexp"
	"slices"
	"time"
)

func CloudPlay(ctx context.Context, args model.LinkArgs, driverType, driverPath string, downloadingFile model.Obj, magnetGetter func(obj model.Obj) (string, error)) (*model.Link, error) {

	if driverPath == "" {
		switch driverType {
		case "115 Cloud":
			driverPath = setting.GetStr(conf.Pan115TempDir)
		case "PikPak":
			driverPath = setting.GetStr(conf.PikPakTempDir)
		}
	}

	if driverPath == "" {
		return nil, errors.New("尚未配置用于云播的网盘")
	}

	storage := op.GetBalancedStorage(driverPath)
	if storage == nil {
		return nil, errors.New("网盘配置未找到")
	}

	// 1. 尝试获取缓存文件
	fileName := downloadingFile.GetName()

	// 1.1 获取缓存的文件ID
	fileCache := db.QueryMagnetCacheByName(driverType, fileName)

	// 1.2 缓存文件不为空，返回该文件
	if fileCache.FileId != "" {
		cache, err := getLinkByCache(ctx, args, driverType, storage, fileCache)
		if err != nil {
			return cache, err
		} else if cache != nil {
			return cache, nil
		}
	}

	// 2. 获取磁力链接
	start := time.Now().UnixMilli()
	magnet, err := magnetGetter(downloadingFile)
	if err != nil {
		utils.Log.Info("磁力链接获取失败", err)
	}
	utils.Log.Infof("获取:[%s]的磁力链接结果为:[%s]耗时:[%d]", downloadingFile.GetName(), magnet, time.Now().UnixMilli()-start)
	if magnet == "" {
		return nil, errors.New("磁力信息获取为空")
	}

	// 3. 下载文件
	status, _, err := downloadMagnet(ctx, driverType, driverPath, magnet, fileName)
	if err != nil {
		return nil, err
	}

	// 4. 解析下载结果
	downloadedFile := &model.ObjThumb{
		Object: model.Object{ID: status.FileInfo.FileId},
	}
	// 4.1 获取下载完毕的文件
	fileList, err2 := storage.List(ctx, downloadedFile, model.ListArgs{})
	if err2 != nil {
		return nil, err2
	}

	switch driverType {
	case "PikPak":
		if len(fileList) == 0 {
			downloadedFile.ID = status.FileInfo.FileId
			err1 := db.CreateMagnetCache(model.MagnetCache{
				DriverType: driverType,
				Magnet:     magnet,
				FileId:     status.FileInfo.FileId,
				Name:       fileName,
			})
			if err1 != nil {
				utils.Log.Infof("文件缓存失败:%s", err1.Error())
			}
			return storage.Link(ctx, downloadedFile, args)
		} else {
			lookedFile := cacheFiles(driverType, magnet, fileName, fileList, func(obj model.Obj) map[string]string {
				return nil
			})
			return storage.Link(ctx, lookedFile, args)
		}
	case "115 Cloud":
		lookedFile := cacheFiles(driverType, magnet, fileName, fileList, func(obj model.Obj) map[string]string {
			return map[string]string{
				"pickCode": obj.(*_115.FileObj).PickCode,
			}
		})

		return storage.Link(ctx, lookedFile, args)
	}

	return nil, nil

}

func downloadMagnet(ctx context.Context, driverType string, driverPath string, magnet string, fileName string) (*Status, *DownloadTask, error) {

	// 1. 下载该文件
	task, err := AddURL(ctx, &AddURLArgs{
		URL:          magnet,
		DstDirPath:   driverPath,
		Tool:         driverType,
		DeletePolicy: DeleteOnUploadSucceed,
	})
	if err != nil {
		return nil, nil, err
	} else if task.GetErr() != nil {
		return nil, nil, task.GetErr()
	}

	utils.Log.Infof("提交离线下载任务：%s", fileName)
	downloadTask := task.(*DownloadTask)

	i := 0
	completed := false
	status, err := downloadTask.tool.Status(downloadTask)

	for i < 30 && !completed {
		if err != nil {
			return nil, nil, err
		}

		utils.Log.Infof("当前任务下载进度：%f", func() float64 {
			if status == nil {
				return 0.0
			} else {
				return status.Progress
			}
		}())

		if status == nil || !(status.Completed || math.Dim(100.0, status.Progress) <= 0.01) {
			i++
			time.Sleep(2 * time.Second)
			status, err = downloadTask.tool.Status(downloadTask)
		} else {
			completed = true
		}

	}

	if status == nil || !completed {
		return nil, nil, errors.New("文件仍未下载完成")
	}

	return status, downloadTask, nil

}

func cacheFiles(driverType, magnet, lookingFileName string, files []model.Obj, cacheOptionFunc func(obj model.Obj) map[string]string) model.Obj {

	// 仅包含100M大小以上的文件
	validFiles := utils.SliceFilter(files, func(f model.Obj) bool {
		return f.GetSize()/(1024*1024) > 100
	})

	// 按名称正序
	slices.SortFunc(validFiles, func(a, b model.Obj) int {
		return cmp.Compare(a.GetName(), b.GetName())
	})

	if len(validFiles) == 0 {
		return nil
	} else if len(validFiles) == 1 {
		err := db.CreateMagnetCache(model.MagnetCache{
			DriverType: driverType,
			Magnet:     magnet,
			FileId:     validFiles[0].GetID(),
			Name:       lookingFileName,
			Option:     cacheOptionFunc(validFiles[0]),
		})
		if err != nil {
			utils.Log.Warnf("文件缓存失败:%s", err.Error())
		}
		return validFiles[0]
	} else {

		var lookedFile model.Obj
		nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+).mp4")

		if !nameRegexp.MatchString(lookingFileName) {
			lookedFile = validFiles[0]
			err := db.CreateMagnetCache(model.MagnetCache{
				DriverType: driverType,
				Magnet:     magnet,
				FileId:     lookedFile.GetID(),
				Name:       lookingFileName,
				Option:     cacheOptionFunc(lookedFile),
			})
			if err != nil {
				utils.Log.Warnf("文件缓存失败:%s", err.Error())
			}
		} else {
			code := nameRegexp.ReplaceAllString(lookingFileName, "$1")
			for index, file := range validFiles {
				realName := fmt.Sprintf("%s-cd%d.mp4", code, index+1)
				if realName == lookingFileName {
					lookedFile = file
				}
				err := db.CreateMagnetCache(model.MagnetCache{
					DriverType: driverType,
					Magnet:     magnet,
					FileId:     file.GetID(),
					Name:       realName,
					Option:     cacheOptionFunc(file),
				})
				if err != nil {
					utils.Log.Warnf("文件缓存失败:%s", err.Error())
				}
			}
		}
		return lookedFile
	}

}

func getLinkByCache(ctx context.Context, args model.LinkArgs, driverType string, storage driver.Driver, magnetCache model.MagnetCache) (*model.Link, error) {

	switch driverType {
	case "PikPak":
		link, err := storage.Link(ctx, &model.ObjThumb{
			Object: model.Object{ID: magnetCache.FileId},
		}, args)

		if err != nil {
			utils.Log.Infof("缓存文件已被删除，清除原缓存的文件：%s", err.Error())
		}
		return link, nil
	case "115 Cloud":
		link, err := storage.Link(ctx, &_115.FileObj{
			File: driver2.File{
				FileID:   magnetCache.FileId,
				PickCode: magnetCache.Option["pickCode"],
			},
		}, args)

		if err != nil {
			utils.Log.Infof("缓存文件已被删除，清除原缓存的文件：%s", err.Error())
		}
		return link, nil
	}

	return nil, nil

}
