package models

type PolicyGroup struct {
	BaseModel

	Name        string `json:"name" gorm:"not null;size:128;comment:策略组名称" example:"安全合规策略组"`
	Description string `json:"description" gorm:"type:text;comment:描述" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool   `json:"enabled" gorm:"default:true;comment:是否全局启用" example:"true"`
}

func (PolicyGroup) TableName() string {
	return "iac_policy"
}

func (g *PolicyGroup) GetId() Id {
	if g.Id == "" {
		return NewId("pog")
	}
	return g.Id
}
