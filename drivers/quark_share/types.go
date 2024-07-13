package quark_share

import "github.com/alist-org/alist/v3/internal/model"

type Resp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ShareTokenResp struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Data      struct {
		Subscribed bool   `json:"subscribed"`
		Stoken     string `json:"stoken"`
		ShareType  int    `json:"share_type"`
		Author     struct {
			MemberType string `json:"member_type"`
			AvatarURL  string `json:"avatar_url"`
			NickName   string `json:"nick_name"`
		} `json:"author"`
		ExpiredType int    `json:"expired_type"`
		ExpiredAt   int64  `json:"expired_at"`
		Title       string `json:"title"`
		FileNum     int    `json:"file_num"`
	} `json:"data"`
}

type ShareTokenReq struct {
	PwdId    string `json:"pwd_id"`
	PassCode string `json:"passcode"`
}

type FileListResp struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Data      struct {
		IsOwner int `json:"is_owner"`
		List    []struct {
			Fid           string `json:"fid"`
			FileName      string `json:"file_name"`
			PdirFid       string `json:"pdir_fid"`
			Size          int64  `json:"size"`
			OperatedAt    int64  `json:"operated_at"`
			Thumbnail     string `json:"thumbnail"`
			UpdateViewAt  int64  `json:"update_view_at"`
			LastUpdateAt  int64  `json:"last_update_at"`
			ShareFidToken string `json:"share_fid_token"`
			Dir           bool   `json:"dir"`
			CreatedAt     int64  `json:"created_at"`
			UpdatedAt     int64  `json:"updated_at"`
		} `json:"list"`
	} `json:"data"`
	Metadata struct {
		Size          int `json:"_size"`
		Page          int `json:"_page"`
		Count         int `json:"_count"`
		Total         int `json:"_total"`
		CheckFidToken int `json:"check_fid_token"`
	} `json:"metadata"`
}
type FileObj struct {
	model.ObjThumb
	ShareFidToken string
}

type TransformResult struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Data      struct {
		TaskID     string `json:"task_id"`
		EventID    string `json:"event_id"`
		TaskType   int    `json:"task_type"`
		TaskTitle  string `json:"task_title"`
		Status     int    `json:"status"`
		CreatedAt  int64  `json:"created_at"`
		FinishedAt int64  `json:"finished_at"`
		Share      struct {
		} `json:"share"`
		SaveAs struct {
			SearchExit    bool     `json:"search_exit"`
			ToPdirFid     string   `json:"to_pdir_fid"`
			IsPack        string   `json:"is_pack"`
			SaveAsTopFids []string `json:"save_as_top_fids"`
			ToPdirName    string   `json:"to_pdir_name"`
		} `json:"save_as"`
	} `json:"data"`
	Metadata struct {
		TqGap int `json:"tq_gap"`
	} `json:"metadata"`
}
