package pikpak_share

import (
	"errors"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

var shareTokenCache = cache.NewMemCache(cache.WithShards[string](128))

func (d *PikPakShare) login() error {
	url := "https://user.mypikpak.com/v1/auth/signin"
	var e RespErr
	res, err := base.RestyClient.R().SetError(&e).SetBody(base.Json{
		"captcha_token": "",
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"username":      d.Username,
		"password":      d.Password,
	}).Post(url)
	if err != nil {
		return err
	}
	if e.ErrorCode != 0 {
		return errors.New(e.Error)
	}
	data := res.Body()
	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
	return nil
}

func (d *PikPakShare) refreshToken() error {
	url := "https://user.mypikpak.com/v1/auth/token"
	var e RespErr
	res, err := base.RestyClient.R().SetError(&e).
		SetHeader("user-agent", "").SetBody(base.Json{
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"grant_type":    "refresh_token",
		"refresh_token": d.RefreshToken,
	}).Post(url)
	if err != nil {
		d.Status = err.Error()
		op.MustSaveDriverStorage(d)
		return err
	}
	if e.ErrorCode != 0 {
		if e.ErrorCode == 4126 {
			// refresh_token invalid, re-login
			return d.login()
		}
		d.Status = e.Error
		op.MustSaveDriverStorage(d)
		return errors.New(e.Error)
	}
	data := res.Body()
	d.Status = "work"
	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *PikPakShare) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if e.ErrorCode != 0 {
		if e.ErrorCode == 16 {
			// login / refresh token
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(url, method, callback, resp)
		}
		return nil, errors.New(e.Error)
	}
	return res.Body(), nil
}

func (d *PikPakShare) getSharePassToken(virtualFile model.VirtualFile) (string, error) {

	if virtualFile.SharePwd == "" {
		return "", nil
	}

	shareToken, exist := shareTokenCache.Get(virtualFile.ShareID)
	if exist {
		return shareToken, nil
	}

	query := map[string]string{
		"share_id":       virtualFile.ShareID,
		"pass_code":      virtualFile.SharePwd,
		"thumbnail_size": "SIZE_LARGE",
		"limit":          "100",
	}
	var resp ShareResp
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/share", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return "", err
	}

	shareTokenCache.Set(virtualFile.ShareID, resp.PassCodeToken, cache.WithEx[string](time.Minute*time.Duration(100)))

	return resp.PassCodeToken, nil
}

func (d *PikPakShare) getFiles(virtualFile model.VirtualFile, parentId string) ([]File, error) {

	sharePassToken, err := d.getSharePassToken(virtualFile)
	if err != nil {
		utils.Log.Warnf("share token获取错误, share Id:[%s],error:[%s]", virtualFile.ShareID, err.Error())
		return nil, err
	}

	res := make([]File, 0)
	pageToken := "first"
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}

		query := map[string]string{
			"parent_id":       parentId,
			"share_id":        virtualFile.ShareID,
			"thumbnail_size":  "SIZE_LARGE",
			"with_audit":      "true",
			"limit":           "100",
			"filters":         `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			"page_token":      pageToken,
			"pass_code_token": sharePassToken,
		}
		var resp ShareResp
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/share/detail", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		if resp.ShareStatus != "OK" {
			return nil, errors.New(resp.ShareStatusText)
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}

func (d *PikPakShare) aggeFiles(virtualFile model.VirtualFile) ([]File, error) {

	res := make([]File, 0)

	firstAccess := true
	queue := generic.NewQueue[string]()
	queue.Push(virtualFile.ParentDir)

	for queue.Len() > 0 {

		tempParentFileId := queue.Pop()
		if !firstAccess {
			time.Sleep(100 * time.Millisecond)
		}

		files, err := d.getFiles(virtualFile, tempParentFileId)
		firstAccess = false

		if err != nil {
			return res, err
		}

		for _, item := range files {

			size, err := strconv.ParseInt(item.Size, 10, 64)
			if err != nil {
				utils.Log.Info("convert file size error:", err)
				return res, err
			}

			if size/(1024*1024) >= virtualFile.MinFileSize || (item.Kind == "drive#folder" && !virtualFile.AppendSubFolder) {
				res = append(res, item)
			}

			if item.Kind == "drive#folder" && virtualFile.AppendSubFolder {
				utils.Log.Infof("递归遍历子文件夹：[%s]", item.Name)
				queue.Push(item.Id)
			}

		}
	}

	return res, nil

}
