// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils/logs"
	"fmt"
	"strings"

	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

func VscTokenCheckByID(tx *db.Session, id models.Id, withNewToken string) error {
	vcs, err := QueryVcsByVcsId(id, tx)
	if err != nil {
		return err
	}
	if len(withNewToken) != 0 {
		vcs.VcsToken = withNewToken
	}
	if err := vcsrv.VerifyVcsToken(vcs); err != nil {
		return err
	}
	return nil
}

func QueryVcs(orgId models.Id, status, q string, isShowdefaultVcs, isShowRegistryVcs bool, query *db.Session) *db.Session {
	query = query.Model(&models.Vcs{}).Omit("vcs_token").Where("org_id = ? or org_id = ''", orgId)
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
	return query
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

func FindEnableVcs(orgId models.Id, query *db.Session) ([]models.Vcs, e.Error) {
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
	Sensitive   bool   `json:"sensitive" form:"sensitive"`
	Required    bool   `json:"required" form:"required"`
	Value       string `json:"value" form:"value"`
	Name        string `json:"name" form:"name"`
}

type tfVariableConfig struct {
	Upstreams []*tfVariableBlock `hcl:"variable,block"`
}

type tfVariableBlock struct {
	Name        string      `hcl:",label"`
	Default     interface{} `hcl:"default,optional"`
	Type        interface{} `hcl:"type,optional"`
	Description string      `hcl:"description,optional"`
	Sensitive   bool        `hcl:"sensitive,optional"`
	Validation  *struct {
		Condition    interface{} `hcl:"condition,attr"`
		ErrorMessage string      `hcl:"error_message,optional"`
	} `hcl:"validation,block"`
}

// ParseTfVariables hcl parse doc: https://pkg.go.dev/github.com/hashicorp/hcl/v2/gohcl
func ParseTfVariables(filename string, content []byte) ([]TemplateVariable, e.Error) {
	logger := logs.Get().WithField("filename", filename)
	file, diagErrs := hclsyntax.ParseConfig(content, filename, hcl.Pos{Line: 1, Column: 1})
	if diagErrs != nil && diagErrs.HasErrors() {
		logger.Error(fmt.Errorf("ParseConfig: %w", diagErrs))
		return nil, e.New(e.HCLParseError, diagErrs)
	}

	c := &tfVariableConfig{}
	diagErrs = gohcl.DecodeBody(file.Body, nil, c)
	for _, d := range diagErrs {
		logger.Warnf(d.Error())
	}
	tv := make([]TemplateVariable, 0)
	for _, s := range c.Upstreams {
		v, ok := s.Default.(*hcl.Attribute)
		if ok {
			// 有默认值，则通过描述中的关键字判断是否必填
			required := strings.Contains(s.Description, "（必填）") || strings.Contains(s.Description, "(必填)")
			val, _ := v.Expr.Value(nil)
			if val.IsWhollyKnown() {
				valJSON, err := ctyjson.Marshal(val, val.Type())
				if err != nil {
					return nil, e.New(e.HCLParseError, fmt.Errorf("failed to serialize default value as JSON: %s", err))
				}
				tv = append(tv, TemplateVariable{
					Value:       strings.Trim(string(valJSON), "\""),
					Name:        s.Name,
					Sensitive:   s.Sensitive,
					Required:    required,
					Description: s.Description,
				})
			}
		} else {
			tv = append(tv, TemplateVariable{
				Name:        s.Name,
				Required:    true, // 无默认值，则变量为必填
				Description: s.Description,
			})
		}
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
