package pikpak_share

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/drivers/virtual_file"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"path/filepath"
	"strconv"
	"strings"
	"golang.org/x/oauth2"
)

type PikPakShare struct {
	model.Storage
	Addition
	oauth2Token   oauth2.TokenSource
	PassCodeToken string
}

func (d *PikPakShare) Config() driver.Config {
	return config
}

func (d *PikPakShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *PikPakShare) Init(ctx context.Context) error {
	if d.ClientID == "" || d.ClientSecret == "" {
		d.ClientID = "YNxT9w7GMdWvEOKa"
		d.ClientSecret = "dbw2OtmVEeuUvIptb1Coyg"
	}

	withClient := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, oauth2.HTTPClient, base.HttpClient)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     d.ClientID,
		ClientSecret: d.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://user.mypikpak.com/v1/auth/signin",
			TokenURL:  "https://user.mypikpak.com/v1/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	oauth2Token, err := oauth2Config.PasswordCredentialsToken(withClient(ctx), d.Username, d.Password)
	if err != nil {
		return err
	}
	d.oauth2Token = oauth2Config.TokenSource(withClient(context.Background()), oauth2Token)

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

	split := strings.Split(file.GetPath(), "/")
	virtualFile := db.QueryVirtualFilm(d.ID, split[0])

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

	downloadUrl := resp.FileInfo.WebContentLink
	if downloadUrl == "" && len(resp.FileInfo.Medias) > 0 {
		downloadUrl = resp.FileInfo.Medias[0].Link.Url
	}

	link := model.Link{
		URL: downloadUrl,
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
