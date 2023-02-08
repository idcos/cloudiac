// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package notificationrc

import (
	"fmt"
	"testing"
)

func TestRobot(t *testing.T) {
	ding := NewDingTalkRobot("https://oapi.dingtalk.com/robot/send?access_token=xxx", "xxx")
	//fmt.Println(ding.SendTextMessage("contnet", []string{"13624015331"}, false))
	fmt.Println(ding.SendMarkdownMessage("test", `
尊敬的 .username：

	【xxx】在CloudIaC平台发起的部署任务执行失败，详情如下：

	所属组织：{{.OrgName}}

	所属项目：{{.ProjectName}}

	云模板：{{.TemplateName}}

	分支/tag：{{.Release}}

	环境名称：{{.EnvName}}

	执行结果：失败

	失败原因：{{.Message}}

	更多详情请点击：{{.Addr}}

	-----该邮件由系统自动发出，请勿回复-----
`, []string{"13624015331"}, false))
}
