package open_ai

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func Translate(text string) string {

	openAiUrl := setting.GetStr(conf.OpenAiUrl)
	openAiApiKey := setting.GetStr(conf.OpenAiApiKey)
	translatePromote := setting.GetStr(conf.OpenAiTranslatePromote)
	translateModel := setting.GetStr(conf.OpenAiTranslateModel)

	if openAiUrl == "" || openAiApiKey == "" {
		return text
	}

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
		"model":             translateModel,
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
		return text
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		utils.Log.Warnf("翻译结果为空,响应信息为:%v", result)
		return text
	}

	return result.Choices[0].Message.Content

}
