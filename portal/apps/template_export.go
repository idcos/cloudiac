package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

const (
	TplExportVersion = "0.1.0"
)

type TplExportResp struct {
	Version   string              `json:"version"`
	Templates []TplExportTpl      `json:"templates"`
	Vcs       []TplExportVcs      `json:"vcs"`
	VarGroups []TplExportVarGroup `json:"varGroups"`
}

type TplExportTpl struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	TplType      string `json:"tplType"`
	Description  string `json:"description"`
	VcsId        string `json:"vcsId"`
	RepoId       string `json:"repoId"`
	RepoAddr     string `json:"repoAddr"`
	RepoToken    string `json:"repoToken"`
	RepoRevision string `json:"repoRevision"`
	Status       string `json:"status"`
	Workdir      string `json:"workdir"`
	TfVarsFile   string `json:"tfVarsFile"`
	Playbook     string `json:"playbook"`
	PlayVarsFile string `json:"playVarsFile"`
	TfVersion    string `json:"tfVersion"`

	Variables []TplExportTplVar `json:"variables"`
}

type TplExportTplVar struct {
	Id          string   `json:"id"`
	Scope       string   `json:"scope"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Value       string   `json:"value"`
	Options     []string `json:"options"`
	Sensitive   bool     `json:"sensitive"`
	Description string   `json:"description"`
}

type TplExportVcs struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	VcsType  string `json:"vcsType"`
	Address  string `json:"address"`
	VcsToken string `json:"vcsToken"`
}

type TplExportVarGroup struct {
	Id        string                    `json:"id"`
	Name      string                    `json:"name"`
	Type      string                    `json:"type"`
	Variables []models.VarGroupVariable `json:"variables"`
}

type TplExportForm struct {
	forms.BaseForm

	IDs      []string `json:"ids" form:"ids" binding:required`
	Download bool     `json:"download" form:"download"`
}

func TemplateExport(c *ctx.ServiceContext, form *TplExportForm) (*TplExportResp, e.Error) {
	tpls := make([]models.Template, 0)
	dbSess := c.DB()
	orgDb := services.QueryWithOrgId(dbSess, c.OrgId, models.Template{}.TableName())

	if err := services.QueryTemplate(orgDb.Where("id IN (?)", form.IDs)).Find(&tpls); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	tplIds := make([]models.Id, 0)
	vcsIdSet := make(map[models.Id]struct{})
	for _, t := range tpls {
		vcsIdSet[t.VcsId] = struct{}{}
		tplIds = append(tplIds, t.Id)
	}

	vars := make([]models.Variable, 0)
	if err := services.QueryVariable(dbSess.Where(
		"scope = ? AND tpl_id IN (?)", consts.ScopeTemplate, tplIds)).Find(&vars); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	vgs, err := services.FindTplsRelVarGroup(dbSess.Debug(), tplIds)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	vcsIds := make([]models.Id, 0)
	for k := range vcsIdSet {
		vcsIds = append(vcsIds, k)
	}
	vcsList := make([]models.Vcs, 0)
	if err := services.QueryVcsSample(dbSess.Where("id IN (?)", vcsIds)).Find(&vcsList); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	resp := TplExportResp{
		Version:   TplExportVersion,
		Templates: make([]TplExportTpl, 0),
		Vcs:       make([]TplExportVcs, 0),
		VarGroups: make([]TplExportVarGroup, 0),
	}

	for _, t := range tpls {
		tpl := TplExportTpl{
			Id:           t.Id.String(),
			Name:         t.Name,
			TplType:      t.TplType,
			Description:  t.Description,
			VcsId:        t.VcsId.String(),
			RepoId:       t.RepoId,
			RepoAddr:     t.RepoAddr,
			RepoToken:    services.ExportSecretStr(t.RepoToken, true),
			RepoRevision: t.RepoRevision,
			Status:       t.Status,
			Workdir:      t.Workdir,
			TfVarsFile:   t.TfVarsFile,
			Playbook:     t.Playbook,
			PlayVarsFile: t.PlayVarsFile,
			TfVersion:    t.TfVersion,
			Variables:    []TplExportTplVar{},
		}

		for _, v := range vars {
			if v.TplId != t.Id {
				continue
			}
			tpl.Variables = append(tpl.Variables, TplExportTplVar{
				Id:          v.Id.String(),
				Scope:       v.Scope,
				Type:        v.Type,
				Name:        v.Name,
				Value:       services.ExportVariableValue(v.Value, v.Sensitive),
				Options:     v.Options,
				Sensitive:   v.Sensitive,
				Description: v.Description,
			})
		}
		resp.Templates = append(resp.Templates, tpl)
	}

	for _, vcs := range vcsList {
		resp.Vcs = append(resp.Vcs, TplExportVcs{
			Id:       vcs.Id.String(),
			Name:     vcs.Name,
			Status:   vcs.Status,
			VcsType:  vcs.VcsType,
			Address:  vcs.Address,
			VcsToken: services.ExportSecretStr(vcs.VcsToken, true),
		})
	}

	for _, vg := range vgs {
		evg := TplExportVarGroup{
			Id:        vg.Id.String(),
			Name:      vg.Name,
			Type:      vg.Type,
			Variables: vg.Variables,
		}
		for i, v := range evg.Variables {
			v.Value = services.ExportVariableValue(v.Value, v.Sensitive)
			evg.Variables[i] = v
		}
		resp.VarGroups = append(resp.VarGroups, evg)
	}

	return &resp, nil
}

func TemplateImport() {

}
