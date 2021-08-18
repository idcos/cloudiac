package models

import "cloudiac/portal/libs/db"

type PolicyGroup struct {
	TimedModel

	OrgId Id `json:"orgId" gorm:"not null;size:32;comment:组织ID" example:"组织ID"`

	Name        string `json:"name" gorm:"not null;size:128;comment:策略组名称" example:"安全合规策略组"`
	Description string `json:"description" gorm:"type:text;comment:描述" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool   `json:"enabled" gorm:"default:true;comment:是否启用" example:"true"`
}

func (PolicyGroup) TableName() string {
	return "iac_policy_group"
}

func (g PolicyGroup) GetId() Id {
	if g.Id == "" {
		return NewId("pog")
	}
	return g.Id
}

func (g PolicyGroup) Migrate(sess *db.Session) error {
	if err := g.AddUniqueIndex(sess, "unique__org__name", "org_id", "name"); err != nil {
		return err
	}
	return nil
}
