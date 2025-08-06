package spider

import (
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"time"
)

func Visit(spiderServer, url string, maxWaitTime time.Duration, pageFunc func(wd selenium.WebDriver)) error {

	caps := selenium.Capabilities{
		"browserName": "chrome", "browserVersion": "135.0", "se:noVncPort": 7900, "se:vncEnabled": true,
	}

	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--headless",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--start-maximized",
			"--user-data-dir=/tmp/chrome-data",
		},
		W3C: true,
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, spiderServer)
	if err != nil {
		utils.Log.Warnf("Failed to start Selenium: %s", err.Error())
		return err
	}
	defer func(wd selenium.WebDriver) {
		err1 := wd.Quit()
		if err1 != nil {
			utils.Log.Warnf("failed to close remote webdriver: %s", err1.Error())
		}
	}(wd)

	err = wd.Get(url)
	if err != nil {
		return err
	}
	time.Sleep(maxWaitTime)

	pageFunc(wd)

	return nil

}
