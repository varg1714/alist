package virtual_file

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dlclark/regexp2"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func List(storageId uint, dir model.Obj, fileFunc func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error)) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	dirName := dir.GetName()
	utils.Log.Infof("list file:[%s]\n", dirName)

	virtualFilms := db.QueryVirtualFiles(strconv.Itoa(int(storageId)))

	virtualFileMap := make(map[string]model.VirtualFile)

	for _, film := range virtualFilms {
		virtualFileMap[film.Name] = film
	}

	if "root" == dirName {
		// 1. 顶级目录
		for _, category := range virtualFilms {
			results = append(results, &model.ObjThumb{
				Object: model.Object{
					Name:     category.Name,
					IsFolder: true,
					ID:       category.Name,
					Size:     622857143,
					Modified: category.Modified,
					Path:     filepath.Join(category.Name, category.ParentDir),
				},
			})
		}
		return results, nil
	}

	// top dir
	paths := strings.Split(dir.GetPath(), "/")
	if len(paths) == 0 {
		return results, nil
	}

	virtualFile, exist := virtualFileMap[paths[0]]

	if exist {

		// list files
		queriedFiles, err := recursiveListFile(dir, fileFunc, virtualFile)
		if err != nil {
			return queriedFiles, err
		}

		return prettyFiles(storageId, virtualFile, queriedFiles), nil

	} else {
		return results, nil
	}

}

func prettyFiles(storageId uint, virtualFile model.VirtualFile, queriedFiles []model.Obj) []model.Obj {

	results := make([]model.Obj, 0)

	// 手动修改的名称，优先级最高
	replacements := db.QueryReplacements(storageId, virtualFile.ShareID)
	replaceMap := make(map[string]string)
	for _, temp := range replacements {
		replaceMap[temp.OldName] = temp.NewName
	}

	for fileIndex, obj := range queriedFiles {

		excludeFile := virtualFile.ExcludeUnMatch
		objNameSetter, canRename := obj.(model.SetName)
		if !canRename && !excludeFile {
			results = append(results, obj)
			continue
		}

		// 规则重命名
		for replacedIndex := range virtualFile.Replace {

			// 匹配上规则，进行重命名
			replace := tryReplace(&virtualFile.Replace[replacedIndex], fileIndex, obj.GetName(), objNameSetter)
			if replace {
				excludeFile = true
				results = append(results, obj)
				break
			}

		}

		// 未匹配上重命名规则且不需要排除
		if !excludeFile {
			results = append(results, obj)
		}

		// 手动重命名
		if newName, ok := replaceMap[obj.GetID()]; ok {
			objNameSetter.SetName(newName)
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

		// 获取该文件夹下的文件
		objs, err := fileFunc(virtualFile, tempDir)
		if err != nil {
			return tempResults, err
		}

		for _, item := range objs {

			// 添加文件到结果集中，若是文件夹，则根据是否递归子文件夹来决定是否显示该文件夹
			if (!item.IsDir() && item.GetSize()/(1024*1024) >= virtualFile.MinFileSize) || (item.IsDir() && !virtualFile.AppendSubFolder) {
				tempResults = append(tempResults, item)
			}

			// 递归遍历子文件夹
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

	split := strings.Split(dir, "/")
	virtualFile := db.QueryVirtualFilm(storageId, split[0])

	return db.Rename(storageId, virtualFile.ShareID, oldName, newName)
}

func tryReplace(replaceItem *model.ReplaceItem, index int, oldName string, nameSetter model.SetName) bool {

	if replaceItem.SourceName == "" {
		return false
	}

	if replaceItem.Type == 0 {

		// 不在给定的重命名规则中
		if index < replaceItem.Start || (replaceItem.End != -1 && index > replaceItem.End) {
			return false
		}

		// 获取文件后缀名
		nameSetter.SetName(fmt.Sprintf("%s%02d%s", replaceItem.SourceName, replaceItem.StartNum, filepath.Ext(oldName)))
		replaceItem.StartNum += 1

	} else if replaceItem.Type == 1 && replaceItem.OldNameRegexp != "" {

		regexp := regexp2.MustCompile(replaceItem.OldNameRegexp, 0)
		matchString, err := regexp.MatchString(oldName)
		if err != nil {
			utils.Log.Infof("正则重命名影片:[%s]时出现错误,正则规则:[%v],错误原因:%v", oldName, replaceItem.OldNameRegexp, err)
			return false
		}

		if !matchString {
			return false
		}

		replacedName, err := regexp.Replace(oldName, replaceItem.SourceName, -1, -1)
		if err != nil {
			utils.Log.Infof("正则重命名影片:[%s]时出现错误,正则规则:[%v],错误原因:%v", oldName, replaceItem.OldNameRegexp, err)
			return false
		}
		if replacedName != "" && replacedName != oldName {
			nameSetter.SetName(replacedName)
		}

	}

	return true

}

// 转换为xml
func mediaToXML(m *Media) ([]byte, error) {
	// 转换
	x, err := xml.MarshalIndent(m, "", "  ")
	// 检查
	if err != nil {
		return nil, err
	}

	// 转码为[]byte
	x = []byte(xml.Header + string(x))

	return x, nil
}
