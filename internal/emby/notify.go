package emby

import (
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"strings"
)

var client *resty.Client

func Refresh(path string) {

	split := strings.Split(path, "\n")
	urlMap := make(map[string]bool)
	for _, url := range split {
		if url != "" {
			urlMap[url] = true
		}
	}

	for url := range urlMap {
		go func() {
			resp, err := client.R().Post(url)

			if err != nil {
				respStr := ""
				if resp != nil {
					respStr = resp.String()
				}
				utils.Log.Warnf("failed to refresh emby infos, error message: %s, response: %s", err.Error(), respStr)
			}
		}()
	}

}

func init() {
	client = resty.New().
		SetRetryCount(3).
		SetRetryResetReaders(true)
}
