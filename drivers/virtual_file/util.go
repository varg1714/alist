package virtual_file

import (
	"errors"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"strconv"
	"time"
)

func List(storageId uint, dir model.Obj, fileFunc func(virtualFile model.VirtualFile) ([]model.Obj, error)) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	dirName := dir.GetName()
	utils.Log.Infof("list file:[%s]\n", dirName)

	virtualFilms := db.QueryVirtualFilms(strconv.Itoa(int(storageId)))

	if "root" == dirName {
		// 1. 顶级目录
		for _, category := range virtualFilms {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     category.Name,
					IsFolder: true,
					ID:       category.Name,
					Size:     622857143,
					Modified: time.Now(),
				},
			})
		}
		return results, nil
	}

	virtualFile, exist := virtualFilms[dirName]

	if exist {
		// 分享文件夹
		return fileFunc(virtualFile)
	} else {
		return results, nil
	}

}

func MakeDir(storageId uint, param string) error {

	var req model.VirtualFile
	err := utils.Json.Unmarshal([]byte(param), &req)
	if err != nil {
		return err
	}

	virtualFiles := db.QueryVirtualFilm(storageId, req.Name)
	if virtualFiles.ShareID != "" {
		return errors.New("文件夹已存在")
	}

	req.StorageId = storageId
	req.Modified = time.Now()

	return db.CreateVirtualFile(req)

}

func Rename(storageId uint, dir, oldName, newName string) error {

	return db.Rename(storageId, dir, oldName, newName)
}
