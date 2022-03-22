// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/vcsrv"
	"encoding/json"
	"errors"
	"fmt"

	legacyhcl "github.com/hashicorp/hcl"
	legacyast "github.com/hashicorp/hcl/hcl/ast"
	"gorm.io/gorm"
)

func CreateVcs(tx *db.Session, vcs models.Vcs) (*models.Vcs, e.Error) {
	if vcs.Id == "" {
		vcs.Id = vcs.NewId()
	}
	if err := models.Create(tx, &vcs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &vcs, nil
}

func UpdateVcs(tx *db.Session, id models.Id, attrs models.Attrs) (vcs *models.Vcs, er e.Error) {
	vcs = &models.Vcs{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Vcs{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update vcs error: %v", err))
	} //nolint
	if err := tx.Where("id = ?", id).First(vcs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs error: %v", err))
	}
	return
}

func QueryVcs(orgId models.Id, status, q string, isShowdefaultVcs, isShowRegistryVcs bool, query *db.Session) *db.Session {
	query = query.Model(&models.Vcs{}).Where("org_id = ? or org_id = ''", orgId)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("name LIKE ?", qs)
	}
	if !isShowdefaultVcs {
		query = query.Where("vcs_type != ?", consts.GitTypeLocal)
	}
	if !isShowRegistryVcs {
		query = query.Where("vcs_type != ?", consts.GitTypeRegistry)
	}
	return query.LazySelectAppend("id, org_id, project_id, name, status, vcs_type, address")
}

func QueryVcsSample(query *db.Session) *db.Session {
	return query.Model(&models.Vcs{})
}

func QueryVcsByVcsId(vcsId models.Id, query *db.Session) (*models.Vcs, e.Error) {
	vcs := &models.Vcs{}
	if vcsId == "" {
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

func GetVcsById(sess *db.Session, id models.Id) (*models.Vcs, e.Error) {
	vcs := models.Vcs{}
	err := sess.Model(&models.Vcs{}).Where("id = ?", id).First(&vcs)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.VcsNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &vcs, nil
}

func GetVcsListByIds(sess *db.Session, ids []string) ([]models.Vcs, e.Error) {
	vcs := make([]models.Vcs, 0)
	err := sess.Model(&models.Vcs{}).Where("id in (?)", ids).Find(&vcs)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return vcs, nil
}

func GetVcsRepoByTplId(sess *db.Session, tplId models.Id) (vcsrv.RepoIface, e.Error) {
	tpl, err := GetTemplateById(sess, tplId)
	if err != nil {
		return nil, err
	}

	vcs, err := GetVcsById(sess, tpl.VcsId)
	if err != nil {
		return nil, err
	}

	if repo, err := vcsrv.GetRepo(vcs, tpl.RepoId); err != nil {
		return nil, e.AutoNew(err, e.VcsError)
	} else {
		return repo, nil
	}
}

func QueryEnableVcs(orgId models.Id, query *db.Session) (interface{}, e.Error) {
	vcs := make([]models.Vcs, 0)
	if err := query.Model(&models.Vcs{}).Where("org_id = ? or org_id = 0", orgId).Where("status = 'enable'").Find(&vcs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return vcs, nil
}

func DeleteVcs(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Vcs{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete vcs error: %v", err))
	}
	return nil
}

type TemplateVariable struct {
	Description string `json:"description" form:"description" `
	Value       string `json:"value" form:"value" `
	Name        string `json:"name" form:"name" `
}

type tfVariableBlock struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`

	Default   interface{} `json:"default"`
	Required  bool        `json:"required"`
	Sensitive bool        `json:"sensitive,omitempty"`
}

// ParseTfVariables hcl parse doc: https://pkg.go.dev/github.com/hashicorp/hcl/v2/gohcl
func ParseTfVariables(filename string, content []byte) ([]TemplateVariable, e.Error) {
	tfVariable, err := parseVariables(content)
	if err != nil {
		return nil, e.New(e.HCLParseError, err)
	}

	tv := make([]TemplateVariable, 0)
	for _, s := range tfVariable {
		b, err := json.Marshal(s.Default)
		if err != nil {
			return nil, e.New(e.JSONParseError)
		}
		tv = append(tv, TemplateVariable{
			Value:       string(b),
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return tv, nil
}

func GetDefaultVcs(session *db.Session) (*models.Vcs, error) {
	vcs := &models.Vcs{}
	err := session.Where("org_id = '' AND name = ?", consts.DefaultVcsName).Find(vcs)
	if vcs.Id == "" {
		return vcs, gorm.ErrRecordNotFound
	}
	return vcs, err
}

func GetRegistryVcs(session *db.Session) (*models.Vcs, error) {
	vcs := &models.Vcs{}
	err := session.Where("org_id = '' AND name = ?", consts.RegistryVcsName).Find(vcs)
	if vcs.Id == "" {
		return vcs, gorm.ErrRecordNotFound
	}
	return vcs, err
}

func CreateVcsPr(session *db.Session, vcsPr models.VcsPr) e.Error {
	if err := models.Create(session, &vcsPr); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func GetVcsPrByTaskId(session *db.Session, task *models.Task) (models.VcsPr, error) {
	vp := models.VcsPr{}
	if err := session.Model(&models.VcsPr{}).
		Where("env_id = ?", task.EnvId).
		Where("task_id = ?", task.Id).First(&vp); err != nil {
		return vp, err
	}
	return vp, nil
}

func parseVariables(src []byte) ([]tfVariableBlock, error) {
	tv := make([]tfVariableBlock, 0)
	hclRoot, err := legacyhcl.Parse(string(src))
	if err != nil {
		return tv, errors.New(fmt.Sprintf("Error parsing: %s", err))
	}

	list, ok := hclRoot.Node.(*legacyast.ObjectList)
	if !ok {
		return tv, errors.New("error parsing no root object")
	}

	if vars := list.Filter("variable"); len(vars.Items) > 0 {
		vars = vars.Children()
		type VariableBlock struct {
			Type        string `hcl:"type"`
			Default     interface{}
			Description string
			Fields      []string `hcl:",decodedFields"`
		}

		for _, item := range vars.Items {
			unwrapLegacyHCLObjectKeysFromJSON(item, 1)

			if len(item.Keys) != 1 {
				return nil, errors.New(fmt.Sprintf("variable block at %s has no label", item.Pos()))
			}

			name := item.Keys[0].Token.Value().(string)

			var block VariableBlock
			err := legacyhcl.DecodeObject(&block, item.Val)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("invalid variable block at %s: %s", item.Pos(), err))
			}

			if ms, ok := block.Default.([]map[string]interface{}); ok {
				def := make(map[string]interface{})
				for _, m := range ms {
					for k, v := range m {
						def[k] = v
					}
				}
				block.Default = def
			}

			tv = append(tv, tfVariableBlock{
				Name:        name,
				Type:        block.Type,
				Description: block.Description,
				Default:     block.Default,
				Required:    block.Default == nil,
			})

		}
	}

	return tv,nil
}

func unwrapLegacyHCLObjectKeysFromJSON(item *legacyast.ObjectItem, depth int) {
	if len(item.Keys) > depth && item.Keys[0].Token.JSON {
		for len(item.Keys) > depth {
			// Pop off the last key
			n := len(item.Keys)
			key := item.Keys[n-1]
			item.Keys[n-1] = nil
			item.Keys = item.Keys[:n-1]

			// Wrap our value in a list
			item.Val = &legacyast.ObjectType{
				List: &legacyast.ObjectList{
					Items: []*legacyast.ObjectItem{
						{
							Keys: []*legacyast.ObjectKey{key},
							Val:  item.Val,
						},
					},
				},
			}
		}
	}
}
