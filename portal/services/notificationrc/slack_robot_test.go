// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package notificationrc

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	webhookUrl := "https://hooks.slack.com/services/xxxx"
	err := SendSlack(webhookUrl,
		Payload{Text: "``` xiaohei```", Markdown: true})
	fmt.Println(err)
}
