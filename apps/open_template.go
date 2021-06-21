package apps

import (
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/services/vcsrv"
	"encoding/json"
	"fmt"
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

	vcsService, vcsErr := vcsrv.GetVcsInstance(vcs)
	if vcsErr != nil {
		return nil, e.New(e.GitLabError, vcsErr)
	}

	repo, vcsErr := vcsService.GetRepo(tpl.RepoId)
	if vcsErr != nil {
		return nil, e.New(e.GitLabError, vcsErr)
	}

	commitId, vcsErr := repo.BranchCommitId(tpl.RepoBranch)
	if vcsErr != nil {
		return nil, e.New(e.GitLabError, vcsErr)
	}

	tpl.CommitId = commitId

	return tpl, nil
}
