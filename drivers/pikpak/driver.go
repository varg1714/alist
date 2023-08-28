package pikpak

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	if len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
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
	tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	// cal sha1
	s := sha1.New()
	_, err = io.Copy(s, tempFile)
	if err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	sha1Str := hex.EncodeToString(s.Sum(nil))
	data := base.Json{
		"kind":        "drive#file",
		"name":        stream.GetName(),
		"size":        stream.GetSize(),
		"hash":        strings.ToUpper(sha1Str),
		"upload_type": "UPLOAD_TYPE_RESUMABLE",
		"objProvider": base.Json{"provider": "UPLOAD_TYPE_UNKNOWN"},
		"parent_id":   dstDir.GetID(),
	}
	res, err := d.request("https://api-drive.mypikpak.com/drive/v1/files", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	if err != nil {
		return err
	}
	if stream.GetSize() == 0 {
		log.Debugln(string(res))
		return nil
	}
	params := jsoniter.Get(res, "resumable").Get("params")
	endpoint := params.Get("endpoint").ToString()
	endpointS := strings.Split(endpoint, ".")
	endpoint = strings.Join(endpointS[1:], ".")
	accessKeyId := params.Get("access_key_id").ToString()
	accessKeySecret := params.Get("access_key_secret").ToString()
	securityToken := params.Get("security_token").ToString()
	key := params.Get("key").ToString()
	bucket := params.Get("bucket").ToString()
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, securityToken),
		Region:      aws.String("pikpak"),
		Endpoint:    &endpoint,
	}
	ss, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(ss)
	input := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   tempFile,
	}
	_, err = uploader.UploadWithContext(ctx, input)
	return err
}

func (d *PikPak) CloudDownload(ctx context.Context, parentDir string, dir string, name string, magnet string) ([]model.Obj, error) {

	var resultFile File
	fileIdCache := db.QueryFileId(magnet)

	var fileDir string
	parentFiles, err := d.getFiles(parentDir)
	if err != nil {
		return []model.Obj{}, err
	}

	for _, parentFile := range parentFiles {
		if parentFile.Name == dir {
			fileDir = parentFile.Id
			break
		}
	}
	if fileDir == "" {
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

	if fileIdCache == "" {
		// file don't exist
		downloadFile, err := d.downloadMagnet(fileDir, name, magnet)
		resultFile = downloadFile
		if err != nil || resultFile.Id == "" {
			return []model.Obj{}, err
		}

		err = db.CreateCacheFile(magnet, resultFile.Id)
		if err != nil {
			return []model.Obj{}, err
		}

	} else {
		// get cache file
		files, err := d.getFiles(fileDir)
		if err != nil {
			return nil, err
		}
		for _, tempFile := range files {
			if tempFile.Id == fileIdCache {
				resultFile = tempFile
				break
			}
		}

		if resultFile.Id == "" {
			resultFile, err = d.downloadMagnet(fileDir, name, magnet)
			if err != nil || resultFile.Id == "" {
				return []model.Obj{}, err
			}

			err = db.UpdateCacheFile(magnet, resultFile.Id)
			if err != nil {
				return []model.Obj{}, err
			}
		}

	}

	// File
	if resultFile.Kind == "drive#file" {
		return utils.SliceConvert([]File{resultFile}, func(src File) (model.Obj, error) {
			return fileToObj(src), nil
		})
	} else {
		// Folder
		// pretty file
		newFileId := d.prettyFile(fileDir, resultFile.Id, name)
		if newFileId != resultFile.Id {
			err = db.UpdateCacheFile(magnet, newFileId)
			if err != nil {
				return []model.Obj{}, err
			}
		}

		return d.List(ctx, &model.Object{
			ID: newFileId,
		}, model.ListArgs{})

	}

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

		if count != 0 {
			time.Sleep(1 * time.Second)
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

			if size/(1024*1024) > 50 {
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
		return resultFile, nil
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

func (d *PikPak) prettyFile(parentDirId string, dirId string, name string) string {

	files, err := d.getFiles(dirId)
	if err != nil {
		utils.Log.Info("get file error:", err)
		return dirId
	}

	deletingFileIds := make([]string, 0)
	savedFileIds := make([]File, 0)
	for _, file := range files {

		size, err := strconv.Atoi(file.Size)
		if err != nil {
			utils.Log.Info("pretty file error:", err)
			return dirId
		}

		if size/(1024*1024) < 50 {
			deletingFileIds = append(deletingFileIds, file.Id)
		} else {
			savedFileIds = append(savedFileIds, file)
		}

	}

	if len(savedFileIds) == 1 {
		// rename file
		oldName := savedFileIds[0].Name
		index := strings.LastIndex(oldName, ".")
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files/"+savedFileIds[0].Id, http.MethodPatch, func(req *resty.Request) {
			req.SetBody(base.Json{
				"name": fmt.Sprintf("%s.%s", d.prettyName(name), oldName[index+1:]),
			})
		}, nil)

		if err != nil {
			utils.Log.Info("file rename error:", err)
			return dirId
		}

		// move file
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{savedFileIds[0].Id},
				"to": base.Json{
					"parent_id": parentDirId,
				},
			})
		}, nil)

		if err != nil {
			utils.Log.Info("move file error:", err)
			return dirId
		}

		// delete garbage file
		_, err = d.request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"ids": []string{dirId},
			})
		}, nil)
		if err != nil {
			utils.Log.Info("delete file error:", err)
			return parentDirId
		}

		time.Sleep(1 * time.Second)

		return parentDirId

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

		return dirId
	}

	return dirId

}

var _ driver.Driver = (*PikPak)(nil)
