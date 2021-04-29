package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/services"
	"github.com/xanzy/go-gitlab"
)

func OpenSearchTemplate(c *ctx.ServiceCtx) (interface{}, e.Error) {
	resp := make([]struct {
		Name string `json:"name"`
		Guid string `json:"guid"`
	}, 0)
	if err := services.OpenSearchTemplate(c.DB()).Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return resp, nil
}

type OpenTemplate struct {
	models.Template
	CommitId string `json:"commit_id,omitempty"`
}

func OpenDetailTemplate(c *ctx.ServiceCtx, gUid string) (interface{}, e.Error) {
	tx := c.DB()
	tpl := OpenTemplate{}
	if err := tx.Table(OpenTemplate{}.TableName()).Where("guid = ?", gUid).First(&tpl); err != nil {
		return nil, e.New(e.DBError, err)
	}
	git, err := services.GetGitConn(tx, tpl.OrgId)
	if err != nil {
		return nil, err
	}
	commits, _, commitErr := git.Commits.ListCommits(tpl.RepoId, &gitlab.ListCommitsOptions{})
	if commitErr != nil {
		return nil, e.New(e.GitLabError, commitErr)
	}
	if commits != nil {
		tpl.CommitId = commits[0].ID
	}
	return tpl, nil
}
