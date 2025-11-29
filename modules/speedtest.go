package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	speedtestPingIdx     = 5
	speedtestDownloadIdx = 4
	speedtestUploadIdx   = 3
	speedtestImageIdx    = 2
)

func runSpeedTest() (float64, float64, int, string, error) {
	out, err := utils.RunCommand("timeout 90 speedtest-cli --csv --share")
	if err != nil {
		if strings.Contains(out, "timeout") || strings.Contains(err.Error(), "timeout") {
			return 0, 0, 0, "", fmt.Errorf("speedtest timed out")
		}
		return 0, 0, 0, "", fmt.Errorf("speedtest failed: %w: %s", err, out)
	}

	result := strings.Split(strings.TrimSpace(out), ",")
	if len(result) < 9 {
		return 0, 0, 0, "", fmt.Errorf("unexpected speedtest output format")
	}

	ping, _ := strconv.ParseFloat(strings.TrimSpace(result[len(result)-speedtestPingIdx]), 64)
	download, _ := strconv.ParseFloat(strings.TrimSpace(result[len(result)-speedtestDownloadIdx]), 64)
	upload, _ := strconv.ParseFloat(strings.TrimSpace(result[len(result)-speedtestUploadIdx]), 64)
	imageURL := strings.TrimSpace(result[len(result)-speedtestImageIdx])

	return download / 1e6, upload / 1e6, int(ping), imageURL, nil
}

func speedTestHandler(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("speedtest.running"))

	download, upload, ping, imageURL, err := runSpeedTest()
	if err != nil {
		_, err := msg.Edit(locales.Tr("speedtest.failed"))
		return err
	}

	resultText := fmt.Sprintf(locales.Tr("speedtest.result"), download, upload, ping)

	args := strings.ToLower(strings.TrimSpace(m.Args()))
	if args == "image" && imageURL != "" {
		speedImg := fmt.Sprintf("%s?%d", imageURL, time.Now().Unix())
		_, err = msg.Edit(resultText, &telegram.SendOptions{
			ParseMode: "HTML",
			Media:     speedImg,
		})
	} else {
		_, err = msg.Edit(resultText)
	}

	return err
}

func LoadSpeedTestModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "speedtest", Func: speedTestHandler, Description: "Run a network speedtest", ModuleName: "SpeedTest", DisAllowSudos: true},
		{Command: "st", Func: speedTestHandler, Description: "Run a network speedtest (alias)", ModuleName: "SpeedTest", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)
}
