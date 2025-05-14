package pikpak_share

import (
	"context"
	"github.com/alist-org/alist/v3/internal/op"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"path/filepath"
)

type PikPakShare struct {
	model.Storage
	Addition
	*Common
	PassCodeToken string
}

func (d *PikPakShare) Config() driver.Config {
	return config
}

func (d *PikPakShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPakShare) Init(ctx context.Context) error {
	if d.Common == nil {
		d.Common = &Common{
			DeviceID:  utils.GetMD5EncodeStr(d.Addition.ShareId + d.Addition.SharePwd + time.Now().String()),
			UserAgent: "",
			RefreshCTokenCk: func(token string) {
				d.Common.CaptchaToken = token
				op.MustSaveDriverStorage(d)
			},
		}
	}

	if d.Addition.DeviceID != "" {
		d.SetDeviceID(d.Addition.DeviceID)
	} else {
		d.Addition.DeviceID = d.Common.DeviceID
		op.MustSaveDriverStorage(d)
	}

	if d.Platform == "android" {
		d.ClientID = AndroidClientID
		d.ClientSecret = AndroidClientSecret
		d.ClientVersion = AndroidClientVersion
		d.PackageName = AndroidPackageName
		d.Algorithms = AndroidAlgorithms
		d.UserAgent = BuildCustomUserAgent(d.GetDeviceID(), AndroidClientID, AndroidPackageName, AndroidSdkVersion, AndroidClientVersion, AndroidPackageName, "")
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
		d.UserAgent = "MainWindow Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) PikPak/2.6.11.4955 Chrome/100.0.4896.160 Electron/18.3.15 Safari/537.36"
	}

	// 获取CaptchaToken
	err := d.RefreshCaptchaToken(GetAction(http.MethodGet, "https://api-drive.mypikpak.net/drive/v1/share:batch_file_info"), "")
	if err != nil {
		return err
	}

	return nil
}

func (d *PikPakShare) Drop(ctx context.Context) error {
	return nil
}

func (d *PikPakShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	return virtual_file.List(d.ID, dir, func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error) {

		files, err := d.getFiles(virtualFile, filepath.Base(dir.GetPath()))
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

func (d *PikPakShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	var resp ShareResp

	virtualFile := virtual_file.GetVirtualFile(d.ID, file.GetPath())

	sharePassToken, err := d.getSharePassToken(virtualFile)

	if err != nil {
		utils.Log.Warnf("share token获取错误, share Id:[%s],error:[%s]", file.GetPath(), err.Error())
		return nil, err
	}

	query := map[string]string{
		"share_id":        virtualFile.ShareID,
		"file_id":         file.GetID(),
		"pass_code_token": sharePassToken,
	}
	_, err = d.request("https://api-drive.mypikpak.net/drive/v1/share/file_info", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return nil, err
	}

	downloadUrl := resp.FileInfo.WebContentLink
	if downloadUrl == "" && len(resp.FileInfo.Medias) > 0 {
		// 使用转码后的链接
		if d.Addition.UseTransCodingAddress && len(resp.FileInfo.Medias) > 1 {
			downloadUrl = resp.FileInfo.Medias[1].Link.Url
		} else {
			downloadUrl = resp.FileInfo.Medias[0].Link.Url
		}

	}

	return &model.Link{
		URL: downloadUrl,
	}, nil
}

func (d *PikPakShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	return virtual_file.MakeDir(d.ID, parentDir, dirName)

}

func (d *PikPakShare) Remove(ctx context.Context, obj model.Obj) error {
	return virtual_file.DeleteVirtualFile(d.ID, obj)
}

func (d *PikPakShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return virtual_file.Move(d.ID, srcObj, dstDir)
}

var _ driver.Driver = (*PikPakShare)(nil)
