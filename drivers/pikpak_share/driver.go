package pikpak_share

import (
	"context"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"net/http"
	"path/filepath"
	"strconv"
)

type PikPakShare struct {
	model.Storage
	Addition
	RefreshToken  string
	AccessToken   string
	PassCodeToken string
}

func (d *PikPakShare) Config() driver.Config {
	return config
}

func (d *PikPakShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPakShare) Init(ctx context.Context) error {
	err := d.login()
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

		files, err := d.getFiles(virtualFile, filepath.Dir(filepath.Base(dir.GetPath())))
		if err != nil {
			return nil, err
		}

		return utils.SliceConvert(files, func(src File) (model.Obj, error) {
			obj := fileToObj(src)
			obj.Path = virtualFile.Name
			return obj, nil
		})

	})

}

func (d *PikPakShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {

	var resp ShareResp

	virtualFile := db.QueryVirtualFilm(d.ID, file.GetPath())
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
	_, err = d.request("https://api-drive.mypikpak.com/drive/v1/share/file_info", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return nil, err
	}
	link := model.Link{
		URL: resp.FileInfo.WebContentLink,
	}
	return &link, nil
}

func (d *PikPakShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {

	return virtual_file.MakeDir(d.ID, dirName)

}

func (d *PikPakShare) Remove(ctx context.Context, obj model.Obj) error {
	return db.DeleteVirtualFile(strconv.Itoa(int(d.ID)), obj.GetName())
}

var _ driver.Driver = (*PikPakShare)(nil)
