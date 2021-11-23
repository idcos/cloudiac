package vcsrv

import (
	"cloudiac/portal/models"
	"fmt"
	"testing"
)

func Test_giteaRepoIface_AddWebhook(t *testing.T) {
	url:="http://10.0.2.135:7777/xiaohei"
	gitea := &giteaRepoIface{
		vcs:          &models.Vcs{
			Address:   "http://10.0.3.124:3000",
			VcsToken:  "27b9b370eb32bcbe32200ac37497c0f3a5de10a3",
		},
		repository:   &Repository{
			FullName:      "gitea/tf-vpc",
		},
	}
	err := gitea.AddWebhook(url)
	fmt.Println(err)


}

func Test_giteaRepoIface_ListBranches(t *testing.T) {
	gitea := &giteaRepoIface{
		vcs:          &models.Vcs{
			Address:   "http://10.0.3.124:3000",
			VcsToken:  "27b9b370eb32bcbe32200ac37497c0f3a5de10a3",
		},
		repository:   &Repository{
			FullName:      "gitea/cloudiac-example",
		},
	}
	got, err := gitea.ListBranches()
	fmt.Println(err)
	fmt.Println(got)
}