package aliyundrive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dustinxie/ecc"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

func (d *AliDrive) createSession() error {
	state, ok := global.Load(d.UserID)
	if !ok {
		return fmt.Errorf("can't load user state, user_id: %s", d.UserID)
	}
	d.sign()
	state.retry++
	if state.retry > 3 {
		state.retry = 0
		return fmt.Errorf("createSession failed after three retries")
	}
	_, err, _ := d.request("https://api.aliyundrive.com/users/v1/users/device/create_session", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"deviceName":   "samsung",
			"modelName":    "SM-G9810",
			"nonce":        0,
			"pubKey":       PublicKeyToHex(&state.privateKey.PublicKey),
			"refreshToken": d.RefreshToken,
		})
	}, nil)
	if err == nil {
		state.retry = 0
	}
	return err
}

// func (d *AliDrive) renewSession() error {
// 	_, err, _ := d.request("https://api.aliyundrive.com/users/v1/users/device/renew_session", http.MethodPost, nil, nil)
// 	return err
// }

func (d *AliDrive) sign() {
	state, _ := global.Load(d.UserID)
	secpAppID := "5dde4e1bdf9e4966b387ba58f4b3fdc3"
	singdata := fmt.Sprintf("%s:%s:%s:%d", secpAppID, state.deviceID, d.UserID, 0)
	hash := sha256.Sum256([]byte(singdata))
	data, _ := ecc.SignBytes(state.privateKey, hash[:], ecc.RecID|ecc.LowerS)
	state.signature = hex.EncodeToString(data) //strconv.Itoa(state.nonce)
}

// do others that not defined in Driver interface

func (d *AliDrive) refreshToken() error {
	url := "https://auth.aliyundrive.com/v2/account/token"
	var resp base.TokenResp
	var e RespErr
	_, err := base.RestyClient.R().
		//ForceContentType("application/json").
		SetBody(base.Json{"refresh_token": d.RefreshToken, "grant_type": "refresh_token"}).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return err
	}
	if e.Code != "" {
		return fmt.Errorf("failed to refresh token: %s", e.Message)
	}
	if resp.RefreshToken == "" {
		return errors.New("failed to refresh token: refresh token is empty")
	}
	d.RefreshToken, d.AccessToken = resp.RefreshToken, resp.AccessToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *AliDrive) request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error, RespErr) {
	req := base.RestyClient.R()
	state, ok := global.Load(d.UserID)
	if !ok {
		if url == "https://api.aliyundrive.com/v2/user/get" {
			state = &State{}
		} else {
			return nil, fmt.Errorf("can't load user state, user_id: %s", d.UserID), RespErr{}
		}
	}
	req.SetHeaders(map[string]string{
		"Authorization": "Bearer\t" + d.AccessToken,
		"content-type":  "application/json",
		"origin":        "https://www.aliyundrive.com",
		"Referer":       "https://aliyundrive.com/",
		"X-Signature":   state.signature,
		"x-request-id":  uuid.NewString(),
		"X-Canary":      "client=Android,app=adrive,version=v4.1.0",
		"X-Device-Id":   state.deviceID,
	})
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err, e
	}
	if e.Code != "" {
		switch e.Code {
		case "AccessTokenInvalid":
			err = d.refreshToken()
			if err != nil {
				return nil, err, e
			}
		case "DeviceSessionSignatureInvalid":
			err = d.createSession()
			if err != nil {
				return nil, err, e
			}
		default:
			return nil, errors.New(e.Message), e
		}
		return d.request(url, method, callback, resp)
	} else if res.IsError() {
		return nil, errors.New("bad status code " + res.Status()), e
	}
	return res.Body(), nil, e
}

func (d *AliDrive) getFiles(fileId string) ([]File, error) {
	marker := "first"
	res := make([]File, 0)
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		var resp Files
		data := base.Json{
			"drive_id":                d.DriveId,
			"fields":                  "*",
			"image_thumbnail_process": "image/resize,w_400/format,jpeg",
			"image_url_process":       "image/resize,w_1920/format,jpeg",
			"limit":                   200,
			"marker":                  marker,
			"order_by":                d.OrderBy,
			"order_direction":         d.OrderDirection,
			"parent_file_id":          fileId,
			"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
			"url_expire_sec":          14400,
		}
		_, err, _ := d.request("https://api.aliyundrive.com/adrive/v3/file/list", http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, &resp)

		if err != nil {
			return nil, err
		}
		marker = resp.NextMarker
		res = append(res, resp.Items...)
	}
	return res, nil
}

func (d *AliDrive) getShareFiles(ctx context.Context, shareId string, parentFileId string, appendSubFolder bool) ([]File, error) {

	err := limiter.WaitN(ctx, 1)
	if err != nil {
		return nil, err
	}

	token, err := d.getShareToken(shareId)
	if err != nil {
		return nil, err
	}

	res := make([]File, 0)

	firstAccess := true
	queue := generic.NewQueue[string]()
	queue.Push(parentFileId)

	for queue.Len() > 0 {

		tempParentFileId := queue.Pop()
		if !firstAccess {
			time.Sleep(250 * time.Millisecond)
		}

		marker := "first"
		for marker != "" {
			if marker == "first" {
				marker = ""
			}
			var resp Files
			data := base.Json{
				"share_id":                shareId,
				"parent_file_id":          tempParentFileId,
				"limit":                   200,
				"image_thumbnail_process": "image/resize,w_256/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg/interlace,1",
				"video_thumbnail_process": "video/snapshot,t_1000,f_jpg,ar_auto,w_256",
				"order_by":                "name",
				"order_direction":         "ASC",
				"marker":                  marker,
			}
			_, err, _ = d.request("https://api.aliyundrive.com/adrive/v2/file/list_by_share", http.MethodPost, func(req *resty.Request) {
				req.SetBody(data)
				req.SetHeader("x-share-token", token)
			}, &resp)

			firstAccess = false

			if err != nil {
				return nil, err
			}
			marker = resp.NextMarker

			for _, item := range resp.Items {
				if item.Size/(1024*1024) > 100 || (item.Type == "folder" && !appendSubFolder) {
					res = append(res, item)
				}

				if item.Type == "folder" && appendSubFolder {
					utils.Log.Infof("递归遍历子文件夹：[%s]", item.Name)
					queue.Push(item.FileId)
				}

			}

		}
	}

	return res, nil
}

func (d *AliDrive) getShareToken(shareId string) (string, error) {

	shareToken, exist := shareTokenCache.Get(shareId)
	if exist {
		return shareToken, nil
	}

	var shareResp ShareResp
	data := base.Json{
		"share_id": shareId,
	}
	_, err, _ := d.request("https://api.aliyundrive.com/v2/share_link/get_share_token", http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, &shareResp)

	if err != nil {
		return "", err
	}

	shareTokenCache.Set(shareId, shareResp.ShareToken, cache.WithEx[string](time.Minute*time.Duration(100)))

	return shareResp.ShareToken, nil
}

func (d *AliDrive) batch(srcId, dstId string, url string) error {
	res, err, _ := d.request("https://api.aliyundrive.com/v3/batch", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"requests": []base.Json{
				{
					"headers": base.Json{
						"Content-Type": "application/json",
					},
					"method": "POST",
					"id":     srcId,
					"body": base.Json{
						"drive_id":          d.DriveId,
						"file_id":           srcId,
						"to_drive_id":       d.DriveId,
						"to_parent_file_id": dstId,
					},
					"url": url,
				},
			},
			"resource": "file",
		})
	}, nil)
	if err != nil {
		return err
	}
	status := utils.Json.Get(res, "responses", 0, "status").ToInt()
	if status < 400 && status >= 100 {
		return nil
	}
	return errors.New(string(res))
}

func (d *AliDrive) SaveShare(shareId string, srcId string, dstId string) (string, error) {

	token, err := d.getShareToken(shareId)
	if err != nil {
		utils.Log.Warn("获取shareToken错误:", token, err)
		return "", err
	}

	var shareSaveResp ShareSaveResp

	_, err, _ = d.request("https://api.aliyundrive.com/v3/batch", http.MethodPost, func(req *resty.Request) {
		req.SetHeader("x-share-token", token)
		req.SetBody(base.Json{
			"requests": []base.Json{
				{
					"headers": base.Json{
						"Content-Type": "application/json",
					},
					"method": "POST",
					"id":     srcId,
					"body": base.Json{
						"file_id":           srcId,
						"share_id":          shareId,
						"auto_rename":       true,
						"to_parent_file_id": dstId,
						"drive_id":          d.DriveId,
						"to_drive_id":       d.DriveId,
					},
					"url": "/file/copy",
				},
			},
			"resource": "file",
		})
	}, &shareSaveResp)
	if err != nil || len(shareSaveResp.Responses) == 0 {
		utils.Log.Warn("转存错误:", shareSaveResp, err)
		return "", err
	}
	return shareSaveResp.Responses[0].Body.FileID, err
}

func replace(test model.VirtualFile, index int) bool {

	if test.SourceName == "" {
		return false
	}

	if index >= test.Start && ((index <= test.End) || (test.End == -1)) {
		return true
	} else if test.Start == -1 && test.End == -1 {
		return true
	}

	return false

}
