package handles

import (
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type SetOpenAiUrlReq struct {
	OpenAiUrl              string `json:"open_ai_url" form:"open_ai_url"`
	OpenAiApiKey           string `json:"open_ai_api_key" form:"open_ai_api_key"`
	OpenAiTranslatePromote string `json:"open_ai_translate_promote" form:"open_ai_translate_promote"`
	OpenAiTranslateModel   string `json:"open_ai_translate_model" form:"open_ai_translate_model"`
}

func SetOpenAi(c *gin.Context) {
	var req SetOpenAiUrlReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	items := []model.SettingItem{
		{Key: conf.OpenAiUrl, Value: req.OpenAiUrl, Type: conf.TypeString, Group: model.OpenAi, Flag: model.PRIVATE},
		{Key: conf.OpenAiApiKey, Value: req.OpenAiApiKey, Type: conf.TypeString, Group: model.OpenAi, Flag: model.PRIVATE},
		{Key: conf.OpenAiTranslatePromote, Value: req.OpenAiTranslatePromote, Type: conf.TypeString, Group: model.OpenAi, Flag: model.PRIVATE},
		{Key: conf.OpenAiTranslateModel, Value: req.OpenAiTranslateModel, Type: conf.TypeString, Group: model.OpenAi, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, "ok")
}
