package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func CreateVcs(tx *db.Session, vcs models.Vcs) (*models.Vcs, e.Error) {
	if err := models.Create(tx, &vcs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &vcs, nil
}

func UpdateVcs(tx *db.Session, id uint, attrs models.Attrs) (vcs *models.Vcs, er e.Error) {
	vcs = &models.Vcs{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Vcs{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update vcs error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(vcs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs error: %v", err))
	}
	return
}

func QueryVcs(orgId uint, status, q string, query *db.Session) *db.Session {
	query = query.Model(&models.Vcs{})
	if status != "" {
		query = query.Where("status = ?", status).
			Where("org_id = ? or org_id = 0", orgId)
	} else {
		query = query.
			Where("org_id = ?", orgId)
	}
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("name LIKE ?", qs)
	}
	return query
}

func QueryVcsByVcsId(vcsId uint, query *db.Session) (*models.Vcs, e.Error) {
	vcs := &models.Vcs{}
	if vcsId == 0 {
		query = query.Where("org_id = 0")
	} else {
		query = query.Where("id = ?", vcsId)
	}

	err := query.First(vcs)
	if err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs detail error: %v", err))
	}
	return vcs, nil

}

func QueryEnableVcs(orgId uint, query *db.Session) (interface{}, e.Error) {
	vcs := make([]models.Vcs, 0)
	if err := query.Model(&models.Vcs{}).Where("org_id = ? or org_id = 0", orgId).Where("status = 'enable'").Find(&vcs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return vcs, nil
}

func DeleteVcs(tx *db.Session, id uint) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Vcs{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete vcs error: %v", err))
	}
	return nil
}

type TemplateVariable struct {
	Description string `json:"description" form:"description" `
	Default     string `json:"default" form:"default" `
	Name        string `json:"name" form:"name" `
}

func TemplateVariableSearch(content []byte) ([]TemplateVariable, e.Error) {
	return readHCLFile(content)
}

type Config struct {
	Upstreams []*TfVariable `hcl:"variable,block"`
}

type TfVariable struct {
	Name string `hcl:",label"`
	// Default     string `hcl:"default,optional"`
	Description string `hcl:"description,optional"`
	// validation block
	Default string `hcl:"default,optional"`
}

func readHCLFile(content []byte) ([]TemplateVariable, e.Error) {
	file, diags := hclsyntax.ParseConfig(content, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		logs.Get().Error(fmt.Errorf("ParseConfig: %w", diags))
	}
	c := &Config{}
	diags = gohcl.DecodeBody(file.Body, nil, c)
	if diags.HasErrors() {
		return nil, e.New(e.GitLabError, fmt.Errorf("DecodeBody: %w", diags))
	}
	tv := make([]TemplateVariable, 0)
	for _, s := range c.Upstreams {
		tv = append(tv, TemplateVariable{
			Default:     s.Default,
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return tv, nil
}

func GetDefaultVcs(session *db.Session) (*models.Vcs, error) {
	vcs := &models.Vcs{}
	err := session.Where("org_id = 0").First(vcs)
	return vcs, err

}
