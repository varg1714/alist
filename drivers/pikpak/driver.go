package pikpak

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	"regexp"
	"strconv"
	"time"
)

type PikPak struct {
	model.Storage
	Addition
	RefreshToken string
	AccessToken  string
}

func (d *PikPak) Config() driver.Config {
	return config
}

func (d *PikPak) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPak) Init(ctx context.Context) error {
	return d.login()
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
		link.URL = resp.Medias[0].Link.Url
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

func (d *PikPak) CloudDownload(ctx context.Context, parentDir string, dir string, name string, magnet string) ([]model.Obj, error) {

	// 1. 获取临时目录下所有文件
	var fileDir string
	parentFiles, err := d.getFiles(parentDir)
	if err != nil {
		return []model.Obj{}, err
	}

	// 1.1 在临时目录下寻找待下载的文件夹
	for _, parentFile := range parentFiles {
		if parentFile.Name == dir {
			fileDir = parentFile.Id
			break
		}
	}

	// 2.2 正在下载的文件所属文件夹不存在，新建文件夹
	if fileDir == "" {
		utils.Log.Infof("新建文件夹:[%s]\n", dir)
		var newDir CloudDownloadResp
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"kind":      "drive#folder",
				"parent_id": parentDir,
				"name":      dir,
			})
		}, &newDir)
		if err != nil || newDir.File.Id == "" {
			return []model.Obj{}, err
		}
		fileDir = newDir.File.Id
	}

	// 2. 尝试获取缓存文件
	var resultFile File

	// 2.1 获取缓存的文件ID
	fileIdCache := db.QueryFileId(magnet)

	// 2.2 判断该文件是否已下载
	files, err := d.getFiles(fileDir)
	if err != nil {
		return nil, err
	}
	for _, tempFile := range files {
		if tempFile.Id == fileIdCache || strings.HasPrefix(magnet, tempFile.Params.URL) {
			resultFile = tempFile
			break
		}
	}

	// 2.3 该文件在云盘已存在，直接返回该文件
	if resultFile.Id != "" {

		// 2.3.1 缓存此文件
		if fileIdCache == "" {
			err = db.CreateCacheFile(magnet, resultFile.Id, name)
			if err != nil {
				return []model.Obj{}, err
			}
		} else if fileIdCache != resultFile.Id {
			err = db.UpdateCacheFile(magnet, resultFile.Id, name)
			if err != nil {
				return []model.Obj{}, err
			}
		}

		// 2.3.2 返回结果
		if resultFile.Kind == "drive#file" {
			return utils.SliceConvert([]File{resultFile}, func(src File) (model.Obj, error) {
				return fileToObj(src), nil
			})
		} else {
			return d.List(ctx, &model.Object{
				ID: resultFile.Id,
			}, model.ListArgs{})
		}

	}

	// 3. 文件不存在，下载文件

	// 3.1 下载文件
	resultFile, err = d.downloadMagnet(fileDir, name, magnet)
	if err != nil || resultFile.Id == "" {
		return []model.Obj{}, err
	}

	// 3.2 缓存文件
	if resultFile.Kind == "drive#file" {
		// 3.2.1 下载结果为单文件，直接缓存
		err = db.CreateCacheFile(magnet, resultFile.Id, name)
		if err != nil {
			return []model.Obj{}, err
		}

	} else {
		// 3.2.2 下载结果为文件夹，进行文件夹清理
		utils.Log.Info("开始重命名文件")
		prettyFiles := d.prettyFile(fileDir, resultFile.Id, name)
		utils.Log.Info("重命名文件完成")
		var newFileId string
		if len(prettyFiles) == 0 {
		} else if len(prettyFiles) == 1 {
			newFileId = prettyFiles[0].Id
		} else {
			newFileId = resultFile.Id
		}
		err = db.CreateCacheFile(magnet, newFileId, name)
		if err != nil {
			return []model.Obj{}, err
		}

	}

	return utils.SliceConvert([]File{resultFile}, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})

}

func (d *PikPak) downloadMagnet(parentDir string, name string, magnet string) (File, error) {

	var resultFile File
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
		return resultFile, err
	}

	var count int
	for resultFile.Id == "" && count < 10 {

		utils.Log.Infof("文件未下载完毕，第[%d]次等待\n", count)
		if count != 0 {
			time.Sleep(3 * time.Second)
		}

		count++
		files, err := d.getFiles(parentDir)
		if err != nil {
			return resultFile, err
		}
		for _, tempFile := range files {
			if tempFile.Id == result.Task.FileID {
				resultFile = tempFile
				break
			}
		}
	}

	count = 0
	completed := false
	for resultFile.Kind != "drive#file" && !completed && count < 10 {

		if count != 0 {
			time.Sleep(1 * time.Second)
		}

		count++
		files, err := d.getFiles(resultFile.Id)
		if err != nil {
			return resultFile, err
		}
		for _, tempFile := range files {

			size, err := strconv.Atoi(tempFile.Size)
			if err != nil {
				utils.Log.Info("pretty file error:", err)
				return resultFile, err
			}

			if size/(1024*1024) > 100 {
				completed = true
				break
			}

		}
	}

	_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files/"+resultFile.Id, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(base.Json{
			"name": d.prettyName(name),
		})
	}, nil)

	if err != nil {
		return resultFile, err
	}

	return resultFile, nil
}

func (d *PikPak) prettyName(name string) string {

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

func (d *PikPak) prettyFile(parentDirId string, dirId string, name string) []File {

	files, err := d.getFiles(dirId)
	if err != nil {
		utils.Log.Info("get file error:", err)
		return []File{}
	}

	deletingFileIds := make([]string, 0)
	savedFiles := make([]File, 0)
	for _, file := range files {

		size, err := strconv.Atoi(file.Size)
		if err != nil {
			utils.Log.Info("pretty file error:", err)
			return files
		}

		if size/(1024*1024) < 150 {
			deletingFileIds = append(deletingFileIds, file.Id)
		} else {
			savedFiles = append(savedFiles, file)
		}

	}

	if len(savedFiles) == 1 {
		// rename file
		oldName := savedFiles[0].Name
		index := strings.LastIndex(oldName, ".")
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files/"+savedFiles[0].Id, http.MethodPatch, func(req *resty.Request) {
			req.SetBody(base.Json{
				"name": fmt.Sprintf("%s.%s", d.prettyName(name), oldName[index+1:]),
			})
		}, nil)

		if err != nil {
			utils.Log.Info("file rename error:", err)
			return savedFiles
		}

		utils.Log.Infof("重名名文件:[%s]\n", oldName)

		// move file
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{savedFiles[0].Id},
				"to": base.Json{
					"parent_id": parentDirId,
				},
			})
		}, nil)

		if err != nil {
			utils.Log.Info("move file error:", err)
			return savedFiles
		}

		utils.Log.Infof("移动文件:[%s]\n", oldName)

		// delete garbage file
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{dirId},
			})
		}, nil)
		if err != nil {
			utils.Log.Info("delete file error:", err)
			return savedFiles
		}

		time.Sleep(1 * time.Second)

		return savedFiles

	} else if len(deletingFileIds) > 0 {

		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": deletingFileIds,
			})
		}, nil)
		if err != nil {
			utils.Log.Info("pretty file error:", err)
		}

		time.Sleep(1 * time.Second)

		return savedFiles
	}

	utils.Log.Infof("重名名文件失败:[%v]\n", files)

	return []File{}

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
