package models

type PolicyRel struct {
	BaseModel

	GroupId   Id `json:"group_id" gorm:"size:32;comment:策略组ID" example:"lg-c3lcrjxczjdywmk0go90"`
	OrgId     Id `json:"org_id" gorm:"size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"project_id" gorm:"size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`
	TplId     Id `json:"tpl_id" gorm:"size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`
	EnvId     Id `json:"env_id" gorm:"size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`
}

func (PolicyRel) TableName() string {
	return "iac_policy_rel"
}

func (g *PolicyRel) GetId() Id {
	if g.Id == "" {
		return NewId("pol")
	}
	return g.Id
}
