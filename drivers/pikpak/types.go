package pikpak

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
)

type RespErr struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

type Files struct {
	Files         []File `json:"files"`
	NextPageToken string `json:"next_page_token"`
}

type File struct {
	Id             string    `json:"id"`
	Kind           string    `json:"kind"`
	Name           string    `json:"name"`
	CreatedTime    time.Time `json:"created_time"`
	ModifiedTime   time.Time `json:"modified_time"`
	Hash           string    `json:"hash"`
	Size           string    `json:"size"`
	ThumbnailLink  string    `json:"thumbnail_link"`
	WebContentLink string    `json:"web_content_link"`
	Medias         []Media   `json:"medias"`
	Params         struct {
		Duration     string `json:"duration"`
		Height       string `json:"height"`
		PlatformIcon string `json:"platform_icon"`
		URL          string `json:"url"`
		Width        string `json:"width"`
	} `json:"params"`
}

func fileToObj(f File) *model.ObjThumb {
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.Name,
			Size:     size,
			Ctime:    f.CreatedTime,
			Modified: f.ModifiedTime,
			IsFolder: f.Kind == "drive#folder",
			HashInfo: utils.NewHashInfo(hash_extend.GCID, f.Hash),
		},
		Thumbnail: model.Thumbnail{
			Thumbnail: f.ThumbnailLink,
		},
	}
}

type Media struct {
	MediaId   string `json:"media_id"`
	MediaName string `json:"media_name"`
	Video     struct {
		Height     int    `json:"height"`
		Width      int    `json:"width"`
		Duration   int    `json:"duration"`
		BitRate    int    `json:"bit_rate"`
		FrameRate  int    `json:"frame_rate"`
		VideoCodec string `json:"video_codec"`
		AudioCodec string `json:"audio_codec"`
		VideoType  string `json:"video_type"`
	} `json:"video"`
	Link struct {
		Url    string    `json:"url"`
		Token  string    `json:"token"`
		Expire time.Time `json:"expire"`
	} `json:"link"`
	NeedMoreQuota  bool          `json:"need_more_quota"`
	VipTypes       []interface{} `json:"vip_types"`
	RedirectLink   string        `json:"redirect_link"`
	IconLink       string        `json:"icon_link"`
	IsDefault      bool          `json:"is_default"`
	Priority       int           `json:"priority"`
	IsOrigin       bool          `json:"is_origin"`
	ResolutionName string        `json:"resolution_name"`
	IsVisible      bool          `json:"is_visible"`
	Category       string        `json:"category"`
}

type UploadTaskData struct {
	UploadType string `json:"upload_type"`
	//UPLOAD_TYPE_RESUMABLE
	Resumable *struct {
		Kind   string `json:"kind"`
		Params struct {
			AccessKeyID     string    `json:"access_key_id"`
			AccessKeySecret string    `json:"access_key_secret"`
			Bucket          string    `json:"bucket"`
			Endpoint        string    `json:"endpoint"`
			Expiration      time.Time `json:"expiration"`
			Key             string    `json:"key"`
			SecurityToken   string    `json:"security_token"`
		} `json:"params"`
		Provider string `json:"provider"`
	} `json:"resumable"`

	File File `json:"file"`
}

type CloudDownloadResp struct {
	UploadType string `json:"upload_type"`
	URL        struct {
		Kind string `json:"kind"`
	} `json:"url"`
	File File `json:"file"`
	Task struct {
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		UserID     string `json:"user_id"`
		Statuses   []any  `json:"statuses"`
		StatusSize int    `json:"status_size"`
		Params     struct {
			PredictSpeed  string `json:"predict_speed"`
			PredictType   string `json:"predict_type"`
			ThumbnailLink string `json:"thumbnail_link"`
		} `json:"params"`
		FileID            string    `json:"file_id"`
		FileName          string    `json:"file_name"`
		FileSize          string    `json:"file_size"`
		Message           string    `json:"message"`
		CreatedTime       time.Time `json:"created_time"`
		UpdatedTime       time.Time `json:"updated_time"`
		ThirdTaskID       string    `json:"third_task_id"`
		Phase             string    `json:"phase"`
		Progress          int       `json:"progress"`
		IconLink          string    `json:"icon_link"`
		Callback          string    `json:"callback"`
		ReferenceResource any       `json:"reference_resource"`
		Space             string    `json:"space"`
	} `json:"task"`
}
