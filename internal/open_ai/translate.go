package open_ai

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"strings"
)

func Translate(text string) string {

	openAiUrl := setting.GetStr(conf.OpenAiUrl)
	openAiApiKey := setting.GetStr(conf.OpenAiApiKey)
	translatePromote := setting.GetStr(conf.OpenAiTranslatePromote)
	translateModel := setting.GetStr(conf.OpenAiTranslateModel)

	if openAiUrl == "" || openAiApiKey == "" {
		return text
	}

	execTranslateFunc := func(model, text string) string {

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		utils.Log.Debugf("开始翻译:%s", text)
		response, err := base.RestyClient.R().SetAuthToken(openAiApiKey).SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		}).SetBody(base.Json{
			"messages": func() []map[string]any {

				var param []map[string]any

				if translatePromote != "" {
					param = append(param, map[string]any{
						"role":    "system",
						"content": translatePromote,
					})
				}

				param = append(param, map[string]any{
					"role":    "user",
					"content": text,
				})

				return param
			}(),
			"model":             model,
			"temperature":       0.5,
			"presence_penalty":  0,
			"frequency_penalty": 0,
			"top_p":             1,
		}).SetResult(&result).Post(fmt.Sprintf("%s/v1/chat/completions", openAiUrl))

		if err != nil {
			var detail string
			if response != nil {
				detail = string(response.Body())
			}
			utils.Log.Warnf("翻译失败:%s,响应信息为:%s", err.Error(), detail)
			return ""
		}

		if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
			utils.Log.Warnf("翻译结果为空,响应信息为:%s", response.String())
			return ""
		}

		return result.Choices[0].Message.Content
	}

	for _, model := range strings.Split(translateModel, ",") {
		ans := execTranslateFunc(model, text)
		if ans != "" {
			return ans
		}
	}

	return text

}
