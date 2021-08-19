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
