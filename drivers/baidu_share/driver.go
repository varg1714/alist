package baidu_share

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"io"
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
			shareTokenCache.Set(shareId, respJson, cache.WithEx[ShareInfo](time.Hour*time.Duration(7)))
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

	results := make([]model.Obj, 0)

	dirName := dir.GetName()
	utils.Log.Infof("list file:[%s]\n", dirName)

	virtualFilms := db.QueryVirtualFilms(strconv.Itoa(int(d.ID)))

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
				},
			})
		}
		return results, nil
	}

	split := strings.Split(dir.GetPath(), "/")
	virtualFile, exist := virtualFilms[split[1]]

	if !exist {
		// 分享文件夹
		return results, nil
	}

	var reqDir string
	if len(split) == 2 {
		reqDir = "/"
	} else {
		reqDir = "/" + strings.Join(split[2:], "/")
	}

	isRoot := "0"
	if reqDir == d.RootFolderPath {
		reqDir = path.Join(virtualFile.ParentDir, reqDir)
	}
	if reqDir == virtualFile.ParentDir || reqDir == "/" {
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
						Path:     "/" + split[1] + v.Path,
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
}

func (d *BaiduShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	// TODO return link of file, required

	split := strings.Split(file.GetPath(), "/")
	virtualFile := db.QueryVirtualFilm(d.ID, split[1])

	_, secKey, shareId, uk, _ := d.getShareInfo(virtualFile.ShareID, virtualFile.SharePwd)

	link := model.Link{Header: d.client.Header}
	sign := ""
	stamp := ""
	signJson := struct {
		Errno int64 `json:"errno"`
		Data  struct {
			Stamp json.Number `json:"timestamp"`
			Sign  string      `json:"sign"`
		} `json:"data"`
	}{}
	resp, err := d.client.R().
		SetQueryParam("surl", virtualFile.ShareID).
		SetResult(&signJson).
		Get("share/tplconfig?fields=sign,timestamp&channel=chunlei&web=1&app_id=250528&clienttype=0")
	if err == nil {
		if resp.IsSuccess() && signJson.Errno == 0 {
			stamp = signJson.Data.Stamp.String()
			sign = signJson.Data.Sign
		} else {
			err = fmt.Errorf(" %s; %s; ", resp.Status(), resp.Body())
		}
	}
	if err == nil {
		respJson := struct {
			Errno int64 `json:"errno"`
			List  [1]struct {
				Dlink string `json:"dlink"`
			} `json:"list"`
		}{}
		resp, err = d.client.R().
			SetQueryParam("sign", sign).
			SetQueryParam("timestamp", stamp).
			SetBody(url.Values{
				"encrypt":   {"0"},
				"extra":     {fmt.Sprintf(`{"sekey":"%s"}`, secKey)},
				"fid_list":  {fmt.Sprintf("[%s]", file.GetID())},
				"primaryid": {shareId},
				"product":   {"share"},
				"type":      {"nolimit"},
				"uk":        {uk},
			}.Encode()).
			SetResult(&respJson).
			Post("api/sharedownload?app_id=250528&channel=chunlei&clienttype=12&web=1")
		if err == nil {
			if resp.IsSuccess() && respJson.Errno == 0 && respJson.List[0].Dlink != "" {
				link.URL = respJson.List[0].Dlink
			} else {
				err = fmt.Errorf(" %s; %s; ", resp.Status(), resp.Body())
			}
		}
		if err == nil {
			resp, err = d.client.R().
				SetDoNotParseResponse(true).
				Get(link.URL)
			if err == nil {
				defer resp.RawBody().Close()
				if resp.IsError() {
					byt, _ := io.ReadAll(resp.RawBody())
					err = fmt.Errorf(" %s; %s; ", resp.Status(), byt)
				}
			}
		}
	}
	return &link, err
}

func (d *BaiduShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return virtual_file.MakeDir(d.ID, dirName)
}

func (d *BaiduShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotSupport
}

func (d *BaiduShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj, optional
	return errs.NotSupport
}

func (d *BaiduShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotSupport
}

func (d *BaiduShare) Remove(ctx context.Context, obj model.Obj) error {
	return db.DeleteVirtualFile(strconv.Itoa(int(d.ID)), obj.GetName())
}

func (d *BaiduShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotSupport
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*BaiduShare)(nil)
