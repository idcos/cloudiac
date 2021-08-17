package notificationrc

import (
	"fmt"
	"testing"
)

func TestRobot(t *testing.T) {
	ding := NewDingTalkRobot("https://oapi.dingtalk.com/robot/send?access_token=2cc268f8ed58ba4c357b7e8e64c833197ba1211c6311493dc34323f013a4bb45", "SEC46f142ce9cf820f39cc79d826a9eec89f4d1010c540974844c3ec93fbb448725")
	//fmt.Println(ding.SendTextMessage("contnet", []string{"13624015331"}, false))
	fmt.Println(ding.SendMarkdownMessage("test", "## 任务发起通知\n尊敬的 xxx：\n\n\t【xxx】在CloudIaC平台发起了部署任务，详情如下：\n\t\n\t所属组织：zzz\n\t所属项目：project-1\n\t云模板：template-1\n\t分支/tag：release/0.5.0\n\t环境名称：测试环境\n\t\n\t更多详情请点击：http://\n\t\n\t-----该邮件由系统自动发出，请勿回复-----\n\t", []string{"13624015331"}, false))
}
