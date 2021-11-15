package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

const (
	TplExportVersion = "0.1.0"
)

type TplExportedData struct {
	Version   string             `json:"version"`
	Templates []exportedTpl      `json:"templates"`
	Vcs       []exportedVcs      `json:"vcs"`
	VarGroups []exportedVarGroup `json:"varGroups"`
}

type exportedTpl struct {
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

	Variables   []exportedTplVar `json:"variables"`
	VarGroupIds []models.Id      `json:"varGroupIds"`
}

type exportedTplVar struct {
	Id          string   `json:"id"`
	Scope       string   `json:"scope"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Value       string   `json:"value"`
	Options     []string `json:"options"`
	Sensitive   bool     `json:"sensitive"`
	Description string   `json:"description"`
}

type exportedVcs struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	VcsType  string `json:"vcsType"`
	Address  string `json:"address"`
	VcsToken string `json:"vcsToken"`
}

type exportedVarGroup struct {
	Id        string                    `json:"id"`
	Name      string                    `json:"name"`
	Type      string                    `json:"type"`
	Variables []models.VarGroupVariable `json:"variables"`
}

func ExportTemplates(dbSess *db.Session, orgId models.Id, ids []models.Id) (*TplExportedData, e.Error) {
	tpls := make([]models.Template, 0)
	orgDb := QueryWithOrgId(dbSess, orgId, models.Template{}.TableName())

	if err := QueryTemplate(orgDb.Where("id IN (?)", ids)).Find(&tpls); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	tplIds := make([]models.Id, 0)
	vcsIdSet := make(map[models.Id]struct{})
	for _, t := range tpls {
		vcsIdSet[t.VcsId] = struct{}{}
		tplIds = append(tplIds, t.Id)
	}

	vars := make([]models.Variable, 0)
	if err := QueryVariable(dbSess.Where(
		"scope = ? AND tpl_id IN (?)", consts.ScopeTemplate, tplIds)).Find(&vars); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	vgs, err := FindTplsRelVarGroup(dbSess, tplIds)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	vcsIds := make([]models.Id, 0)
	for k := range vcsIdSet {
		vcsIds = append(vcsIds, k)
	}
	vcsList := make([]models.Vcs, 0)
	if err := QueryVcsSample(dbSess.Where("id IN (?)", vcsIds)).Find(&vcsList); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	resp := TplExportedData{
		Version:   TplExportVersion,
		Templates: make([]exportedTpl, 0),
		Vcs:       make([]exportedVcs, 0),
		VarGroups: make([]exportedVarGroup, 0),
	}

	for _, t := range tpls {
		tpl := exportedTpl{
			Id:           t.Id.String(),
			Name:         t.Name,
			TplType:      t.TplType,
			Description:  t.Description,
			VcsId:        t.VcsId.String(),
			RepoId:       t.RepoId,
			RepoAddr:     t.RepoAddr,
			RepoToken:    ExportSecretStr(t.RepoToken, true),
			RepoRevision: t.RepoRevision,
			Status:       t.Status,
			Workdir:      t.Workdir,
			TfVarsFile:   t.TfVarsFile,
			Playbook:     t.Playbook,
			PlayVarsFile: t.PlayVarsFile,
			TfVersion:    t.TfVersion,
			Variables:    []exportedTplVar{},
		}

		for _, v := range vars {
			if v.TplId != t.Id {
				continue
			}
			tpl.Variables = append(tpl.Variables, exportedTplVar{
				Id:          v.Id.String(),
				Scope:       v.Scope,
				Type:        v.Type,
				Name:        v.Name,
				Value:       ExportVariableValue(v.Value, v.Sensitive),
				Options:     v.Options,
				Sensitive:   v.Sensitive,
				Description: v.Description,
			})
		}

		vgIds, er := FindTemplateVgIds(dbSess, t.Id)
		if er != nil {
			return nil, er
		}
		tpl.VarGroupIds = vgIds

		resp.Templates = append(resp.Templates, tpl)
	}

	for _, vcs := range vcsList {
		resp.Vcs = append(resp.Vcs, exportedVcs{
			Id:       vcs.Id.String(),
			Name:     vcs.Name,
			Status:   vcs.Status,
			VcsType:  vcs.VcsType,
			Address:  vcs.Address,
			VcsToken: ExportSecretStr(vcs.VcsToken, true),
		})
	}

	for _, vg := range vgs {
		evg := exportedVarGroup{
			Id:        vg.Id.String(),
			Name:      vg.Name,
			Type:      vg.Type,
			Variables: vg.Variables,
		}
		for i, v := range evg.Variables {
			v.Value = ExportVariableValue(v.Value, v.Sensitive)
			evg.Variables[i] = v
		}
		resp.VarGroups = append(resp.VarGroups, evg)
	}

	return &resp, nil
}
