package pikpak

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"regexp"
	"time"
)

type PikPak struct {
	model.Storage
	Addition
	*Common
	RefreshToken string
	AccessToken  string
	oauth2Token  oauth2.TokenSource
}

func (d *PikPak) Config() driver.Config {
	return config
}

func (d *PikPak) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPak) Init(ctx context.Context) (err error) {

	if d.Common == nil {
		d.Common = &Common{
			client:       base.NewRestyClient(),
			CaptchaToken: "",
			UserID:       "",
			DeviceID:     utils.GetMD5EncodeStr(d.Username + d.Password),
			UserAgent:    "",
			RefreshCTokenCk: func(token string) {
				d.Common.CaptchaToken = token
				op.MustSaveDriverStorage(d)
			},
		}
	}

	if d.Platform == "android" {
		d.ClientID = AndroidClientID
		d.ClientSecret = AndroidClientSecret
		d.ClientVersion = AndroidClientVersion
		d.PackageName = AndroidPackageName
		d.Algorithms = AndroidAlgorithms
		d.UserAgent = BuildCustomUserAgent(utils.GetMD5EncodeStr(d.Username+d.Password), AndroidClientID, AndroidPackageName, AndroidSdkVersion, AndroidClientVersion, AndroidPackageName, "")
	} else if d.Platform == "web" {
		d.ClientID = WebClientID
		d.ClientSecret = WebClientSecret
		d.ClientVersion = WebClientVersion
		d.PackageName = WebPackageName
		d.Algorithms = WebAlgorithms
		d.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
	} else if d.Platform == "pc" {
		d.ClientID = PCClientID
		d.ClientSecret = PCClientSecret
		d.ClientVersion = PCClientVersion
		d.PackageName = PCPackageName
		d.Algorithms = PCAlgorithms
		d.UserAgent = "MainWindow Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) PikPak/2.5.6.4831 Chrome/100.0.4896.160 Electron/18.3.15 Safari/537.36"
	}

	if d.Addition.CaptchaToken != "" && d.Addition.RefreshToken == "" {
		d.SetCaptchaToken(d.Addition.CaptchaToken)
	}

	if d.Addition.DeviceID != "" {
		d.SetDeviceID(d.Addition.DeviceID)
	} else {
		d.Addition.DeviceID = d.Common.DeviceID
		op.MustSaveDriverStorage(d)
	}
	// 初始化 oauth2Config
	oauth2Config := &oauth2.Config{
		ClientID:     d.ClientID,
		ClientSecret: d.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://user.mypikpak.net/v1/auth/signin",
			TokenURL:  "https://user.mypikpak.net/v1/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	// 如果已经有RefreshToken，直接获取AccessToken
	if d.Addition.RefreshToken != "" {
		if d.RefreshTokenMethod == "oauth2" {
			// 使用 oauth2 刷新令牌
			// 初始化 oauth2Token
			d.initializeOAuth2Token(ctx, oauth2Config, d.Addition.RefreshToken)
			if err := d.refreshTokenByOAuth2(); err != nil {
				return err
			}
		} else {
			if err := d.refreshToken(d.Addition.RefreshToken); err != nil {
				return err
			}
		}

	} else {
		// 如果没有填写RefreshToken，尝试登录 获取 refreshToken
		if err := d.login(); err != nil {
			return err
		}
		if d.RefreshTokenMethod == "oauth2" {
			d.initializeOAuth2Token(ctx, oauth2Config, d.RefreshToken)
		}

	}

	// 获取CaptchaToken
	err = d.RefreshCaptchaTokenAtLogin(GetAction(http.MethodGet, "https://api-drive.mypikpak.net/drive/v1/files"), d.Common.GetUserID())
	if err != nil {
		return err
	}

	// 更新UserAgent
	if d.Platform == "android" {
		d.Common.UserAgent = BuildCustomUserAgent(utils.GetMD5EncodeStr(d.Username+d.Password), AndroidClientID, AndroidPackageName, AndroidSdkVersion, AndroidClientVersion, AndroidPackageName, d.Common.UserID)
	}

	// 保存 有效的 RefreshToken
	d.Addition.RefreshToken = d.RefreshToken
	op.MustSaveDriverStorage(d)

	return nil
}

func (d *PikPak) Drop(ctx context.Context) error {
	return nil
}

func (d *PikPak) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *PikPak) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	var resp File
	queryParams := map[string]string{
		"_magic":         "2021",
		"usage":          "FETCH",
		"thumbnail_size": "SIZE_LARGE",
	}
	if !d.DisableMediaLink {
		queryParams["usage"] = "CACHE"
	}
	_, err := d.request(fmt.Sprintf("https://api-drive.mypikpak.net/drive/v1/files/%s", file.GetID()),
		http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(queryParams)
		}, &resp)
	if err != nil {
		return nil, err
	}
	exp := time.Minute
	link := model.Link{
		URL:        resp.WebContentLink,
		Expiration: &exp,
	}

	if !d.DisableMediaLink && len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		if len(resp.Medias) > int(d.LinkIndex) {
			link.URL = resp.Medias[d.LinkIndex].Link.Url
		} else {
			link.URL = resp.Medias[0].Link.Url
		}
	}

	utils.Log.Infof("pikpak返回的地址: %s", link.URL)
	return &link, err
}

func (d *PikPak) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":      "drive#folder",
			"parent_id": parentDir.GetID(),
			"name":      dirName,
		})
	}, nil)
	return err
}

func (d *PikPak) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{srcObj.GetID()},
			"to": base.Json{
				"parent_id": dstDir.GetID(),
			},
		})
	}, nil)
	return err
}

func (d *PikPak) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files/"+srcObj.GetID(), http.MethodPatch, func(req *resty.Request) {
		req.SetBody(base.Json{
			"name": newName,
		})
	}, nil)
	return err
}

func (d *PikPak) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files:batchCopy", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{srcObj.GetID()},
			"to": base.Json{
				"parent_id": dstDir.GetID(),
			},
		})
	}, nil)
	return err
}

func (d *PikPak) Remove(ctx context.Context, obj model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"ids": []string{obj.GetID()},
		})
	}, nil)
	return err
}

func (d *PikPak) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	hi := stream.GetHash()
	sha1Str := hi.GetHash(hash_extend.GCID)
	if len(sha1Str) < hash_extend.GCID.Width {
		tFile, err := stream.CacheFullInTempFile()
		if err != nil {
			return err
		}

		sha1Str, err = utils.HashFile(hash_extend.GCID, tFile, stream.GetSize())
		if err != nil {
			return err
		}
	}

	var resp UploadTaskData
	res, err := d.request("https://api-drive.mypikpak.net/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":        "drive#file",
			"name":        stream.GetName(),
			"size":        stream.GetSize(),
			"hash":        strings.ToUpper(sha1Str),
			"upload_type": "UPLOAD_TYPE_RESUMABLE",
			"objProvider": base.Json{"provider": "UPLOAD_TYPE_UNKNOWN"},
			"parent_id":   dstDir.GetID(),
			"folder_type": "NORMAL",
		})
	}, &resp)
	if err != nil {
		return err
	}

	// 秒传成功
	if resp.Resumable == nil {
		log.Debugln(string(res))
		return nil
	}

	params := resp.Resumable.Params
	//endpoint := strings.Join(strings.Split(params.Endpoint, ".")[1:], ".")
	// web 端上传 返回的endpoint 为 `mypikpak.net` | android 端上传 返回的endpoint 为 `vip-lixian-07.mypikpak.net`·
	if d.Addition.Platform == "android" {
		params.Endpoint = "mypikpak.net"
	}

	if stream.GetSize() <= 10*utils.MB { // 文件大小 小于10MB，改用普通模式上传
		return d.UploadByOSS(&params, stream, up)
	}
	// 分片上传
	return d.UploadByMultipart(&params, stream.GetSize(), stream, up)
}

func (d *PikPak) CloudDownload(ctx context.Context, parentDir string, downloadingFile model.Obj, magnetGetter func(obj model.Obj) (string, error)) ([]model.Obj, error) {

	// 1. 异步获取磁力链接
	magnet := ""
	var magnetWaiter sync.WaitGroup
	magnetWaiter.Add(1)
	go func() {
		start := time.Now().UnixMilli()
		defer magnetWaiter.Done()
		tempMagnet, err := magnetGetter(downloadingFile)
		if err != nil {
			utils.Log.Info("磁力链接获取失败", err)
		}
		magnet = tempMagnet
		utils.Log.Infof("获取:%s的磁力链接结果为:[%s]耗时:[%d]", downloadingFile.GetName(), magnet, time.Now().UnixMilli()-start)
	}()

	// 2. 尝试获取缓存文件
	fileName := downloadingFile.GetName()

	// 2.1 获取缓存的文件ID
	fileCache := db.QueryFileCacheByName(fileName)
	if fileCache.FileId != "" {
		existFile := d.getFile(fileCache.FileId)
		if existFile.Id != "" {
			return d.buildDownloadResult(ctx, existFile)
		}
	}

	// 2.2 判断该文件是否已下载
	// 2.2.1. 获取临时目录下的文件夹
	fileDir, err := d.getDir(parentDir, downloadingFile.GetPath())
	if fileDir == "" || err != nil {
		utils.Log.Info("文件夹创建失败", err)
		return []model.Obj{}, err
	}

	// 2.3 该文件在云盘不存在，下载该文件
	var downloadFile File
	magnetWaiter.Wait()
	if magnet == "" {
		return []model.Obj{}, nil
	}

	downloadFile, err = d.downloadMagnet(fileDir, fileName, magnet)
	if err != nil || downloadFile.Id == "" {
		return []model.Obj{}, err
	}

	return d.buildDownloadResult(ctx, downloadFile)

}

func (d *PikPak) buildDownloadResult(ctx context.Context, resultFile File) ([]model.Obj, error) {

	if resultFile.Kind == "drive#file" {
		// 1. 单文件，直接返回
		return utils.SliceConvert([]File{resultFile}, func(src File) (model.Obj, error) {
			return fileToObj(src), nil
		})
	} else {
		// 2. 文件夹，返回文件大小最大的文件
		fileList, err := d.List(ctx, &model.Object{
			ID: resultFile.Id,
		}, model.ListArgs{})
		if err != nil {
			return fileList, err
		}
		slices.SortFunc(fileList, func(a, b model.Obj) int {
			return cmp.Compare(b.GetSize(), a.GetSize())
		})

		return fileList, err
	}

}

func (d *PikPak) getDir(parentDirId string, dirName string) (string, error) {

	var fileDir string

	// 1. 获取父级文件夹下的文件
	parentFiles, err := d.getFiles(parentDirId)
	if err != nil {
		return "", err
	}

	// 1.1 在临时目录下寻找待下载的文件夹
	for _, parentFile := range parentFiles {
		if parentFile.Name == dirName {
			fileDir = parentFile.Id
			break
		}
	}

	// 2. 正在下载的文件所属文件夹不存在，新建文件夹
	if fileDir == "" {
		utils.Log.Infof("新建文件夹:[%s]\n", dirName)
		var newDir CloudDownloadResp
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"kind":      "drive#folder",
				"parent_id": parentDirId,
				"name":      dirName,
			})
		}, &newDir)
		if err != nil || newDir.File.Id == "" {
			return "", err
		}
		fileDir = newDir.File.Id
	}

	return fileDir, nil
}

func (d *PikPak) downloadMagnet(parentDir string, fileName, magnet string) (File, error) {

	var downloadFile File

	var downloadResp CloudDownloadResp

	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":        "drive#file",
			"upload_type": "UPLOAD_TYPE_URL",
			"params": base.Json{
				"with_thumbnail": "true",
				"from":           "manual",
			},
			"url": base.Json{
				"url": magnet,
			},
			"parent_id": parentDir,
		})
	}, &downloadResp)

	if err != nil {
		return downloadFile, err
	}

	var count int
	downloading := true

	for downloading && count < 10 {

		if count != 0 {
			time.Sleep(2 * time.Second)
		}

		utils.Log.Infof("文件下载任务尚未完成，第[%d]次等待", count)
		count++

		tasks, taskErr := d.getTasks()
		if taskErr != nil {
			return downloadFile, err
		}

		find := false
		for _, tempFile := range tasks.Tasks {
			if tempFile.FileID == downloadResp.Task.FileID {
				// 还在下载中
				find = true
				utils.Log.Infof("当前下载进度:%d", tempFile.Progress)
				break
			}
		}

		if !find {
			// 下载完成
			downloading = false
		}

	}

	var validFiles []File

	files, listFileErr := d.getFiles(downloadResp.Task.FileID)
	if listFileErr != nil {
		return downloadFile, err
	}

	// 下载完的文件
	for _, tempFile := range files {
		size, err1 := strconv.Atoi(tempFile.Size)
		if err1 != nil {
			utils.Log.Info("get file size error:", err)
			return downloadFile, err
		}

		if size/(1024*1024) > 100 {
			validFiles = append(validFiles, tempFile)
		}

	}

	if len(validFiles) == 0 {
		return downloadFile, nil
	}

	slices.SortFunc(validFiles, func(a, b File) int {
		return strings.Compare(a.Name, b.Name)
	})

	if len(validFiles) == 1 {
		downloadFile = validFiles[0]
		err1 := db.CreateCacheFile(magnet, validFiles[0].Id, fileName)
		if err1 != nil {
			utils.Log.Info("文件缓存失败:%s", err1.Error())
			return downloadFile, err1
		}
	} else {
		nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+)?.mp4")
		code := nameRegexp.ReplaceAllString(fileName, "$1")
		for index, file := range validFiles {
			realName := fmt.Sprintf("%s-cd%d.mp4", code, index+1)
			if realName == fileName {
				downloadFile = file
			}
			err1 := db.CreateCacheFile(magnet, file.Id, realName)
			if err1 != nil {
				utils.Log.Info("文件缓存失败:%s", err1.Error())
				return downloadFile, err1
			}
		}
	}

	if downloadFile.Id == "" {
		downloadFile = validFiles[0]
		err1 := db.CreateCacheFile(magnet, downloadFile.Id, fileName)
		if err1 != nil {
			utils.Log.Info("文件缓存失败:%s", err1.Error())
			return downloadFile, err1
		}
	}

	return downloadFile, nil

}

// clearIllegalChar 清理文件名中的非法字符
func (d *PikPak) clearIllegalChar(name string) string {

	pattern := d.FileNameBlackChars
	if pattern == "" {
		pattern = "*:"
	}

	renamePattern, err := regexp.Compile(fmt.Sprintf("[%s]", pattern))
	if err != nil {
		utils.Log.Info("fileNameBlackChars error:", err)
		return name
	}

	return renamePattern.ReplaceAllString(name, "")
}

// prettyFile 清零下载的文件：删除垃圾文件；单个文件不保留文件夹；文件夹重命名；
func (d *PikPak) prettyFile(parentDirId string, prettyFileId string, prettyName, oldName string, isDir bool) string {

	clearFunc := func(fileId, name string) {
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files/"+fileId, http.MethodPatch, func(req *resty.Request) {
			req.SetBody(base.Json{
				"name": d.clearIllegalChar(name),
			})
		}, nil)
		if err != nil {
			utils.Log.Warn("文件重命名失败", err)
		}
		utils.Log.Infof("重名名文件:[%s]", oldName)
	}

	if !isDir {
		// 1. 不是文件夹，直接重命名该文件
		index := strings.LastIndex(oldName, ".")
		clearFunc(prettyFileId, fmt.Sprintf("%s.%s", prettyName, oldName[:index]))
		return prettyFileId
	}

	// 2. 文件夹，开始清理文件
	files, err := d.getFiles(prettyFileId)
	if err != nil {
		utils.Log.Info("get file error:", err)
		return prettyFileId
	}

	// 2.1 扫描出要保留及要删除的文件
	deletingFileIds := make([]string, 0)
	savedFiles := make([]File, 0)
	for _, file := range files {

		size, err := strconv.Atoi(file.Size)
		if err != nil {
			utils.Log.Info("pretty file error:", err)
			return prettyFileId
		}

		if size/(1024*1024) < 150 {
			deletingFileIds = append(deletingFileIds, file.Id)
		} else {
			savedFiles = append(savedFiles, file)
		}

	}

	// 2.2 开始清理文件
	if len(savedFiles) == 1 {

		// 2.2.1 文件夹下只有一个文件，开始清理文件夹
		oldName := savedFiles[0].Name
		index := strings.LastIndex(oldName, ".")
		newFileId := savedFiles[0].Id

		clearFunc(newFileId, fmt.Sprintf("%s.%s", d.clearIllegalChar(prettyName), oldName[index+1:]))

		// 2.2.2 移动文件到指定文件夹
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{newFileId},
				"to": base.Json{
					"parent_id": parentDirId,
				},
			})
		}, nil)

		if err != nil {
			utils.Log.Info("move file error:", err)
			return newFileId
		}
		utils.Log.Infof("移动文件:[%s]", oldName)

		// 2.3.3 删除保留单个文件外的其他文件
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{prettyFileId},
			})
		}, nil)
		if err != nil {
			utils.Log.Info("delete file error:", err)
			return newFileId
		}

		return newFileId

	}

	// 3. 文件夹下不止一个文件，保留文件夹并清理文件夹下其他的垃圾文件
	if len(deletingFileIds) > 0 {
		utils.Log.Infof("文件夹下有多个视频文件，仅重命名文件夹:[%v]", savedFiles)
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": deletingFileIds,
			})
		}, nil)
		if err != nil {
			utils.Log.Info("pretty file error:", err)
		}

		return prettyFileId
	}

	return prettyFileId

}

func (d *PikPak) Download(ctx context.Context, url, downloadDir string) error {

	var result CloudDownloadResp

	resp, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":        "drive#file",
			"upload_type": "UPLOAD_TYPE_URL",
			"params": base.Json{
				"with_thumbnail": "true",
				"from":           "manual",
			},
			"url": base.Json{
				"url": url,
			},
			"parent_id": downloadDir,
		})
	}, &result)

	if err != nil {
		utils.Log.Warnf("文件下载失败，失败原因为:[%s]", string(resp))
	}

	return err

}

// 离线下载文件
func (d *PikPak) OfflineDownload(ctx context.Context, fileUrl string, parentDir model.Obj, fileName string) (*OfflineTask, error) {
	requestBody := base.Json{
		"kind":        "drive#file",
		"name":        fileName,
		"upload_type": "UPLOAD_TYPE_URL",
		"url": base.Json{
			"url": fileUrl,
		},
		"parent_id":   parentDir.GetID(),
		"folder_type": "",
	}

	var resp OfflineDownloadResp
	_, err := d.request("https://api-drive.mypikpak.net/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(requestBody)
	}, &resp)

	if err != nil {
		return nil, err
	}

	return &resp.Task, err
}

/*
获取离线下载任务列表
phase 可能的取值：
PHASE_TYPE_RUNNING, PHASE_TYPE_ERROR, PHASE_TYPE_COMPLETE, PHASE_TYPE_PENDING
*/
func (d *PikPak) OfflineList(ctx context.Context, nextPageToken string, phase []string) ([]OfflineTask, error) {
	res := make([]OfflineTask, 0)
	url := "https://api-drive.mypikpak.net/drive/v1/tasks"

	if len(phase) == 0 {
		phase = []string{"PHASE_TYPE_RUNNING", "PHASE_TYPE_ERROR", "PHASE_TYPE_COMPLETE", "PHASE_TYPE_PENDING"}
	}
	params := map[string]string{
		"type":           "offline",
		"thumbnail_size": "SIZE_SMALL",
		"limit":          "10000",
		"page_token":     nextPageToken,
		"with":           "reference_resource",
	}

	// 处理 phase 参数
	if len(phase) > 0 {
		filters := base.Json{
			"phase": map[string]string{
				"in": strings.Join(phase, ","),
			},
		}
		filtersJSON, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filters: %w", err)
		}
		params["filters"] = string(filtersJSON)
	}

	var resp OfflineListResp
	_, err := d.request(url, http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx).
			SetQueryParams(params)
	}, &resp)

	if err != nil {
		return nil, fmt.Errorf("failed to get offline list: %w", err)
	}
	res = append(res, resp.Tasks...)
	return res, nil
}

func (d *PikPak) DeleteOfflineTasks(ctx context.Context, taskIDs []string, deleteFiles bool) error {
	url := "https://api-drive.mypikpak.net/drive/v1/tasks"
	params := map[string]string{
		"task_ids":     strings.Join(taskIDs, ","),
		"delete_files": strconv.FormatBool(deleteFiles),
	}
	_, err := d.request(url, http.MethodDelete, func(req *resty.Request) {
		req.SetContext(ctx).
			SetQueryParams(params)
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete tasks %v: %w", taskIDs, err)
	}
	return nil
}

var _ driver.Driver = (*PikPak)(nil)
