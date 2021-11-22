package vcsrv

import (
	"cloudiac/portal/models"
	"fmt"
	"net/url"
	"testing"
)

func Test_giteeRepoIface_AddWebhook(t *testing.T) {
	param := url.Values{}
	param.Add("access_token", "a4f073690065fdad7a94d598ed4cc15b")
	gitee := &giteeRepoIface{
		vcs:        &models.Vcs{
			Address:   "https://gitee.com/api/v5",
			VcsToken:  "a4f073690065fdad7a94d598ed4cc15b",
		},
		repository: &RepositoryGitee{
			FullName:      "duyewei/terraform-aliyun-ecs",
		},
		urlParam:   param,
	}

	err:=gitee.AddWebhook("http://baidu.com:7777/xiaohei")
	fmt.Println(err)
}
