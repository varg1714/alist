package av

import (
	"fmt"
	"github.com/OpenListTeam/OpenList/v4/internal/spider"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/tebeka/selenium"
	"strings"
	"time"
)

func QueryJavSql(spiderServer, sql string, spiderMaxWaitTime int) []string {

	var res []string

	url := fmt.Sprintf("https://jinjier.art/sql?q=%s;", sql)

	err := spider.Visit(spiderServer, url, time.Duration(spiderMaxWaitTime)*time.Second, func(wd selenium.WebDriver) {

		buttonEle, err := wd.FindElement(selenium.ByID, "execute")
		if err != nil {
			utils.Log.Warnf("failed to find button element: %s", err.Error())
			return
		}

		err = buttonEle.Click()
		if err != nil {
			utils.Log.Warnf("failed to click button element: %s", err.Error())
			return
		}

		time.Sleep(time.Duration(spiderMaxWaitTime) * time.Second)

		tableEle, err := wd.FindElement(selenium.ByID, "output")
		if err != nil {
			utils.Log.Warnf("failed to find table element: %s", err.Error())
			return
		}

		aElements, err := tableEle.FindElements(selenium.ByTagName, "a")
		if err != nil {
			utils.Log.Warnf("failed to find a elements: %s", err.Error())
			return
		}

		for _, element := range aElements {
			href, err1 := element.GetAttribute("href")
			if err1 == nil {
				res = append(res, strings.ReplaceAll(href, "/search?q=", ""))
			}
		}

	})

	if err != nil {
		utils.Log.Warnf("failed to get jav sql info, error message: %s", err.Error())
	}

	return res

}
