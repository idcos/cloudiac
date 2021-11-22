package vcsrv

import (
	"cloudiac/portal/models"
	"fmt"
	"testing"
)

func Test_githubRepoIface_AddWebhook(t *testing.T) {
	github := &githubRepoIface{
		vcs:        &models.Vcs{
			Address:   "https://api.github.com/",
			VcsToken:  "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",
		},
		repository: &RepositoryGithub{
			FullName:      "xiaohei417/test",
		},
	}
	err := github.AddWebhook("http://10.0.2.135:7777/xiaohei")
	fmt.Println(err)
}
