package pikpak

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"net/http"
	"strings"
	"sync"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"regexp"
	"strconv"
	"time"
)

type PikPak struct {
	model.Storage
	Addition

	oauth2Token oauth2.TokenSource
}

func (d *PikPak) Config() driver.Config {
	return config
}

func (d *PikPak) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPak) Init(ctx context.Context) (err error) {
	if d.ClientID == "" || d.ClientSecret == "" {
		d.ClientID = "YNxT9w7GMdWvEOKa"
		d.ClientSecret = "dbw2OtmVEeuUvIptb1Coyg"
	}

	withClient := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, oauth2.HTTPClient, base.HttpClient)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     d.ClientID,
		ClientSecret: d.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://user.mypikpak.com/v1/auth/signin",
			TokenURL:  "https://user.mypikpak.com/v1/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	oauth2Token, err := oauth2Config.PasswordCredentialsToken(withClient(ctx), d.Username, d.Password)
	if err != nil {
		return err
	}
	d.oauth2Token = oauth2Config.TokenSource(withClient(context.Background()), oauth2Token)
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
	_, err := d.request(fmt.Sprintf("https://api-drive.mypikpak.com/drive/v1/files/%s?_magic=2021&thumbnail_size=SIZE_LARGE", file.GetID()),
		http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}
	link := model.Link{
		URL: resp.WebContentLink,
	}
	if !d.DisableMediaLink && len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		link.URL = resp.Medias[d.LinkIndex].Link.Url
	}
	return &link, nil
}

func (d *PikPak) HdLink(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp File
	_, err := d.request(fmt.Sprintf("https://api-drive.mypikpak.com/drive/v1/files/%s?_magic=2021&thumbnail_size=SIZE_LARGE", file.GetID()),
		http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}
	link := model.Link{
		URL: resp.WebContentLink,
	}
	if len(resp.Medias) > 1 && resp.Medias[1].Link.Url != "" {
		log.Debugln("use media link")
		link.URL = resp.Medias[1].Link.Url
		return &link, nil
	}

	if len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		link.URL = resp.Medias[0].Link.Url
		return &link, nil
	}

	return &link, nil
}

func (d *PikPak) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"kind":      "drive#folder",
			"parent_id": parentDir.GetID(),
			"name":      dirName,
		})
	}, nil)
	return err
}

func (d *PikPak) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
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
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files/"+srcObj.GetID(), http.MethodPatch, func(req *resty.Request) {
		req.SetBody(base.Json{
			"name": newName,
		})
	}, nil)
	return err
}

func (d *PikPak) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchCopy", http.MethodPost, func(req *resty.Request) {
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
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
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
	res, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
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
	endpoint := strings.Join(strings.Split(params.Endpoint, ".")[1:], ".")
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(params.AccessKeyID, params.AccessKeySecret, params.SecurityToken),
		Region:      aws.String("pikpak"),
		Endpoint:    &endpoint,
	}
	ss, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(ss)
	if stream.GetSize() > s3manager.MaxUploadParts*s3manager.DefaultUploadPartSize {
		uploader.PartSize = stream.GetSize() / (s3manager.MaxUploadParts - 1)
	}
	input := &s3manager.UploadInput{
		Bucket: &params.Bucket,
		Key:    &params.Key,
		Body:   stream,
	}
	_, err = uploader.UploadWithContext(ctx, input)
	return err
}

func (d *PikPak) CloudDownload(ctx context.Context, parentDir string, dir model.Obj, magnetGetter func(obj model.Obj) (string, error)) ([]model.Obj, error) {

	// 1. 异步获取磁力链接
	magnet := ""
	var magnetWaiter sync.WaitGroup
	magnetWaiter.Add(1)
	go func() {
		start := time.Now().UnixMilli()
		defer magnetWaiter.Done()
		tempMagnet, err := magnetGetter(dir)
		if err != nil {
			utils.Log.Info("磁力链接获取失败", err)
		}
		magnet = tempMagnet
		utils.Log.Infof("获取磁力链接耗时:[%d]", time.Now().UnixMilli()-start)
	}()

	// 2. 尝试获取缓存文件
	index := strings.LastIndex(dir.GetName(), ".")
	name := dir.GetName()[:index]
	var resultFile File

	// 2.1 获取缓存的文件ID
	fileCache := db.QueryFileId(name)
	if fileCache.FileId != "" {
		existFile := d.getFile(fileCache.FileId)
		if existFile.Id != "" {
			return d.buildDownloadResult(ctx, existFile)
		}
	}

	// 2.2 判断该文件是否已下载
	// 2.2.1. 获取临时目录下的文件夹
	fileDir, err := d.getDir(parentDir, dir.GetPath())
	if fileDir == "" || err != nil {
		utils.Log.Info("文件夹创建失败", err)
		return []model.Obj{}, err
	}

	files, err := d.getFiles(fileDir)
	if err != nil {
		utils.Log.Info("文件夹信息获取失败", err)
		return []model.Obj{}, err
	}
	for _, tempFile := range files {
		if tempFile.Id == fileCache.FileId || fileCache.Magnet == tempFile.Params.URL || strings.Split(tempFile.Name, " ")[0] == fileCache.Code {
			resultFile = tempFile
			break
		}
	}

	// 2.3 该文件在云盘不存在，下载该文件
	newDownloaded := false
	var newFileDir File
	if resultFile.Id == "" {
		magnetWaiter.Wait()
		if magnet == "" {
			return []model.Obj{}, errors.New("磁力链接爬取结果为空！")
		}

		newFileDir, resultFile, err = d.downloadMagnet(fileDir, magnet)
		if err != nil || resultFile.Id == "" {
			return []model.Obj{}, err
		}
		newDownloaded = true
	}

	if newDownloaded {
		// 2.4.3 新下载的文件，进行文件夹清理
		go func() {
			utils.Log.Info("开始重命名文件")
			prettyFileId := d.prettyFile(fileDir, newFileDir.Id, name, newFileDir.Name, newFileDir.Kind != "drive#file")
			utils.Log.Info("重命名文件完成")

			if fileCache.FileId != "" {
				err = db.UpdateCacheFile(magnet, prettyFileId, name)
			} else {
				err = db.CreateCacheFile(magnet, prettyFileId, name)
			}
			if err != nil {
				utils.Log.Infof("缓存文件更新失败:%s-%s", name, newFileDir.Id)
			}

		}()
	} else if fileCache.FileId != resultFile.Id {
		utils.Log.Infof("更新缓存文件:%s-%s", name, resultFile.Id)
		err = db.UpdateCacheFile(magnet, resultFile.Id, name)
		if err != nil {
			utils.Log.Infof("缓存文件更新失败:%s-%s", name, resultFile.Id)
		}
	}

	// 2.4.2 返回结果
	return d.buildDownloadResult(ctx, resultFile)

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

func (d *PikPak) downloadMagnet(parentDir string, magnet string) (File, File, error) {

	var resultFile File
	var realFile File

	var result CloudDownloadResp

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
	}, &result)

	if err != nil {
		return resultFile, realFile, err
	}

	var count int
	var downloadedFiles []File

	for resultFile.Id == "" && count < 10 {

		if count != 0 {
			time.Sleep(2 * time.Second)
		}

		utils.Log.Infof("文件下载任务尚未提交完成，第[%d]次等待", count)
		count++
		downloadedFiles, err = d.getFiles(parentDir)
		if err != nil {
			return resultFile, realFile, err
		}
		for _, tempFile := range downloadedFiles {
			if tempFile.Id == result.Task.FileID {
				resultFile = tempFile
				break
			}
		}
	}

	fileDownloadCheck := func(checkingFiles []File) File {

		for _, tempFile := range checkingFiles {

			size, err := strconv.Atoi(tempFile.Size)
			if err != nil {
				utils.Log.Info("get file size error:", err)
				return realFile
			}

			if size/(1024*1024) > 100 {
				return tempFile
			}

		}

		return realFile
	}

	count = 0
	realFile = fileDownloadCheck(downloadedFiles)

	for resultFile.Kind != "drive#file" && realFile.Id == "" && count < 10 {

		if count != 0 {
			time.Sleep(1 * time.Second)
		}

		utils.Log.Infof("文件未下载完毕，第[%d]次等待", count)
		count++
		downloadedFiles, err = d.getFiles(resultFile.Id)
		if err != nil {
			return resultFile, realFile, err
		}
		realFile = fileDownloadCheck(downloadedFiles)

	}

	return resultFile, realFile, nil
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

var _ driver.Driver = (*PikPak)(nil)
