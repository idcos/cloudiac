package notificationrc

import (
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	err := SendSlack("https://hooks.slack.com/services/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",
		Payload{Text: "``` xiaohei```", Markdown: true})
	fmt.Println(err)
}
