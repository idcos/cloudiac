package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"github.com/xanzy/go-gitlab"
)

type OpenTemplate struct {
	models.Template
	CommitId string `json:"commit_id,omitempty"`
}

func TemplateDetail(c *ctx.GinRequestCtx) {
	form := forms.OpenApiDetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(OpenDetailTemplate(c.ServiceCtx().DB(), form.Guid))
}

func RunnerListSearch(c *ctx.GinRequestCtx) {
	c.JSONOpenResultList(apps.RunnerListSearch())
}

func OpenDetailTemplate(tx *db.Session, gUid string) (interface{}, e.Error) {
	tpl := OpenTemplate{}
	if err := tx.Table(OpenTemplate{}.TableName()).Where("guid = ?", gUid).First(&tpl); err != nil {
		return nil, e.New(e.DBError, err)
	}
	git, err := services.GetGitConn(tx, tpl.OrgId)
	if err != nil {
		return nil, err
	}
	commits,_,_:=git.Commits.ListCommits(tpl.RepoId,&gitlab.ListCommitsOptions{})
	tpl.CommitId = commits[0].ID
	return tpl, nil
}