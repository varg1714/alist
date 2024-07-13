package quark_share

import (
	"context"
	"errors"
	quark "github.com/alist-org/alist/v3/drivers/quark_uc"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"path/filepath"
	"strconv"
	"strings"
)

type QuarkShare struct {
	model.Storage
	Addition
	client *resty.Client
	conf   Conf
}

type Conf struct {
	ua      string
	referer string
	api     string
	pr      string
}

func (d *QuarkShare) Config() driver.Config {
	return config
}

func (d *QuarkShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *QuarkShare) Init(ctx context.Context) error {
	//op.MustSaveDriverStorage(d)
	d.client = resty.New().
		SetBaseURL("https://drive-pc.quark.cn").
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36").
		SetHeader("cookie", d.Cookie)

	return nil
}

func (d *QuarkShare) Drop(ctx context.Context) error {
	return nil
}

func (d *QuarkShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {

	return virtual_file.List(d.ID, dir, func(virtualFile model.VirtualFile, dir model.Obj) ([]model.Obj, error) {

		files, err := d.getShareFiles(ctx, virtualFile, filepath.Base(dir.GetPath()))
		if err != nil {
			return nil, err
		}

		return utils.SliceConvert(files, func(src FileObj) (model.Obj, error) {
			src.Path = filepath.Join(dir.GetPath(), src.GetID())
			return &src, nil
		})

	})

}

func (d *QuarkShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	fileObject, exist := file.(*FileObj)
	if !exist {
		return nil, errors.New("文件格式错误")
	}

	storage := op.GetBalancedStorage(d.QuarkDriverPath)
	quarkDriver, ok := storage.(*quark.QuarkOrUC)
	if !ok {
		return &model.Link{
			URL: "",
		}, nil
	}

	split := strings.Split(file.GetPath(), "/")
	virtualFile := db.QueryVirtualFilm(d.ID, split[0])

	transformFile, err := d.transformFile(virtualFile, *fileObject)
	if err != nil {
		return nil, err
	}
	if transformFile == "" {
		return nil, errors.New("文件转存失败")
	}

	link, err := quarkDriver.Link(ctx, &model.Object{ID: transformFile}, args)
	if err != nil {
		utils.Log.Infof("获取转存文件:%s的地址失败:%v", transformFile, err)
	}

	go func() {
		err2 := quarkDriver.Remove(ctx, &model.Object{ID: transformFile})
		if err2 != nil {
			utils.Log.Infof("文件:%s转存结果删除失败:%v", fileObject.GetName(), err2)
		} else {
			utils.Log.Infof("文件:%s转存结果已删除", fileObject.GetName())
		}
	}()

	return link, err
}

func (d *QuarkShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return virtual_file.MakeDir(d.ID, dirName)
}

func (d *QuarkShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotSupport
}

func (d *QuarkShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return virtual_file.Rename(d.ID, srcObj.GetPath(), srcObj.GetID(), newName)
}

func (d *QuarkShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotSupport
}

func (d *QuarkShare) Remove(ctx context.Context, obj model.Obj) error {
	return db.DeleteVirtualFile(strconv.Itoa(int(d.ID)), obj.GetName())
}

func (d *QuarkShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotSupport
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*QuarkShare)(nil)
