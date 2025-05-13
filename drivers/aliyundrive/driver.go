package aliyundrive

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/aliyundrive_open"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/op"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type AliDrive struct {
	model.Storage
	Addition
	AccessToken string
	cron        *cron.Cron
	DriveId     string
	UserID      string
}

func (d *AliDrive) Config() driver.Config {
	return config
}

func (d *AliDrive) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AliDrive) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	err := d.refreshToken()
	if err != nil {
		return err
	}
	// get driver id
	res, err, _ := d.request("https://api.aliyundrive.com/v2/user/get", http.MethodPost, nil, nil)
	if err != nil {
		return err
	}
	d.DriveId = utils.Json.Get(res, "default_drive_id").ToString()
	d.UserID = utils.Json.Get(res, "user_id").ToString()
	d.cron = cron.NewCron(time.Hour * 2)
	d.cron.Do(func() {
		err := d.refreshToken()
		if err != nil {
			log.Errorf("%+v", err)
		}
	})
	if global.Has(d.UserID) {
		return nil
	}
	// init deviceID
	deviceID := utils.HashData(utils.SHA256, []byte(d.UserID))
	// init privateKey
	privateKey, _ := NewPrivateKeyFromHex(deviceID)
	state := State{
		privateKey: privateKey,
		deviceID:   deviceID,
	}
	// store state
	global.Store(d.UserID, &state)
	// init signature
	d.sign()
	return nil
}

func (d *AliDrive) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *AliDrive) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	return virtual_file.List(d.ID, dir, func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error) {

		files, err := d.getShareFiles(ctx, virtualFile, dir)
		if err != nil {
			return nil, err
		}

		return utils.SliceConvert(files, func(src File) (model.Obj, error) {
			obj := fileToObj(src)
			obj.Path = filepath.Join(dir.GetPath(), obj.GetID())
			return obj, nil
		})

	})

}

func (d *AliDrive) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	storage := op.GetBalancedStorage(d.StoragePath)
	aliDrive, ok := storage.(*aliyundrive_open.AliyundriveOpen)
	if !ok {
		return &model.Link{
			URL: "",
		}, nil
	}

	// 转存
	start := time.Now().UnixMilli()

	utils.Log.Infof("开始转存文件:[%s]", file.GetName())

	virtualFile := virtual_file.GetVirtualFile(d.ID, file.GetPath())

	shareFileId, err := d.SaveShare(virtualFile.ShareID, file.GetID(), d.TempFolderPath)
	if err != nil {
		return &model.Link{
			URL: "",
		}, nil
	}

	obj := &model.Object{
		ID: shareFileId,
	}
	link, err := aliDrive.Link(ctx, obj, args)

	utils.Log.Infof("文件转存与获取地址耗时：:[%d]ms\n", time.Now().UnixMilli()-start)

	if err != nil {
		return link, err
	}

	go func() {
		err = aliDrive.Remove(ctx, obj)
		if err != nil {
			utils.Log.Infof("清除文件:[%s]失败,错误原因:[%s]\n", file.GetName(), err.Error())
			return
		}
		utils.Log.Infof("清除文件:[%s]完毕\n", file.GetName())
	}()

	return link, err

}

func (d *AliDrive) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	return virtual_file.MakeDir(d.ID, parentDir, dirName)

}

func (d *AliDrive) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	err := d.batch(srcObj.GetID(), dstDir.GetID(), "/file/move")
	return err
}

func (d *AliDrive) Rename(ctx context.Context, srcObj model.Obj, newName string) error {

	return virtual_file.Rename(d.ID, srcObj.GetPath(), srcObj.GetID(), newName)

}

func (d *AliDrive) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	err := d.batch(srcObj.GetID(), dstDir.GetID(), "/file/copy")
	return err
}

func (d *AliDrive) Remove(ctx context.Context, obj model.Obj) error {
	return virtual_file.DeleteVirtualFile(d.ID, obj)
}

func (d *AliDrive) Put(ctx context.Context, dstDir model.Obj, streamer model.FileStreamer, up driver.UpdateProgress) error {
	file := stream.FileStream{
		Obj:      streamer,
		Reader:   streamer,
		Mimetype: streamer.GetMimetype(),
	}
	const DEFAULT int64 = 10485760
	var count = int(math.Ceil(float64(streamer.GetSize()) / float64(DEFAULT)))

	partInfoList := make([]base.Json, 0, count)
	for i := 1; i <= count; i++ {
		partInfoList = append(partInfoList, base.Json{"part_number": i})
	}
	reqBody := base.Json{
		"check_name_mode": "overwrite",
		"drive_id":        d.DriveId,
		"name":            file.GetName(),
		"parent_file_id":  dstDir.GetID(),
		"part_info_list":  partInfoList,
		"size":            file.GetSize(),
		"type":            "file",
	}

	var localFile *os.File
	if fileStream, ok := file.Reader.(*stream.FileStream); ok {
		localFile, _ = fileStream.Reader.(*os.File)
	}
	if d.RapidUpload {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		_, err := utils.CopyWithBufferN(buf, file, 1024)
		if err != nil {
			return err
		}
		reqBody["pre_hash"] = utils.HashData(utils.SHA1, buf.Bytes())
		if localFile != nil {
			if _, err := localFile.Seek(0, io.SeekStart); err != nil {
				return err
			}
		} else {
			// 把头部拼接回去
			file.Reader = struct {
				io.Reader
				io.Closer
			}{
				Reader: io.MultiReader(buf, file),
				Closer: &file,
			}
		}
	} else {
		reqBody["content_hash_name"] = "none"
		reqBody["proof_version"] = "v1"
	}

	var resp UploadResp
	_, err, e := d.request("https://api.aliyundrive.com/adrive/v2/file/createWithFolders", http.MethodPost, func(req *resty.Request) {
		req.SetBody(reqBody)
	}, &resp)

	if err != nil && e.Code != "PreHashMatched" {
		return err
	}

	if d.RapidUpload && e.Code == "PreHashMatched" {
		delete(reqBody, "pre_hash")
		h := sha1.New()
		if localFile != nil {
			if err = utils.CopyWithCtx(ctx, h, localFile, 0, nil); err != nil {
				return err
			}
			if _, err = localFile.Seek(0, io.SeekStart); err != nil {
				return err
			}
		} else {
			tempFile, err := os.CreateTemp(conf.Conf.TempDir, "file-*")
			if err != nil {
				return err
			}
			defer func() {
				_ = tempFile.Close()
				_ = os.Remove(tempFile.Name())
			}()
			if err = utils.CopyWithCtx(ctx, io.MultiWriter(tempFile, h), file, 0, nil); err != nil {
				return err
			}
			localFile = tempFile
		}
		reqBody["content_hash"] = hex.EncodeToString(h.Sum(nil))
		reqBody["content_hash_name"] = "sha1"
		reqBody["proof_version"] = "v1"

		/*
			js 隐性转换太坑不知道有没有bug
			var n = e.access_token，
			r = new BigNumber('0x'.concat(md5(n).slice(0, 16)))，
			i = new BigNumber(t.file.size)，
			o = i ? r.mod(i) : new gt.BigNumber(0);
			(t.file.slice(o.toNumber(), Math.min(o.plus(8).toNumber(), t.file.size)))
		*/
		buf := make([]byte, 8)
		r, _ := new(big.Int).SetString(utils.GetMD5EncodeStr(d.AccessToken)[:16], 16)
		i := new(big.Int).SetInt64(file.GetSize())
		o := new(big.Int).SetInt64(0)
		if file.GetSize() > 0 {
			o = r.Mod(r, i)
		}
		n, _ := io.NewSectionReader(localFile, o.Int64(), 8).Read(buf[:8])
		reqBody["proof_code"] = base64.StdEncoding.EncodeToString(buf[:n])

		_, err, e := d.request("https://api.aliyundrive.com/adrive/v2/file/createWithFolders", http.MethodPost, func(req *resty.Request) {
			req.SetBody(reqBody)
		}, &resp)
		if err != nil && e.Code != "PreHashMatched" {
			return err
		}
		if resp.RapidUpload {
			return nil
		}
		// 秒传失败
		if _, err = localFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
		file.Reader = localFile
	}

	rateLimited := driver.NewLimitedUploadStream(ctx, file)
	for i, partInfo := range resp.PartInfoList {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}
		url := partInfo.UploadUrl
		if d.InternalUpload {
			url = partInfo.InternalUploadUrl
		}
		req, err := http.NewRequest("PUT", url, io.LimitReader(rateLimited, DEFAULT))
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		if count > 0 {
			up(float64(i) * 100 / float64(count))
		}
	}
	var resp2 base.Json
	_, err, e = d.request("https://api.aliyundrive.com/v2/file/complete", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":  d.DriveId,
			"file_id":   resp.FileId,
			"upload_id": resp.UploadId,
		})
	}, &resp2)
	if err != nil && e.Code != "PreHashMatched" {
		return err
	}
	if resp2["file_id"] == resp.FileId {
		return nil
	}
	return fmt.Errorf("%+v", resp2)
}

func (d *AliDrive) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	return nil, nil
}

var _ driver.Driver = (*AliDrive)(nil)
