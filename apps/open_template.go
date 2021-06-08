package apps

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	vcs2 "cloudiac/services/vcsrv"
	"encoding/json"
	"fmt"
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

func ParseVars(vars models.JSON) models.JSON {
	varsList := make([]forms.Var, 0)
	_ = json.Unmarshal(vars, &varsList)
	resultVars := make([]forms.Var, 0)
	for _, v := range varsList {
		if *v.IsSecret {
			v.Value = ""
		}
		if v.Type == consts.Terraform {
			v.Key = fmt.Sprintf("%s%s", consts.TerraformVar, v.Key)
		}
		resultVars = append(resultVars, v)
	}
	b, _ := json.Marshal(resultVars)
	return models.JSON(b)
}

func OpenDetailTemplate(c *ctx.ServiceCtx, gUid string) (interface{}, e.Error) {
	tx := c.DB()
	tpl := OpenTemplate{}
	if err := tx.Table(OpenTemplate{}.TableName()).Where("guid = ?", gUid).First(&tpl); err != nil {
		return nil, e.New(e.DBError, err)
	}
	tpl.Vars = ParseVars(tpl.Vars)

	vcs, er := services.QueryVcsByVcsId(tpl.VcsId, tx)
	if er != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs detail error: %v", er))
	}
	if vcs.VcsType == consts.GitTypeGitLab {
		git, err := vcs2.GetGitConn(vcs.VcsToken, vcs.Address)
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
	}

	if vcs.VcsType == consts.GitTypeGitEA {
		commit,err:= vcs2.GetGiteaBranchCommitId(vcs,uint(tpl.RepoId),tpl.RepoBranch)
		if err != nil {
			return nil, e.New(e.GitLabError, fmt.Errorf("query commit id error: %v", er))
		}
		tpl.CommitId = commit
	}


	return tpl, nil
}
