package baidu_share

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/drivers/baidu_netdisk"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
)

type BaiduShare struct {
	model.Storage
	Addition
	client *resty.Client
	info   struct {
		Root    string
		Seckey  string
		Shareid string
		Uk      string
	}
}

type ShareInfo struct {
	Errno int64 `json:"errno"`
	Data  struct {
		List [1]struct {
			Path string `json:"path"`
		} `json:"list"`
		Uk      json.Number `json:"uk"`
		Shareid json.Number `json:"shareid"`
		Seckey  string      `json:"seckey"`
	} `json:"data"`
}

func (d *BaiduShare) Config() driver.Config {
	return config
}

func (d *BaiduShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *BaiduShare) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	d.client = resty.New().
		SetBaseURL("https://pan.baidu.com").
		SetHeader("User-Agent", "netdisk").
		SetCookie(&http.Cookie{Name: "BDUSS", Value: d.BDUSS}).
		SetCookie(&http.Cookie{Name: "ndut_fmt"})

	var err error

	d.info.Root, d.info.Seckey, d.info.Shareid, d.info.Uk, err = d.getShareInfo(d.Surl, d.Pwd)

	return err
}

func (d *BaiduShare) getShareInfo(shareId, pwd string) (string, string, string, string, error) {

	shareToken, exist := shareTokenCache.Get(shareId)
	if exist {
		return path.Dir(shareToken.Data.List[0].Path), shareToken.Data.Seckey, shareToken.Data.Shareid.String(), shareToken.Data.Uk.String(), nil
	}

	respJson := ShareInfo{}

	resp, err := d.client.R().
		SetBody(url.Values{
			"pwd":      {pwd},
			"root":     {"1"},
			"shorturl": {shareId},
		}.Encode()).
		SetResult(&respJson).
		Post("share/wxlist?channel=weixin&version=2.2.2&clienttype=25&web=1")

	if err == nil {
		if resp.IsSuccess() && respJson.Errno == 0 {
			shareTokenCache.Set(shareId, respJson, cache.WithEx[ShareInfo](time.Second*time.Duration(30)))
			return path.Dir(respJson.Data.List[0].Path), respJson.Data.Seckey, respJson.Data.Shareid.String(), respJson.Data.Uk.String(), nil
		} else {
			err = fmt.Errorf(" %s; %s; ", resp.Status(), resp.Body())
			return "", "", "", "", err
		}
	}
	utils.Log.Error("百度云token获取错误", err)
	return "", "", "", "", err
}

func (d *BaiduShare) Drop(ctx context.Context) error {
	return nil
}

func (d *BaiduShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	// TODO return the files list, required

	return virtual_file.List(d.ID, dir, func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error) {

		reqDir := "/"
		split := strings.Split(dir.GetPath(), "/")

		if len(split) > 1 {

			parentDir := path.Join(split[1:]...)
			unescape, err2 := url.QueryUnescape(parentDir)
			if err2 != nil {
				utils.Log.Infof("url unescape error: %s", err2)
				unescape = parentDir
			}
			reqDir = path.Join(reqDir, unescape)
		}

		isRoot := "0"
		if reqDir == "/" {
			isRoot = "1"
		}

		objs := []model.Obj{}
		var err error
		var page uint64 = 1
		more := true
		for more && err == nil {
			respJson := struct {
				Errno int64 `json:"errno"`
				Data  struct {
					More bool `json:"has_more"`
					List []struct {
						Fsid  json.Number `json:"fs_id"`
						Isdir json.Number `json:"isdir"`
						Path  string      `json:"path"`
						Name  string      `json:"server_filename"`
						Mtime json.Number `json:"server_mtime"`
						Size  json.Number `json:"size"`
					} `json:"list"`
				} `json:"data"`
			}{}
			resp, e := d.client.R().
				SetBody(url.Values{
					"dir":      {reqDir},
					"num":      {"1000"},
					"order":    {"time"},
					"page":     {fmt.Sprint(page)},
					"pwd":      {virtualFile.SharePwd},
					"root":     {isRoot},
					"shorturl": {virtualFile.ShareID},
				}.Encode()).
				SetResult(&respJson).
				Post("share/wxlist?channel=weixin&version=2.2.2&clienttype=25&web=1")
			err = e
			if err == nil {
				if resp.IsSuccess() && respJson.Errno == 0 {
					page++
					more = respJson.Data.More
					for _, v := range respJson.Data.List {
						size, _ := v.Size.Int64()
						mtime, _ := v.Mtime.Int64()
						objs = append(objs, &model.Object{
							ID:       v.Fsid.String(),
							Path:     split[0] + v.Path,
							Name:     v.Name,
							Size:     size,
							Modified: time.Unix(mtime, 0),
							IsFolder: v.Isdir.String() == "1",
						})
					}
				} else {
					err = fmt.Errorf(" %s; %s; ", resp.Status(), resp.Body())
				}
			}
		}
		return objs, err

	})

}

func (d *BaiduShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	start := time.Now().UnixMilli()
	virtualFile := virtual_file.GetVirtualFile(d.ID, file.GetPath())

	_, secKey, shareId, uk, _ := d.getShareInfo(virtualFile.ShareID, virtualFile.SharePwd)

	link := &model.Link{Header: d.client.Header}
	signJson := struct {
		Errno int64 `json:"errno"`
		Extra struct {
			List []struct {
				From     string `json:"from"`
				FromFsId int    `json:"from_fs_id"`
				To       string `json:"to"`
				ToFsId   int    `json:"to_fs_id"`
			} `json:"list"`
		} `json:"extra"`
	}{}

	resp, err := d.client.R().
		SetQueryParams(map[string]string{
			"shareid": shareId,
			"from":    uk,
		}).
		SetBody(url.Values{
			"fsidlist": {fmt.Sprintf("[%s]", file.GetID())},
			"path":     {d.TransferPath},
		}.Encode()).
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		SetHeader("Referer", fmt.Sprintf("https://pan.baidu.com/s/%s", virtualFile.ShareID)).
		SetHeader("sekey", secKey).
		SetResult(&signJson).
		Post("share/transfer?channel=chunlei&web=1&app_id=250528&clienttype=0&ondup=newcopy")

	if err != nil {
		return link, err
	}

	if len(signJson.Extra.List) == 0 {
		utils.Log.Infof("文件转存失败:%s", resp.String())
		return nil, errors.New("文件转存失败")
	}

	storage := op.GetBalancedStorage(d.BaiDuDriverPath)
	baiduDrive, ok := storage.(*baidu_netdisk.BaiduNetdisk)
	if !ok {
		return link, nil
	}

	obj := &model.Object{
		ID: strconv.Itoa(signJson.Extra.List[0].ToFsId),
	}

	relLink, err := baiduDrive.Link(ctx, obj, args)
	utils.Log.Infof("文件转存与获取地址耗时：:[%d]ms", time.Now().UnixMilli()-start)

	if err != nil {
		return relLink, err
	}

	go func() {
		obj.Path = signJson.Extra.List[0].To
		err = baiduDrive.Remove(ctx, obj)
		if err != nil {
			utils.Log.Infof("清除文件:[%s]失败,错误原因:[%s]", file.GetName(), err.Error())
			return
		}
		utils.Log.Infof("清除文件:[%s]完毕", file.GetName())
	}()

	return relLink, err
}

func (d *BaiduShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return virtual_file.MakeDir(d.ID, parentDir, dirName)
}

func (d *BaiduShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return virtual_file.Move(d.ID, srcObj, dstDir)
}

func (d *BaiduShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return virtual_file.Rename(d.ID, srcObj.GetPath(), srcObj.GetID(), newName)
}

func (d *BaiduShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotSupport
}

func (d *BaiduShare) Remove(ctx context.Context, obj model.Obj) error {
	return virtual_file.DeleteVirtualFile(d.ID, obj)
}

func (d *BaiduShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotSupport
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*BaiduShare)(nil)
