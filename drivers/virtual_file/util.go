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

	virtualNames := db.QueryVirtualFileNames(strconv.Itoa(int(storageId)))

	if "root" == dirName {
		// 1. 顶级目录
		for category := range virtualNames {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     virtualNames[category],
					IsFolder: true,
					ID:       virtualNames[category],
					Size:     622857143,
					Modified: time.Now(),
				},
			})
		}
		return results, nil
	}

	if utils.SliceContains(virtualNames, dirName) {
		// 分享文件夹
		virtualFile := db.QueryVirtualFilms(storageId, dirName)
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

	virtualFiles := db.QueryVirtualFilms(storageId, req.Name)
	if virtualFiles.ShareID != "" {
		return errors.New("文件夹已存在")
	}

	req.StorageId = storageId

	return db.CreateVirtualFile(req)

}

func Rename(storageId uint, dir, oldName, newName string) error {

	return db.Rename(storageId, dir, oldName, newName)
}
