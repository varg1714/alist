package quark_share

import (
	"context"
	"errors"
	_189pc "github.com/OpenListTeam/OpenList/v4/drivers/189pc"
	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/drivers/virtual_file"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/go-resty/resty/v2"
	"path/filepath"
)

type Cloud189Share struct {
	model.Storage
	Addition
	client *resty.Client
}

func (d *Cloud189Share) Config() driver.Config {
	return config
}

func (d *Cloud189Share) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Cloud189Share) Init(ctx context.Context) error {
	//op.MustSaveDriverStorage(d)
	d.client = base.NewRestyClient().SetHeaders(map[string]string{
		"Accept":  "application/json;charset=UTF-8",
		"Referer": "https://cloud.189.cn",
	})

	return nil
}

func (d *Cloud189Share) Drop(ctx context.Context) error {
	return nil
}

func (d *Cloud189Share) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	return virtual_file.List(d.ID, dir, func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error) {

		files, err := d.getShareFiles(ctx, virtualFile, dir)
		if err != nil {
			return nil, err
		}

		return utils.SliceConvert(files, func(src FileObj) (model.Obj, error) {
			src.Path = filepath.Join(dir.GetPath(), src.GetID())
			return &src, nil
		})

	})

}

func (d *Cloud189Share) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	err := limiter.WaitN(ctx, 1)
	if err != nil {
		return nil, err
	}

	fileObject, exist := file.(*FileObj)
	if !exist {
		return nil, errors.New("文件格式错误")
	}

	storage := op.GetBalancedStorage(d.Cloud189DriverPath)
	cloud189PC, ok := storage.(*_189pc.Cloud189PC)
	if !ok {
		return &model.Link{
			URL: "",
		}, nil
	}

	virtualFile := virtual_file.GetVirtualFile(d.ID, file.GetPath())
	shareInfo, err := d.getShareInfo(virtualFile.ShareID, virtualFile.SharePwd)
	if err != nil {
		return nil, err
	}

	transfer, err := cloud189PC.Transfer(ctx, shareInfo.ShareId, fileObject.ID, fileObject.oldName)
	if transfer != nil && transfer.URL != "" {
		expirationTime := GetExpirationTime(transfer.URL)
		transfer.Expiration = &expirationTime
	}

	return transfer, err

}

func (d *Cloud189Share) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return virtual_file.MakeDir(d.ID, parentDir, dirName)
}

func (d *Cloud189Share) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return virtual_file.Move(d.ID, srcObj, dstDir)
}

func (d *Cloud189Share) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return virtual_file.Rename(d.ID, srcObj.GetPath(), srcObj.GetID(), newName)
}

func (d *Cloud189Share) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotSupport
}

func (d *Cloud189Share) Remove(ctx context.Context, obj model.Obj) error {
	return virtual_file.DeleteVirtualFile(d.ID, obj)
}

func (d *Cloud189Share) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotSupport
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Cloud189Share)(nil)
