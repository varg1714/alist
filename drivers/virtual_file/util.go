package virtual_file

import (
	"errors"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic"
	"github.com/alist-org/alist/v3/pkg/utils"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func List(storageId uint, dir model.Obj, fileFunc func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error)) ([]model.Obj, error) {

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
					Path:     filepath.Join(category.Name),
				},
			})
		}
		return results, nil
	}

	// top dir
	paths := filepath.SplitList(dir.GetPath())
	if len(paths) == 0 {
		return results, nil
	}

	virtualFile, exist := virtualFilms[paths[0]]

	if exist {

		// list files
		tempResults, err := recursiveListFile(dir, fileFunc, virtualFile)
		if err != nil {
			return tempResults, err
		}

		return prettyFiles(storageId, virtualFile, tempResults), nil

	} else {
		return results, nil
	}

}

func prettyFiles(storageId uint, virtualFile model.VirtualFile, tempResults []model.Obj) []model.Obj {

	results := make([]model.Obj, 0)

	// name mapping
	replacements := db.QueryReplacements(storageId, virtualFile.ShareID)
	replaceMap := make(map[string]string)
	for _, temp := range replacements {
		replaceMap[temp.OldName] = temp.NewName
	}

	for fileIndex, obj := range tempResults {

		renameObj, canRename := obj.(model.SetName)

		// transfer file
		excludeFile := virtualFile.ExcludeUnMatch

		for testIndex := range virtualFile.Replace {

			if replace(virtualFile.Replace[testIndex], fileIndex) && canRename {

				var suffix string
				index := strings.LastIndex(obj.GetName(), ".")
				if index != -1 {
					suffix = obj.GetName()[index:]
				}

				tempNum := ""
				if virtualFile.Replace[testIndex].StartNum != -1 {
					tempNum = strconv.Itoa(virtualFile.Replace[testIndex].StartNum)
					if len(tempNum) == 1 {
						tempNum = "0" + tempNum
					}
				}

				renameObj.SetName(virtualFile.Replace[testIndex].SourceName + tempNum + suffix)
				virtualFile.Replace[testIndex].StartNum += 1

				results = append(results, obj)
				excludeFile = true

				break
			}

		}

		if !excludeFile {
			results = append(results, obj)
		}

		if newName, ok := replaceMap[obj.GetID()]; ok && canRename {
			renameObj.SetName(newName)
		}

	}
	return results
}

func recursiveListFile(dir model.Obj, fileFunc func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error), virtualFile model.VirtualFile) ([]model.Obj, error) {

	queue := generic.NewQueue[model.Obj]()
	queue.Push(dir)
	tempResults := make([]model.Obj, 0)

	for queue.Len() > 0 {

		tempDir := queue.Pop()

		// get files
		objs, err := fileFunc(virtualFile, tempDir)
		if err != nil {
			return tempResults, err
		}

		for _, item := range objs {

			if (!item.IsDir() && item.GetSize()/(1024*1024) >= virtualFile.MinFileSize) || (item.IsDir() && !virtualFile.AppendSubFolder) {
				tempResults = append(tempResults, item)
			}

			if item.IsDir() && virtualFile.AppendSubFolder {
				utils.Log.Infof("递归遍历子文件夹：[%s]", item.GetName())
				queue.Push(item)
			}

		}

	}
	return tempResults, nil
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

func replace(test model.ReplaceItem, index int) bool {

	if test.SourceName == "" {
		return false
	}

	if index >= test.Start && ((index <= test.End) || (test.End == -1)) {
		return true
	} else if test.Start == -1 && test.End == -1 {
		return true
	}

	return false

}
