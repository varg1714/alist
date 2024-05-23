package pikpak_share

import (
	"errors"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

var shareTokenCache = cache.NewMemCache(cache.WithShards[string](128))

func (d *PikPakShare) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()

	token, err := d.oauth2Token.Token()
	if err != nil {
		return nil, err
	}
	req.SetAuthScheme(token.TokenType).SetAuthToken(token.AccessToken)

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
