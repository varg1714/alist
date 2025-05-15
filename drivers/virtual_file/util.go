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
	"strings"
	"time"
)

func List(storageId uint, dir model.Obj, fileFunc func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error)) ([]model.Obj, error) {

	results := make([]model.Obj, 0)

	dirName := dir.GetName()
	utils.Log.Infof("list file:[%s]", dirName)

	parent := ""
	dbQuery := false
	if dirName == "root" {
		dbQuery = true
	} else if virDir, ok := dir.(*model.ObjVirtualDir); ok && virDir.VirtualFile.DirType == model.VirtualDirection {
		parent = fmt.Sprintf("%d", virDir.VirtualFile.ID)
		dbQuery = true
	}

	if dbQuery {
		// 1. 顶级目录
		virtualFiles := db.QueryVirtualFiles(storageId, parent)

		results = append(results, buildVirtualFiles(virtualFiles, dir)...)

		if dirName != "root" {
			movedFiles, err := getMovedFiles(storageId, dir, fileFunc)
			if err != nil {
				utils.Log.Warnf("failed to get moved files, error message: %s", err.Error())
			} else if len(movedFiles) > 0 {
				results = append(results, movedFiles...)
			}
		}

		return results, nil
	}

	virtualFile := GetVirtualFile(storageId, dir.GetPath())

	if virtualFile.ShareID != "" {

		// list files
		queriedFiles, err := recursiveListFile(dir, fileFunc, virtualFile)
		if err != nil {
			return queriedFiles, err
		}

		movedItems, err := db.QueryMovedItemByShareId(storageId, virtualFile.ShareID, filepath.Dir(dir.GetPath()))
		if err != nil {
			utils.Log.Warnf("failed to list moved items by shareId, error message: %s", err.Error())
		}
		var movedItemIds []string
		for _, item := range movedItems {
			movedItemIds = append(movedItemIds, item.FileId)
		}

		return prettyFiles(storageId, virtualFile, queriedFiles, movedItemIds), nil

	} else {
		return results, nil
	}

}

func prettyFiles(storageId uint, virtualFile model.VirtualFile, queriedFiles []model.Obj, excludeIds []string) []model.Obj {

	results := make([]model.Obj, 0)

	// 手动修改的名称，优先级最高
	replacements := db.QueryReplacements(storageId, virtualFile.ShareID)
	replaceMap := make(map[string]model.Replacement)
	for _, temp := range replacements {
		replaceMap[temp.OldName] = temp
	}

	excludeIdMap := make(map[string]bool)
	for _, id := range excludeIds {
		excludeIdMap[id] = true
	}

	for fileIndex, obj := range queriedFiles {

		excludeFile := virtualFile.ExcludeUnMatch
		objNameSetter, canRename := obj.(model.SetName)
		if !canRename && !excludeFile {
			results = append(results, obj)
			continue
		}

		if excludeIdMap[obj.GetID()] {
			continue
		}

		// 重命名或者移除文件
		if replacement, ok := replaceMap[obj.GetID()]; ok {
			if replacement.Type == 1 {
				excludeFile = true
			} else {
				objNameSetter.SetName(replacement.NewName)
			}
		} else {
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
		}

		// 未匹配上重命名规则且不需要排除
		if !excludeFile {
			results = append(results, obj)
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

func MakeDir(storageId uint, parentDir model.Obj, param string) error {

	var req model.VirtualFile
	if strings.HasPrefix(param, "{") {
		err := utils.Json.Unmarshal([]byte(param), &req)
		if err != nil {
			return err
		}
	} else {
		req.Name = param
		req.DirType = model.VirtualDirection
	}

	if parentDir.GetName() == "root" {
		req.Parent = ""
	} else if virDir, ok := parentDir.(*model.ObjVirtualDir); ok && virDir.DirType == model.VirtualDirection {
		req.Parent = fmt.Sprintf("%d", virDir.ID)
	} else {
		return errors.New("不允许在此目录下新增文件夹")
	}

	virtualFiles := db.QueryVirtualFiles(storageId, req.Parent)
	for _, file := range virtualFiles {
		if file.Name == req.Name {
			return errors.New("文件夹已存在")
		}
	}

	req.StorageId = storageId
	req.Modified = time.Now()

	return db.CreateVirtualFile(req)

}

func GetVirtualFile(storageId uint, path string) model.VirtualFile {

	virtualMapFunc := func(parent string) map[string]model.VirtualFile {

		virtualFiles := db.QueryVirtualFiles(storageId, parent)
		virtualFileMap := make(map[string]model.VirtualFile)
		for _, film := range virtualFiles {
			virtualFileMap[fmt.Sprintf("%d", film.ID)] = film
		}

		return virtualFileMap
	}

	split := strings.Split(path, "/")
	virtualMap := virtualMapFunc("")
	for _, subPath := range split {
		if virtualFile, exist := virtualMap[subPath]; exist && virtualFile.DirType == model.Subscription {
			return virtualFile
		} else if subPath != "" {
			virtualMap = virtualMapFunc(subPath)
		}
	}

	return model.VirtualFile{}

}

func DeleteVirtualFile(storageId uint, obj model.Obj) error {

	if virDir, ok := obj.(*model.ObjVirtualDir); ok {
		if virDir.VirtualFile.DirType == model.Subscription {
			return db.DeleteVirtualFile(virDir.VirtualFile)
		} else {
			// remove all sub dir
			queue := generic.NewQueue[model.VirtualFile]()
			queue.Push(virDir.VirtualFile)
			for !queue.IsEmpty() {
				pop := queue.Pop()
				files := db.QueryVirtualFiles(storageId, fmt.Sprintf("%d", pop.ID))
				for _, file := range files {
					queue.Push(file)
				}
				err := db.DeleteVirtualFile(pop)
				if err != nil {
					return err
				}
			}
		}
	} else {
		virtualFile := GetVirtualFile(storageId, obj.GetPath())
		// delete file
		replacement := model.Replacement{
			StorageId: storageId,
			DirName:   virtualFile.ShareID,
			Type:      1,
			OldName:   obj.GetID(),
		}
		return db.CreateReplacement(replacement)
	}

	return nil

}

func Move(storageId uint, srcObj model.Obj, targetDir model.Obj) error {

	targetVirtualFile := GetVirtualFile(storageId, targetDir.GetPath())
	if targetVirtualFile.ShareID != "" {
		return errors.New("仅能移动到虚拟文件夹下")
	}

	if virDir, ok := srcObj.(*model.ObjVirtualDir); ok {
		// update parent
		virtualFile := virDir.VirtualFile
		virtualFile.Parent = targetDir.GetID()
		return db.UpdateVirtualFile(virtualFile)
	} else {
		// add moved item
		sourceVirtualFile := GetVirtualFile(storageId, srcObj.GetPath())

		existMovedItem, err := db.QueryMovedItemByFileId(srcObj.GetID(), filepath.Dir(srcObj.GetPath()))
		if err != nil {
			return err
		}

		if existMovedItem.FileId != "" {
			// has been moved, only update the parent info
			existMovedItem.Parent = targetDir.GetID()
			return db.UpdateMovedItem(existMovedItem)
		} else {
			// add new moved info
			return db.CreateMovedItem(model.MovedItem{
				Parent:    targetDir.GetID(),
				StorageId: storageId,
				ShareId:   sourceVirtualFile.ShareID,
				Source:    filepath.Dir(srcObj.GetPath()),
				FileId:    srcObj.GetID(),
			})
		}

	}

}

func Rename(storageId uint, dir, oldName, newName string) error {

	virtualFile := GetVirtualFile(storageId, dir)

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
		if replaceItem.StartNum >= 0 {
			nameSetter.SetName(fmt.Sprintf("%s%02d%s", replaceItem.SourceName, replaceItem.StartNum, filepath.Ext(oldName)))
			replaceItem.StartNum += 1
		} else {
			nameSetter.SetName(fmt.Sprintf("%s%s", replaceItem.SourceName, filepath.Ext(oldName)))
		}

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

func buildVirtualFiles(virtualFiles []model.VirtualFile, dir model.Obj) []model.Obj {

	var result []model.Obj

	for _, virtualFile := range virtualFiles {
		result = append(result, &model.ObjVirtualDir{
			ObjThumb: model.ObjThumb{
				Object: model.Object{
					Name:     virtualFile.Name,
					IsFolder: true,
					ID:       fmt.Sprintf("%d", virtualFile.ID),
					Size:     622857143,
					Modified: virtualFile.Modified,
					Path: func() string {
						if virtualFile.DirType == model.VirtualDirection {
							return filepath.Join(dir.GetPath(), fmt.Sprintf("%d", virtualFile.ID))
						} else {
							return filepath.Join(dir.GetPath(), fmt.Sprintf("%d", virtualFile.ID), virtualFile.ParentDir)
						}
					}(),
				},
			},
			VirtualFile: virtualFile,
		})
	}

	return result

}

func getMovedFiles(storageId uint, dir model.Obj, fileFunc func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error)) ([]model.Obj, error) {

	var result []model.Obj

	// query moved films
	movedItems, err := db.QueryMovedItemByParent(dir.GetID())
	if err != nil {
		utils.Log.Warnf("failed to query moved itms, error message: %s", err.Error())
		return result, err
	}

	if len(movedItems) == 0 {
		return result, nil
	}

	movedPath := make(map[string][]string)
	for _, item := range movedItems {
		movedPath[item.Source] = append(movedPath[item.Source], item.FileId)
	}

	for path, parentObjs := range movedPath {
		virtualFile := GetVirtualFile(storageId, path)
		if virtualFile.ShareID == "" {
			// virtual dir, only need to query subdirectories
			subFiles, err1 := db.QueryVirtualFilesById(storageId, parentObjs)
			if err1 != nil {
				return result, err1
			}
			result = append(result, buildVirtualFiles(subFiles, dir)...)
		} else {

			// query files
			subFiles, err1 := recursiveListFile(&model.ObjThumb{Object: model.Object{
				Path: path,
			}}, fileFunc, virtualFile)
			if err1 != nil {
				return result, err1
			}

			movedIdSet := make(map[string]bool)
			for _, obj := range parentObjs {
				movedIdSet[obj] = true
			}

			var tempFiles []model.Obj
			for _, file := range subFiles {
				if movedIdSet[file.GetID()] {
					tempFiles = append(tempFiles, file)
				}
			}

			result = append(result, prettyFiles(storageId, virtualFile, tempFiles, []string{})...)

		}

	}

	return result, nil

}
