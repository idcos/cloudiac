package notificationrc

import (
	"fmt"
	"testing"
)

func TestWeChatRobot(t *testing.T) {
	robotWebhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxx"
	wr := WeChatRobot{Url: robotWebhook}
	_, err := wr.SendMarkdown("```xiaohei_test```")
	fmt.Println(err)
}
