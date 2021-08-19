package notificationrc

import (
	"fmt"
	"testing"
)

func TestWeChatRobot(t *testing.T) {
	robotWebhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=fcfa9f55-61df-46c8-992e-4ad7d72c5bba"
	wr := WeChatRobot{Url: robotWebhook}
	_, err := wr.SendMarkdown("```xiaohei_test```")
	fmt.Println(err)
}
